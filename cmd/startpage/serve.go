package startpage

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the built startpage on a local web server",
		Long: `Serve the built startpage on a local web server (required for theme to work).

Examples:
  ops-cli startpage serve
  ops-cli startpage serve --port 8080`,
		RunE: runServe,
	}

	cmd.Flags().Int("port", 3000, "Server port")
	cmd.Flags().Bool("background", false, "Run server in background")

	return cmd
}

func runServe(cmd *cobra.Command, args []string) error {
	if err := ensureStartpageDirectory(); err != nil {
		return fmt.Errorf("failed to create startpage directory: %w", err)
	}

	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("no startpage found. Run 'ops-cli startpage init' first")
	}

	distPath := filepath.Join(config.Path, "dist")

	// Check if dist directory exists
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		return fmt.Errorf("no built startpage found. Run 'ops-cli startpage build' first")
	}

	port, _ := cmd.Flags().GetInt("port")
	background, _ := cmd.Flags().GetBool("background")

	if background {
		// Background mode - simplified, just print instructions
		fmt.Printf("Starting startpage server in background...\n")
		fmt.Printf("Serving: %s\n", distPath)
		fmt.Printf("URL: http://localhost:%d\n", port)
		fmt.Println("\nSet this URL as your browser homepage!")
		fmt.Println("\nNote: Background mode not fully implemented. Use foreground mode for now.")
		return nil
	}

	// Foreground mode
	fmt.Printf("Starting static file server for startpage...\n")
	fmt.Printf("Serving: %s\n", distPath)
	fmt.Printf("URL: http://localhost:%d\n", port)
	fmt.Println("\nSet this URL as your browser homepage!")
	fmt.Println("\nPress Ctrl+C to stop the server\n")

	// Create file server
	fs := http.FileServer(http.Dir(distPath))
	http.Handle("/", fs)

	// Start server
	addr := ":" + strconv.Itoa(port)
	fmt.Printf("Server listening on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}
