package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

//go:embed static
var staticFiles embed.FS

// Start registers all routes and starts listening on the given port.
// It returns the *http.Server so the caller can shut it down gracefully.
func Start(port int) (*http.Server, error) {
	mux := http.NewServeMux()

	// Static SPA — strip the "static/" prefix from the embedded FS.
	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return nil, fmt.Errorf("embed sub: %w", err)
	}
	// Wrap with no-cache middleware so browsers always fetch the latest binary build.
	mux.Handle("/", noCacheMiddleware(http.FileServer(http.FS(subFS))))

	// API — order matters: exact paths must be registered before prefix paths.
	mux.HandleFunc("/api/stats/history", handleStatsHistory)
	mux.HandleFunc("/api/stats", handleStats)
	mux.HandleFunc("/api/calendar", handleCalendar)
	mux.HandleFunc("/api/export", handleExport)
	mux.HandleFunc("/api/topics/review", handleTopicsReview)
	mux.HandleFunc("/api/topics", handleTopics)
	mux.HandleFunc("/api/topics/", handleTopic)
	mux.HandleFunc("/api/projects", handleProjects)
	mux.HandleFunc("/api/projects/", handleProject)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("server error: %v\n", err)
		}
	}()

	return srv, nil
}

// OpenBrowser opens the given URL in the system default browser.
func OpenBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

// noCacheMiddleware sets headers that prevent browsers from serving stale cached pages.
func noCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

// Shutdown gracefully stops the server with a 5-second timeout.
func Shutdown(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
