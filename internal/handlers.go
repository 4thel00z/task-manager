package internal

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// Handlers holds the dependencies for HTTP handlers
type Handlers struct {
	DB         *Database
	Kafka      *KafkaProducer
	KafkaTopic string
}

// NewHandlers creates a new Handlers instance
func NewHandlers(db *Database, kafka *KafkaProducer, topic string) *Handlers {
	return &Handlers{
		DB:         db,
		Kafka:      kafka,
		KafkaTopic: topic,
	}
}

// CreateTask handles the creation of a new task
func (h *Handlers) CreateTask(ctx iris.Context) {
	tenant := ctx.Values().GetString("tenant")

	var task Task
	if err := ctx.ReadJSON(&task); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	// Generate a new UUID
	task.ID = uuid.New().String()
	task.Tenant = tenant
	task.Status = Pending
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	// Validate initial state
	if !task.Status.IsValid() {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid task status"})
		return
	}

	// Save to the database
	if err := h.DB.CreateTask(ctx.Request().Context(), &task); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to create task"})
		return
	}

	// Enqueue the task to Kafka
	if err := h.Kafka.EnqueueTask(task, h.KafkaTopic); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to enqueue task"})
		return
	}

	ctx.StatusCode(http.StatusCreated)
	ctx.JSON(task)
}

// GetTasks retrieves all tasks for the tenant
func (h *Handlers) GetTasks(ctx iris.Context) {
	tenant := ctx.Values().GetString("tenant")

	tasks, err := h.DB.GetTasks(ctx.Request().Context(), tenant)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to retrieve tasks"})
		return
	}
	ctx.JSON(tasks)
}

// GetTask retrieves a single task by ID for the tenant
func (h *Handlers) GetTask(ctx iris.Context) {
	tenant := ctx.Values().GetString("tenant")
	id := ctx.Params().Get("id")

	task, err := h.DB.GetTask(ctx.Request().Context(), tenant, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusNotFound)
			ctx.JSON(iris.Map{"error": "Task not found"})
		} else {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "Failed to retrieve task"})
		}
		return
	}
	ctx.JSON(task)
}

// UpdateTask updates a task for the tenant
func (h *Handlers) UpdateTask(ctx iris.Context) {
	tenant := ctx.Values().GetString("tenant")
	id := ctx.Params().Get("id")

	task, err := h.DB.GetTask(ctx.Request().Context(), tenant, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusNotFound)
			ctx.JSON(iris.Map{"error": "Task not found"})
		} else {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "Failed to retrieve task"})
		}
		return
	}

	var updatedData Task
	if err := ctx.ReadJSON(&updatedData); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(iris.Map{"error": "Invalid request payload"})
		return
	}

	// Update fields
	if updatedData.Command != "" {
		task.Command = updatedData.Command
	}
	if updatedData.Args != nil {
		task.Args = updatedData.Args
	}

	// Handle status update with state machine
	if updatedData.Status != "" && updatedData.Status != task.Status {
		if err := h.DB.UpdateTaskStatus(ctx.Request().Context(), task, updatedData.Status); err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(iris.Map{"error": err.Error()})
			return
		}
	}

	// Update UpdatedAt
	task.UpdatedAt = time.Now()

	if err := h.DB.UpdateTask(ctx.Request().Context(), task); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(iris.Map{"error": "Failed to update task"})
		return
	}

	ctx.JSON(task)
}

// DeleteTask performs a soft delete on a task for the tenant
func (h *Handlers) DeleteTask(ctx iris.Context) {
	tenant := ctx.Values().GetString("tenant")
	id := ctx.Params().Get("id")

	if err := h.DB.DeleteTask(ctx.Request().Context(), tenant, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.StatusCode(http.StatusNotFound)
			ctx.JSON(iris.Map{"error": "Task not found"})
		} else {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(iris.Map{"error": "Failed to delete task"})
		}
		return
	}

	ctx.StatusCode(http.StatusNoContent)
}
