package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rcdevgames/modular-monolith-clean/internal/config"
)

func main() {
	storageCfg := config.LoadStorageConfig()
	dir := storageCfg.LocalDir
	if dir == "" {
		dir = "upload"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("ensure upload dir: %v", err)
	}
	addr := ":" + storageCfg.CDNPort
	if storageCfg.CDNPort == "" {
		addr = ":9090"
	}
	absDir, _ := filepath.Abs(dir)
	log.Printf("CDN serving %s on %s", absDir, addr)
	fs := http.FileServer(http.Dir(dir))
	if err := http.ListenAndServe(addr, fs); err != nil {
		log.Fatalf("cdn server: %v", err)
	}
}
