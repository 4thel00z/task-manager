package cmd

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"log"

	"github.com/4thel00z/task-manager/pkg"
	"github.com/kataras/iris/v12"
)

func init() {
	ServeCmd.PersistentFlags().Int("port", 1337, "Port to run the server on")
	ServeCmd.PersistentFlags().String("host", "", "Host to run the server on")
}

// ServeCmd represents the serve command
var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the Iris web server",
	RunE: func(cmd *cobra.Command, args []string) error {
		host, err := cmd.Flags().GetString("host")
		if err != nil {
			return errors.Wrapf(err, "Failed to get host flag")
		}
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return errors.Wrapf(err, "Failed to get port flag")
		}
		// Initialize the database with context
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		db := pkg.NewDatabase(ctx, "tasks.db")

		// Initialize Kafka producer
		kafka := pkg.NewKafkaProducer([]string{"localhost:9092"})
		defer func() {
			if err := kafka.Producer.Close(); err != nil {
				log.Printf("Error closing Kafka producer: %v", err)
			}
		}()

		// Initialize middleware
		middleware := pkg.NewMiddleware()

		// Initialize handlers
		handlers := pkg.NewHandlers(db, kafka, "tasks")

		// Initialize Iris application
		app := iris.New()

		// Register middleware
		app.Use(middleware.TenantMiddleware)

		// Define routes
		app.Post("/tasks", handlers.CreateTask)
		app.Get("/tasks", handlers.GetTasks)
		app.Get("/tasks/{id}", handlers.GetTask)
		app.Put("/tasks/{id}", handlers.UpdateTask)
		app.Delete("/tasks/{id}", handlers.DeleteTask)

		// Start the server
		log.Printf("Starting server on :%d\n", port)
		if err := app.Listen(fmt.Sprintf("%s:%d", host, port)); err != nil {
			return errors.Wrap(err, "Failed to start server")
		}
		return nil
	},
}
