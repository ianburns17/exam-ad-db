package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"
)

//go:embed ui/*
var uiFS embed.FS

func main() {
	addr := os.Getenv("UI_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	apiBase := os.Getenv("API_BASE_URL")
	if apiBase == "" {
		apiBase = "http://localhost:4000"
	}

	mux := http.NewServeMux()

	// Static assets (embedded).
	uiSub, err := fs.Sub(uiFS, "ui")
	if err != nil {
		log.Fatal(err)
	}
	staticHandler := http.FileServer(http.FS(uiSub))
	mux.Handle("/ui/", http.StripPrefix("/ui/", staticHandler))

	// App entrypoint (inject API base URL).
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Vehicles</title>
  <link rel="stylesheet" href="/ui/styles.css" />
</head>
<body>
  <div id="app"></div>
  <script>
    window.__API_BASE_URL__ = %q;
  </script>
  <script defer src="/ui/app.js"></script>
</body>
</html>`, apiBase)
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("UI server running on http://localhost%s (API_BASE_URL=%s)", addr, apiBase)
	log.Fatal(srv.ListenAndServe())
}
