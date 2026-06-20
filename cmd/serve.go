package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"spaced/web"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launch the web UI in your browser",
	Long:  "Start a local HTTP server and open the spaced web UI in your default browser.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		port, _ := cmd.Flags().GetInt("port")
		noBrowser, _ := cmd.Flags().GetBool("no-browser")

		srv, err := web.Start(port)
		if err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}

		url := fmt.Sprintf("http://localhost:%d", port)
		fmt.Fprintf(os.Stdout, "Serving at %s — press Ctrl+C to stop\n", url)

		if !noBrowser {
			go func() {
				time.Sleep(500 * time.Millisecond)
				web.OpenBrowser(url)
			}()
		}

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		fmt.Fprintln(os.Stdout, "\nShutting down...")
		web.Shutdown(srv)
		return nil
	},
}

func init() {
	serveCmd.Flags().Int("port", 7331, "Port to listen on")
	serveCmd.Flags().Bool("no-browser", false, "Do not open the browser automatically")
	rootCmd.AddCommand(serveCmd)
}
