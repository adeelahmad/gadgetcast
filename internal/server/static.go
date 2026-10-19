package server

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

func noUINotice() string {
	return `GadgetCast server ` + Version + `

The Studio UI has not been built into this binary (web/dist/index.html is missing).
Build the UI into web/dist and rebuild the server, or run the UI dev server
against this API. The API is available under /api/ and signaling at /ws.
`
}

// staticHandler serves the embedded SPA: real files as-is, any other path
// falls back to index.html so client-side routes work on reload.
func (s *Server) staticHandler() http.Handler {
	fsys := s.opts.Static
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if fsys != nil && name != "" && name != "index.html" && !hiddenPath(name) {
			if st, err := fs.Stat(fsys, name); err == nil && !st.IsDir() {
				http.ServeFileFS(w, r, fsys, name)
				return
			}
		}
		var index []byte
		if fsys != nil {
			index, _ = fs.ReadFile(fsys, "index.html")
		}
		if index == nil {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte(noUINotice()))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(index)
	})
}

// hiddenPath reports whether any element of name is a dotfile (e.g. .gitkeep).
func hiddenPath(name string) bool {
	for _, part := range strings.Split(name, "/") {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}
