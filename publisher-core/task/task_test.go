package task

import (
	"testing"
	"time"
)

func TestTaskStatusValues(t *testing.T) {
	tests := []struct {
		status TaskStatus
		want   string
	}{
		{TaskStatusPending, "pending"},
		{TaskStatusRunning, "running"},
		{TaskStatusCompleted, "completed"},
		{TaskStatusFailed, "failed"},
		{TaskStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("Status = %v, want %v", tt.status, tt.want)
		}
	}
}

func TestNewTaskManager(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewTaskManager(storage)

	if manager == nil {
		t.Fatal("TaskManager should not be nil")
	}
}

func TestTaskManagerCreateAndGetTask(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewTaskManager(storage)

	task, err := manager.CreateTask("publish", "douyin", map[string]interface{}{
		"title": "Test Content",
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if task.ID == "" {
		t.Error("Task ID should not be empty")
	}

	if task.Type != "publish" {
		t.Errorf("Expected task type 'publish', got %s", task.Type)
	}

	if task.Platform != "douyin" {
		t.Errorf("Expected platform 'douyin', got %s", task.Platform)
	}

	if task.Status != TaskStatusPending {
		t.Errorf("Expected status %s, got %s", TaskStatusPending, task.Status)
	}

	retrieved, err := manager.GetTask(task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if retrieved.ID != task.ID {
		t.Errorf("Task ID mismatch, got %s, want %s", retrieved.ID, task.ID)
	}
}

func TestTaskManagerListTasks(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewTaskManager(storage)

	_, _ = manager.CreateTask("publish", "douyin", nil)
	_, _ = manager.CreateTask("publish", "xiaohongshu", nil)
	_, _ = manager.CreateTask("analytics", "douyin", nil)

	filter := TaskFilter{Limit: 10}
	allTasks, err := manager.ListTasks(filter)
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(allTasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(allTasks))
	}

	filter = TaskFilter{Platform: "douyin", Limit: 10}
	douyinTasks, err := manager.ListTasks(filter)
	if err != nil {
		t.Fatalf("ListTasks with filter failed: %v", err)
	}
	if len(douyinTasks) != 2 {
		t.Errorf("Expected 2 douyin tasks, got %d", len(douyinTasks))
	}
}

func TestTaskManagerCancel(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewTaskManager(storage)

	task, _ := manager.CreateTask("test", "test", nil)

	err := manager.Cancel(task.ID)
	if err != nil {
		t.Errorf("Cancel() failed: %v", err)
	}

	retrieved, _ := manager.GetTask(task.ID)
	if retrieved.Status != TaskStatusCancelled {
		t.Errorf("Expected status %s after Cancel(), got %s", TaskStatusCancelled, retrieved.Status)
	}
}

func TestTaskTiming(t *testing.T) {
	task := &Task{
		ID:        "test",
		Status:    TaskStatusCompleted,
		CreatedAt: time.Now(),
	}

	startTime := time.Now()
	task.StartedAt = &startTime

	endTime := startTime.Add(2 * time.Second)
	task.FinishedAt = &endTime

	if task.StartedAt == nil {
		t.Error("StartedAt should be set")
	}

	if task.FinishedAt == nil {
		t.Error("FinishedAt should be set")
	}
}

func TestMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()

	task := &Task{
		ID:        "test-1",
		Type:      "publish",
		Platform:  "douyin",
		Status:    TaskStatusPending,
		CreatedAt: time.Now(),
	}

	err := storage.Save(task)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := storage.Load("test-1")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if retrieved.ID != task.ID {
		t.Errorf("Task ID mismatch, got %s, want %s", retrieved.ID, task.ID)
	}

	tasks, err := storage.List(TaskFilter{Limit: 10})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}
}
