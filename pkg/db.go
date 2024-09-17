package pkg

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database wraps the GORM DB instance
type Database struct {
	Conn *gorm.DB
}

// NewDatabase initializes the database connection and returns a Database instance
func NewDatabase(ctx context.Context, dbPath string) *Database {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Migrate the schema
	if err := db.WithContext(ctx).AutoMigrate(&Task{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	return &Database{Conn: db}
}

// CreateTask creates a new task in the database
func (db *Database) CreateTask(ctx context.Context, task *Task) error {
	return db.Conn.WithContext(ctx).Create(task).Error
}

// GetTasks retrieves all tasks for a tenant
func (db *Database) GetTasks(ctx context.Context, tenant string) ([]Task, error) {
	var tasks []Task
	result := db.Conn.WithContext(ctx).Where("tenant = ?", tenant).Find(&tasks)
	return tasks, result.Error
}

// GetTask retrieves a single task by ID for a tenant
func (db *Database) GetTask(ctx context.Context, tenant, id string) (*Task, error) {
	var task Task
	result := db.Conn.WithContext(ctx).Where("id = ? AND tenant = ?", id, tenant).First(&task)
	if result.Error != nil {
		return nil, result.Error
	}
	return &task, nil
}

// UpdateTask updates a task's fields
func (db *Database) UpdateTask(ctx context.Context, task *Task) error {
	return db.Conn.WithContext(ctx).Save(task).Error
}

// DeleteTask performs a soft delete on a task
func (db *Database) DeleteTask(ctx context.Context, tenant, id string) error {
	var task Task
	if err := db.Conn.WithContext(ctx).Where("id = ? AND tenant = ?", id, tenant).First(&task).Error; err != nil {
		return err
	}
	return db.Conn.WithContext(ctx).Delete(&task).Error
}

// UpdateTaskStatus updates the status of a task with state machine validation
func (db *Database) UpdateTaskStatus(ctx context.Context, task *Task, newStatus TaskState) error {
	if !newStatus.IsValid() {
		return fmt.Errorf("invalid task state: %s", newStatus)
	}

	// Check if transition is valid
	if !IsValidTransition(task.Status, newStatus) {
		return fmt.Errorf("invalid state transition from %s to %s", task.Status, newStatus)
	}

	task.Status = newStatus
	task.UpdatedAt = time.Now()

	return db.Conn.WithContext(ctx).Save(task).Error
}
