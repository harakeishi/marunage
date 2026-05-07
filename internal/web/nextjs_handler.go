package web

import (
	"io/fs"
	"net/http"
	"strings"
)

// newNextJSHandler serves a Next.js static export.
// For paths that don't match a regular file, it serves index.html (SPA fallback).
func newNextJSHandler(njs fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		// Try the path itself first. If it's a directory (e.g. "tasks"), look
		// for the prerendered "<dir>/index.html" that `next build` emits with
		// trailingSlash. Anything still missing falls back to the root
		// index.html so Next.js client-side routing can take over.
		path = resolveNextJSPath(njs, path)
		http.ServeFileFS(w, r, njs, path)
	})
}

// resolveNextJSPath maps a request path to a concrete file inside the embedded
// Next.js static export. Returns "index.html" for anything that cannot be
// resolved so the SPA fallback handles client-side routes.
func resolveNextJSPath(njs fs.FS, path string) string {
	fi, err := fs.Stat(njs, path)
	if err == nil && !fi.IsDir() {
		return path
	}
	if err == nil && fi.IsDir() {
		idx := path + "/index.html"
		if fi2, err2 := fs.Stat(njs, idx); err2 == nil && !fi2.IsDir() {
			return idx
		}
	}
	// Treat ErrNotExist, ErrInvalid (e.g. trailing slash that fs.ValidPath
	// rejects), and any other miss as "not found" so the SPA fallback wins
	// rather than returning 500 for a static-asset request.
	return "index.html"
}
