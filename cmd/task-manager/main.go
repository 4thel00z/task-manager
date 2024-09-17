package main

import (
	"github.com/4thel00z/task-manager/pkg/cmd"
	"github.com/spf13/cobra"
	"log"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "task-manager",
		Short: "Task Manager is a CLI application to manage and execute tasks",
	}

	// Add subcommands
	rootCmd.AddCommand(cmd.ServeCmd)
	rootCmd.AddCommand(cmd.WorkerCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
