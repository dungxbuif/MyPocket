package httpapi

import (
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"mypocket/internal/identity"
	"mypocket/internal/platform/config"
)

func DocsHandler(cfg config.Config) http.HandlerFunc {
	dir := cfg.DocsDir
	if dir == "" {
		return func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		log.Printf("docs: invalid DOCS_DIR %q: %v", dir, err)
		return func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-store")
		if !isAuthenticated(r, cfg) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/docs")
		if path == "" || path == "/" {
			path = "/index.html"
		}

		fullPath := filepath.Join(absDir, path)
		info, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				path = "/index.html"
				fullPath = filepath.Join(absDir, path)
			}
		} else if info.IsDir() {
			path = filepath.Join(path, "index.html")
			fullPath = filepath.Join(absDir, path)
		}

		file, err := os.Open(fullPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()
		info, err = file.Stat()
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		if contentType := mime.TypeByExtension(filepath.Ext(fullPath)); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		http.ServeContent(w, r, filepath.Base(fullPath), info.ModTime(), file)
	}
}

func isAuthenticated(r *http.Request, cfg config.Config) bool {
	if authenticatedUserIDFromContext(r.Context()) != "" {
		return true
	}
	cookie, err := r.Cookie(identity.AuthCookieName)
	if err != nil {
		return false
	}
	signer := identity.NewCookieSigner([]byte(cfg.CookieSecret))
	claims, err := signer.Verify(cookie.Value)
	if err != nil {
		return false
	}
	return claims.UserID != ""
}
