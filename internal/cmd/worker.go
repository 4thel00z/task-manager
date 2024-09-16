package cmd

import (
	"context"
	"log"
	"time"

	"github.com/4thel00z/task-manager/internal"
	"github.com/spf13/cobra"
)

// WorkerCmd represents the worker command
var WorkerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Starts the worker to process tasks",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize the database with context
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		db := internal.NewDatabase(ctx, "tasks.db")

		// Initialize Kafka consumer
		kafkaConsumer := internal.NewKafkaConsumer([]string{"localhost:9092"})
		defer func() {
			if err := kafkaConsumer.Consumer.Close(); err != nil {
				log.Printf("Error closing Kafka consumer: %v", err)
			}
		}()

		// Initialize worker
		worker := internal.NewWorker(db, kafkaConsumer, "tasks")

		// Start consuming tasks
		worker.Kafka.ConsumeTasks(worker.KafkaTopic, worker.ProcessTask)

		// Keep the worker running
		log.Println("Worker started. Listening for tasks...")
		for {
			time.Sleep(1 * time.Hour)
		}
	},
}
