package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var skipDirs = map[string]struct{}{
	".git":   {},
	"vendor": {},
	"bin":    {},
}

var allowedExt = map[string]struct{}{
	".go":   {},
	".mod":  {},
	".sum":  {},
	".sql":  {},
	".yaml": {},
	".yml":  {},
	".json": {},
	".txt":  {},
	"":      {}, // files without extension (e.g., Makefile)
}

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: go run ./cmd/rename -old <old-module> -new <new-module>")
		flag.PrintDefaults()
	}
	oldModule := flag.String("old", "", "current module/path value")
	newModule := flag.String("new", "", "new module/path value")
	flag.Parse()

	if *oldModule == "" || *newModule == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if _, ok := skipDirs[d.Name()]; ok {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if _, ok := allowedExt[ext]; !ok {
			return nil
		}
		if err := replaceInFile(path, *oldModule, *newModule); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatalf("rename failed: %v", err)
	}

	log.Printf("replacement from %q to %q completed", *oldModule, *newModule)
}

func replaceInFile(path, oldVal, newVal string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	if !strings.Contains(content, oldVal) {
		return nil
	}
	updated := strings.ReplaceAll(content, oldVal, newVal)
	if updated == content {
		return nil
	}
	return os.WriteFile(path, []byte(updated), 0o644)
}
