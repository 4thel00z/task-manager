package main

import (
    "encoding/json"
    "log"
    "os/exec"
    "time"

    "task-manager/internal"
)

// Worker encapsulates the dependencies for the worker
type Worker struct {
    DB         *internal.Database
    Kafka      *internal.KafkaConsumer
    KafkaTopic string
}

// NewWorker creates a new Worker instance
func NewWorker(db *internal.Database, kafka *internal.KafkaConsumer, topic string) *Worker {
    return &Worker{
        DB:         db,
        Kafka:      kafka,
        KafkaTopic: topic,
    }
}

// ProcessTask handles the execution of a task
func (w *Worker) ProcessTask(task internal.Task) {
    log.Printf("Processing Task ID: %s - %s %v", task.ID, task.Command, task.Args)

    // Update task status to In Progress
    if err := w.DB.UpdateTaskStatus(context.Background(), &task, internal.InProgress); err != nil {
        log.Printf("Failed to update task status to In Progress: %v", err)
        return
    }

    // Execute the command
    cmd := exec.Command(task.Command, task.Args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("Task ID %s failed: %v. Output: %s", task.ID, err, string(output))
        // Update task status to Failed
        if err := w.DB.UpdateTaskStatus(context.Background(), &task, internal.Failed); err != nil {
            log.Printf("Failed to update task status to Failed: %v", err)
        }
        return
    }

    log.Printf("Task ID %s completed. Output: %s", task.ID, string(output))

    // Update task status to Completed
    if err := w.DB.UpdateTaskStatus(context.Background(), &task, internal.Completed); err != nil {
        log.Printf("Failed to update task status to Completed: %v", err)
    }
}
