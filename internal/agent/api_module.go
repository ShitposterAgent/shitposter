package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// APIModule is a fully modular HTTP API server for the engine
// It implements the Module interface and delegates to injected services

type APIModule struct {
	host   string
	port   string
	server *http.Server
	stop   chan struct{}
	ollama OllamaService // interface, not concrete type
}

// OllamaService defines the interface for Ollama operations
// This allows the API module to remain decoupled

type OllamaService interface {
	Generate(ctx context.Context, req OllamaGenerateRequest) (OllamaGenerateResponse, error)
}

func NewAPIModule(ollama OllamaService) *APIModule {
	host := os.Getenv("API_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	return &APIModule{
		host:   host,
		port:   port,
		stop:   make(chan struct{}),
		ollama: ollama,
	}
}

func (m *APIModule) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/ping", m.handlePing)
	mux.HandleFunc("/api/ollama/generate", m.handleOllamaGenerate)
	// Add more endpoints and delegate to other modules/services as needed

	m.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", m.host, m.port),
		Handler: mux,
	}
	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
		}
	}()
	log.Printf("API server started at http://%s:%s", m.host, m.port)
	<-ctx.Done()
	_ = m.server.Shutdown(context.Background())
	return nil
}

func (m *APIModule) Stop() error {
	close(m.stop)
	return nil
}

func (m *APIModule) handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (m *APIModule) handleOllamaGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req OllamaGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	resp, err := m.ollama.Generate(r.Context(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
