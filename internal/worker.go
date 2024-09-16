package internal

import (
	"log"
	"os/exec"
	"time"
)

// Worker encapsulates the dependencies for the worker
type Worker struct {
	DB         *Database
	Kafka      *KafkaConsumer
	KafkaTopic string
}

// NewWorker creates a new Worker instance
func NewWorker(db *Database, kafka *KafkaConsumer, topic string) *Worker {
	return &Worker{
		DB:         db,
		Kafka:      kafka,
		KafkaTopic: topic,
	}
}

// ProcessTask handles the execution of a task
func (w *Worker) ProcessTask(task Task) {
	log.Printf("Executing Task ID: %s - %s %v", task.ID, task.Command, task.Args)

	// Update task status to In Progress
	if err := w.DB.Conn.Model(&Task{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"status":     "In Progress",
		"updated_at": time.Now(),
	}).Error; err != nil {
		log.Printf("Failed to update task status: %v", err)
		return
	}

	// Execute the command
	cmd := exec.Command(task.Command, task.Args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Task ID %s failed: %v. Output: %s", task.ID, err, string(output))
		// Update task status to Failed
		if err := w.DB.Conn.Model(&Task{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
			"status":     "Failed",
			"updated_at": time.Now(),
		}).Error; err != nil {
			log.Printf("Failed to update task status: %v", err)
		}
		return
	}

	log.Printf("Task ID %s completed. Output: %s", task.ID, string(output))

	// Update task status to Completed
	if err := w.DB.Conn.Model(&Task{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"status":     "Completed",
		"updated_at": time.Now(),
	}).Error; err != nil {
		log.Printf("Failed to update task status: %v", err)
	}
}
