package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: go run ./cmd/generate modules <name>")
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	switch args[0] {
	case "modules":
		if err := generateModule(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "module generation failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		os.Exit(1)
	}
}

func generateModule(rawName string) error {
	name, err := normalizeName(rawName)
	if err != nil {
		return err
	}

	modulePath, err := currentModulePath()
	if err != nil {
		return err
	}

	base := filepath.Join("internal", "modules", name)
	if _, err := os.Stat(base); err == nil {
		return fmt.Errorf("module %s already exists", name)
	}

	dirs := []string{
		filepath.Join(base, "entity"),
		filepath.Join(base, "repository"),
		filepath.Join(base, "repository", "postgres"),
		filepath.Join(base, "usecase"),
		filepath.Join(base, "handler", "http"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	data := templateData{
		ModuleName:      name,
		ModuleGoName:    toCamel(rawName),
		RouteBase:       toPlural(name),
		ImportRoot:      modulePath,
		RepositoryIface: fmt.Sprintf("%sRepository", toCamel(rawName)),
	}

	files := map[string]string{
		filepath.Join(base, "entity", fmt.Sprintf("%s.go", name)):      entityTemplate(data),
		filepath.Join(base, "repository", "repository.go"):             repositoryTemplate(data),
		filepath.Join(base, "repository", "postgres", "repository.go"): repositoryPostgresTemplate(data),
		filepath.Join(base, "usecase", "service.go"):                   usecaseTemplate(data),
		filepath.Join(base, "usecase", "service_test.go"):              usecaseTestTemplate(data),
		filepath.Join(base, "handler", "http", "handler.go"):           handlerTemplate(data),
		filepath.Join(base, "module.go"):                               moduleTemplate(data),
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}

	if err := ensureModuleImport(modulePath, name); err != nil {
		return err
	}

	return gofmtFiles(files)
}

type templateData struct {
	ModuleName      string
	ModuleGoName    string
	RouteBase       string
	ImportRoot      string
	RepositoryIface string
}

func entityTemplate(d templateData) string {
	return fmt.Sprintf(`package entity

import "time"

// %s represents the domain entity for the %s module.
type %s struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
`, d.ModuleGoName, d.ModuleName, d.ModuleGoName)
}

func repositoryTemplate(d templateData) string {
	return fmt.Sprintf(`package repository

import (
	"context"
	"errors"

	"%s/internal/modules/%s/entity"
)

// %s describes data access behaviour for the %s module.
type %s interface {
	GetByID(ctx context.Context, id int64) (*entity.%s, error)
}

// ErrNotFound signals when a record cannot be located.
var ErrNotFound = errors.New("%s not found")
`, d.ImportRoot, d.ModuleName, d.RepositoryIface, d.ModuleName, d.RepositoryIface, d.ModuleGoName, d.ModuleName)
}

func repositoryPostgresTemplate(d templateData) string {
	return fmt.Sprintf(`package postgres

import (
	"context"
	"database/sql"

	"%s/internal/modules/%s/entity"
	"%s/internal/modules/%s/repository"
)

type repositoryImpl struct {
	db *sql.DB
}

// NewRepository wires a PostgreSQL backed repository implementation.
func NewRepository(db *sql.DB) repository.%s {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) GetByID(ctx context.Context, id int64) (*entity.%s, error) {
	const query = %c
		SELECT id, name, created_at, updated_at
		FROM %s
		WHERE id = $1
	%c
	var result entity.%s
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&result.ID, &result.Name, &result.CreatedAt, &result.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &result, nil
}
`, d.ImportRoot, d.ModuleName, d.ImportRoot, d.ModuleName, d.RepositoryIface, d.ModuleGoName, '`', d.RouteBase, '`', d.ModuleGoName)
}

func usecaseTemplate(d templateData) string {
	return fmt.Sprintf(`package usecase

import (
	"context"

	"%s/internal/modules/%s/entity"
	"%s/internal/modules/%s/repository"
)

// Service exposes business capabilities for the %s module.
type Service interface {
	GetByID(ctx context.Context, id int64) (*entity.%s, error)
}

type serviceImpl struct {
	repo repository.%s
}

// NewService constructs the default service implementation.
func NewService(repo repository.%s) Service {
	return &serviceImpl{repo: repo}
}

func (s *serviceImpl) GetByID(ctx context.Context, id int64) (*entity.%s, error) {
	return s.repo.GetByID(ctx, id)
}
`, d.ImportRoot, d.ModuleName, d.ImportRoot, d.ModuleName, d.ModuleName, d.ModuleGoName, d.RepositoryIface, d.RepositoryIface, d.ModuleGoName)
}

func usecaseTestTemplate(d templateData) string {
	return fmt.Sprintf(`package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"%s/internal/modules/%s/entity"
)

type stubRepo struct {
	result *entity.%s
	err    error
}

func (s stubRepo) GetByID(ctx context.Context, id int64) (*entity.%s, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.result, nil
}

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()
	sample := &entity.%s{ID: 1, Name: "%s", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	t.Run("success", func(t *testing.T) {
		service := NewService(stubRepo{result: sample})
		got, err := service.GetByID(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected: %%v", err)
		}
		if got.ID != sample.ID {
			t.Fatalf("expected ID %%d got %%d", sample.ID, got.ID)
		}
	})

	t.Run("error", func(t *testing.T) {
		expected := errors.New("boom")
		service := NewService(stubRepo{err: expected})
		if _, err := service.GetByID(ctx, 1); !errors.Is(err, expected) {
			t.Fatalf("expected %%v got %%v", expected, err)
		}
	})
}
`, d.ImportRoot, d.ModuleName, d.ModuleGoName, d.ModuleGoName, d.ModuleGoName, d.ModuleGoName)
}

func handlerTemplate(d templateData) string {
	return fmt.Sprintf(`package http

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"%s/internal/modules/%s/repository"
	"%s/internal/modules/%s/usecase"
	"%s/internal/server/httpresp"
)

// Handler hosts HTTP endpoints for the %s module.
type Handler struct {
	service usecase.Service
}

// NewHandler wires a Handler instance.
func NewHandler(service usecase.Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes mounts module endpoints.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/%s", func(r chi.Router) {
		r.Get("/{id}", h.GetByID)
	})
}

// GetByID fetches a resource by ID.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		httpresp.Error(w, http.StatusBadRequest, "Invalid id", nil)
		return
	}
	result, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrNotFound {
			httpresp.Error(w, http.StatusNotFound, "Resource not found", err.Error())
			return
		}
		httpresp.Error(w, http.StatusInternalServerError, "Unable to fetch resource", err.Error())
		return
	}
	httpresp.JSON(w, http.StatusOK, "Success", result)
}
`, d.ImportRoot, d.ModuleName, d.ImportRoot, d.ModuleName, d.ImportRoot, d.ModuleName, d.RouteBase)
}

func moduleTemplate(d templateData) string {
	return fmt.Sprintf(`package %[1]s

import (
	"github.com/go-chi/chi/v5"

	"%[2]s/internal/modules/registry"
	modulehttp "%[2]s/internal/modules/%[1]s/handler/http"
	modulerepo "%[2]s/internal/modules/%[1]s/repository/postgres"
	"%[2]s/internal/modules/%[1]s/usecase"
	"%[2]s/internal/server/container"
)

type Module struct {
	handler *modulehttp.Handler
}

func init() {
	registry.Register("%[1]s", func(c *container.Container) (registry.Module, error) {
		return NewModule(c)
	})
}

// NewModule assembles dependencies for the %[1]s module.
func NewModule(c *container.Container) (registry.Module, error) {
	repo := modulerepo.NewRepository(c.DB())
	service := usecase.NewService(repo)
	handler := modulehttp.NewHandler(service)
	return &Module{handler: handler}, nil
}

// Name identifies the module.
func (m *Module) Name() string {
	return "%[1]s"
}

// RegisterRoutes mounts HTTP routes.
func (m *Module) RegisterRoutes(r chi.Router) {
	m.handler.RegisterRoutes(r)
}
`, d.ModuleName, d.ImportRoot)
}

func ensureModuleImport(importRoot, name string) error {
	modulesFile := filepath.Join("internal", "modules", "modules.go")
	content, err := os.ReadFile(modulesFile)
	if err != nil {
		return err
	}
	line := fmt.Sprintf("\t_ \"%s/internal/modules/%s\"\n", importRoot, name)
	if strings.Contains(string(content), line) {
		return nil
	}
	marker := "// Module registrations"
	idx := strings.Index(string(content), marker)
	if idx == -1 {
		return errors.New("modules.go missing module registration marker")
	}
	insertPos := strings.Index(string(content)[idx:], "\n")
	if insertPos == -1 {
		return errors.New("modules.go malformed imports section")
	}
	insertIdx := idx + insertPos + 1
	updated := string(content[:insertIdx]) + line + string(content[insertIdx:])
	return os.WriteFile(modulesFile, []byte(updated), 0o644)
}

func gofmtFiles(files map[string]string) error {
	paths := make([]string, 0, len(files)+1)
	for path := range files {
		paths = append(paths, path)
	}
	paths = append(paths, filepath.Join("internal", "modules", "modules.go"))
	sort.Strings(paths)
	if _, err := exec.LookPath("gofmt"); err != nil {
		return nil
	}
	args := append([]string{"-w"}, paths...)
	cmd := exec.Command("gofmt", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func currentModulePath() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", errors.New("module path not found in go.mod")
}

var validName = regexp.MustCompile(`^[a-z0-9\-]+$`)

func normalizeName(name string) (string, error) {
	lowered := strings.ToLower(strings.TrimSpace(name))
	if lowered == "" || !validName.MatchString(lowered) {
		return "", errors.New("module name must contain lowercase letters, numbers, or hyphen")
	}
	clean := strings.ReplaceAll(lowered, "-", "")
	if clean == "" {
		return "", errors.New("module name must contain alphanumeric characters")
	}
	return clean, nil
}

func toCamel(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_'
	})
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

func toPlural(name string) string {
	if strings.HasSuffix(name, "s") {
		return name
	}
	return name + "s"
}
