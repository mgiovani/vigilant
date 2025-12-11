package youtube

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)

// EmbedProxy serves YouTube embed HTML over HTTP localhost
// This gives the embed a proper http:// origin that YouTube accepts
type EmbedProxy struct {
	port     int
	server   *http.Server
	listener net.Listener
	mu       sync.Mutex
	running  bool
}

// NewEmbedProxy creates a new embed proxy server
func NewEmbedProxy() *EmbedProxy {
	return &EmbedProxy{}
}

// Start starts the HTTP server on an available port
func (p *EmbedProxy) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return nil
	}

	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}

	p.listener = listener
	p.port = listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/embed/", p.handleEmbed)

	p.server = &http.Server{
		Handler: mux,
	}

	go func() {
		log.Printf("YouTube embed proxy started on http://127.0.0.1:%d", p.port)
		if err := p.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("Embed proxy server error: %v", err)
		}
	}()

	p.running = true
	return nil
}

// Stop stops the HTTP server
func (p *EmbedProxy) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	if p.server != nil {
		if err := p.server.Close(); err != nil {
			return err
		}
	}

	p.running = false
	log.Printf("YouTube embed proxy stopped")
	return nil
}

// GetPort returns the port the server is running on
func (p *EmbedProxy) GetPort() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.port
}

// GetEmbedURL returns the full URL to access the YouTube embed
func (p *EmbedProxy) GetEmbedURL(videoID string) string {
	return fmt.Sprintf("http://127.0.0.1:%d/embed/%s", p.port, videoID)
}

// handleEmbed serves the YouTube embed HTML page
func (p *EmbedProxy) handleEmbed(w http.ResponseWriter, r *http.Request) {
	// Extract video ID from path: /embed/{videoID}
	videoID := r.URL.Path[len("/embed/"):]
	if videoID == "" {
		videoID = "jfKfPfyJRdk" // Default fallback
	}

	// Set proper headers for the HTML page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	// Allow embedding from the Wails app
	w.Header().Set("X-Frame-Options", "ALLOWALL")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Serve the YouTube embed HTML
	// This page has http://127.0.0.1 as its origin, which YouTube accepts
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="referrer" content="strict-origin-when-cross-origin">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Lofi Player</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        html, body { width: 100%%; height: 100%%; overflow: hidden; background: #030712; }
        iframe {
            width: 100%%;
            height: 100%%;
            border: none;
            display: block;
        }
    </style>
</head>
<body>
    <iframe
        src="https://www.youtube-nocookie.com/embed/%s?autoplay=1&controls=1&modestbranding=1&rel=0&playsinline=1"
        referrerpolicy="strict-origin-when-cross-origin"
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
        allowfullscreen>
    </iframe>
</body>
</html>`, videoID)

	w.Write([]byte(html))
}
