package internal

import (
    "encoding/json"
    "time"

    "gorm.io/gorm"
)

// TaskState represents the state of a task
type TaskState string

const (
    Pending    TaskState = "Pending"
    InProgress TaskState = "In Progress"
    Completed  TaskState = "Completed"
    Failed     TaskState = "Failed"
)

// IsValid checks if the TaskState is valid
func (ts TaskState) IsValid() bool {
    switch ts {
    case Pending, InProgress, Completed, Failed:
        return true
    }
    return false
}

// Task represents a subcommand to be executed
type Task struct {
    ID        string         `json:"id" gorm:"primaryKey"`
    Tenant    string         `json:"tenant" gorm:"index;not null"`
    Command   string         `json:"command"`
    Args      StringSlice    `json:"args" gorm:"serializer:json"`
    Status    TaskState      `json:"status"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// StringSlice is a custom type to handle []string serialization
type StringSlice []string

// MarshalJSON serializes the string slice to JSON
func (s StringSlice) MarshalJSON() ([]byte, error) {
    return json.Marshal([]string(s))
}

// UnmarshalJSON deserializes JSON to the string slice
func (s *StringSlice) UnmarshalJSON(data []byte) error {
    var temp []string
    if err := json.Unmarshal(data, &temp); err != nil {
        return err
    }
    *s = temp
    return nil
}
