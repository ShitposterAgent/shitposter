package agent

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
)

// Config holds configuration for the agent
// Extend this struct for modular config
type Config struct {
	// Add config fields here
}

// ClientSet holds external service clients (Ollama, etc.)
type ClientSet struct {
	// Add client fields here
}

// Module is an interface for agent modules
// Each module should implement Start/Stop methods
// Example: type MyModule struct{}
// func (m *MyModule) Start(ctx context.Context) error { ... }
// func (m *MyModule) Stop() error { ... }
type Module interface {
	Start(ctx context.Context) error
	Stop() error
}

// Agent struct holds the core state and logic
// Add modules to the Modules slice for extensibility
// Add config and clients for modularity
type Agent struct {
	Config  *Config
	Clients *ClientSet
	Modules []Module
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewAgent creates a new agent instance
func NewAgent(cfg *Config, clients *ClientSet, modules []Module) (*Agent, error) {
	ctx, cancel := context.WithCancel(context.Background())
	a := &Agent{
		Config:  cfg,
		Clients: clients,
		Modules: modules,
		ctx:     ctx,
		cancel:  cancel,
	}
	fmt.Println("Initializing core agent...")
	return a, nil
}

// RegisterModule allows dynamic registration of modules
func (a *Agent) RegisterModule(m Module) {
	a.Modules = append(a.Modules, m)
}

// StartAllModules starts all registered modules
func (a *Agent) StartAllModules() {
	for _, m := range a.Modules {
		a.wg.Add(1)
		go func(mod Module) {
			defer a.wg.Done()
			_ = mod.Start(a.ctx)
		}(m)
	}
}

// StopAllModules stops all registered modules
func (a *Agent) StopAllModules() {
	for _, m := range a.Modules {
		_ = m.Stop()
	}
}

// defaultSignalHandler handles graceful shutdown on interrupt
func defaultSignalHandler(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("\nInterrupt received, shutting down...")
	cancel()
}

func (a *Agent) Start() error {
	fmt.Println("Starting agent logic...")
	a.StartAllModules()
	go defaultSignalHandler(a.cancel)
	<-a.ctx.Done()
	a.StopAllModules()
	a.wg.Wait()
	fmt.Println("Agent stopped.")
	return nil
}

// ExampleModule demonstrates a simple agent module
// Replace or extend this with your own modules
// This module prints a message every second until stopped
type ExampleModule struct {
	stop chan struct{}
}

func NewExampleModule() *ExampleModule {
	return &ExampleModule{stop: make(chan struct{})}
}

func (m *ExampleModule) Start(ctx context.Context) error {
	fmt.Println("ExampleModule started.")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("ExampleModule received stop signal.")
			return nil
		case <-m.stop:
			fmt.Println("ExampleModule stopped by internal signal.")
			return nil
		case <-time.After(time.Second):
			fmt.Println("ExampleModule is running...")
		}
	}
}

func (m *ExampleModule) Stop() error {
	close(m.stop)
	return nil
}
