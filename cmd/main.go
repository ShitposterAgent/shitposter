package main

import (
	"fmt"
	"os"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"shitposter/internal/agent"
)

func main() {
	_ = godotenv.Load() // Load .env if present

	var rootCmd = &cobra.Command{
		Use:   "shitposter",
		Short: "Shitposter Engine",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting Shitposter Engine/API...")
			runEngine()
		},
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "ollama",
		Short: "Interact with Ollama module",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Running Ollama module...")
			// You can add custom logic for ollama subcommand here
			runEngine() // Optionally start engine as well
		},
	})

	// Add more subcommands as needed

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runEngine() {
	cfg := &agent.Config{}
	clients := &agent.ClientSet{}
	ollama := agent.NewOllamaClient()
	api := agent.NewAPIModule(ollama)
	a, _ := agent.NewAgent(cfg, clients, []agent.Module{api})
	a.Start()
}
