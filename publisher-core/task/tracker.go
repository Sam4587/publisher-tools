package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TaskEvent struct {
	ID        string                 `json:"id"`
	TaskID    string                 `json:"task_id"`
	Type      string                 `json:"type"`
	Message   string                 `json:"message"`
	Progress  int                    `json:"progress"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type TaskTracker struct {
	mu       sync.RWMutex
	storage  EventStorage
	notifyCh chan TaskEvent
	watchers map[string][]TaskWatcher
}

type TaskWatcher func(event TaskEvent)

type EventStorage interface {
	SaveEvent(event *TaskEvent) error
	ListEvents(taskID string, limit int) ([]*TaskEvent, error)
}

func NewTaskTracker(storage EventStorage) *TaskTracker {
	return &TaskTracker{
		storage:  storage,
		notifyCh: make(chan TaskEvent, 1000),
		watchers: make(map[string][]TaskWatcher),
	}
}

func (t *TaskTracker) RecordEvent(event *TaskEvent) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	event.Timestamp = time.Now()

	if t.storage != nil {
		if err := t.storage.SaveEvent(event); err != nil {
			logrus.Warnf("Failed to save task event: %v", err)
		}
	}

	select {
	case t.notifyCh <- *event:
	default:
		logrus.Warn("Task event channel full, dropping event")
	}

	if watchers, ok := t.watchers[event.TaskID]; ok {
		for _, watcher := range watchers {
			go watcher(*event)
		}
	}

	return nil
}

func (t *TaskTracker) WatchTask(taskID string, watcher TaskWatcher) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.watchers[taskID] = append(t.watchers[taskID], watcher)
}

func (t *TaskTracker) GetTaskHistory(taskID string, limit int) ([]*TaskEvent, error) {
	if t.storage == nil {
		return nil, fmt.Errorf("storage not initialized")
	}
	return t.storage.ListEvents(taskID, limit)
}

func (t *TaskTracker) Notify() <-chan TaskEvent {
	return t.notifyCh
}

type JSONEventStorage struct {
	dataDir string
	mu      sync.RWMutex
}

func NewJSONEventStorage(dataDir string) (*JSONEventStorage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	return &JSONEventStorage{dataDir: dataDir}, nil
}

func (s *JSONEventStorage) SaveEvent(event *TaskEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dateDir := filepath.Join(s.dataDir, event.Timestamp.Format("2006-01-02"))
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s_%s.json", event.TaskID, event.ID)
	path := filepath.Join(dateDir, filename)

	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (s *JSONEventStorage) ListEvents(taskID string, limit int) ([]*TaskEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*TaskEvent

	dirs, err := filepath.Glob(filepath.Join(s.dataDir, "*"))
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		files, err := filepath.Glob(filepath.Join(dir, taskID+"_*.json"))
		if err != nil {
			continue
		}

		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			var event TaskEvent
			if err := json.Unmarshal(data, &event); err != nil {
				continue
			}

			events = append(events, &event)
		}
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, nil
}

type TaskProgressReporter struct {
	tracker *TaskTracker
	taskID  string
}

func NewTaskProgressReporter(tracker *TaskTracker, taskID string) *TaskProgressReporter {
	return &TaskProgressReporter{
		tracker: tracker,
		taskID:  taskID,
	}
}

func (r *TaskProgressReporter) ReportProgress(progress int, message string, metadata map[string]interface{}) error {
	event := &TaskEvent{
		TaskID:   r.taskID,
		Type:     "progress",
		Message:  message,
		Progress: progress,
		Metadata: metadata,
	}
	return r.tracker.RecordEvent(event)
}

func (r *TaskProgressReporter) ReportStart(message string) error {
	event := &TaskEvent{
		TaskID:  r.taskID,
		Type:    "started",
		Message: message,
	}
	return r.tracker.RecordEvent(event)
}

func (r *TaskProgressReporter) ReportComplete(message string, metadata map[string]interface{}) error {
	event := &TaskEvent{
		TaskID:   r.taskID,
		Type:     "completed",
		Message:  message,
		Progress: 100,
		Metadata: metadata,
	}
	return r.tracker.RecordEvent(event)
}

func (r *TaskProgressReporter) ReportError(message string, err error) error {
	event := &TaskEvent{
		TaskID:  r.taskID,
		Type:    "failed",
		Message: message,
		Metadata: map[string]interface{}{
			"error": err.Error(),
		},
	}
	return r.tracker.RecordEvent(event)
}
