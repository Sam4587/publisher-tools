package analytics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ScheduledCollector struct {
	mu         sync.RWMutex
	service    *AnalyticsService
	interval   time.Duration
	ticker     *time.Ticker
	running    bool
	taskQueue  []CollectionTask
	maxWorkers int
}

type CollectionTask struct {
	ID         string
	Platform   Platform
	Type       string // "post" or "account"
	TargetID   string
	Priority   int
	CreatedAt  time.Time
	LastError  string
	Retries    int
	MaxRetries int
}

func NewScheduledCollector(service *AnalyticsService, interval time.Duration) *ScheduledCollector {
	return &ScheduledCollector{
		service:    service,
		interval:   interval,
		taskQueue:  make([]CollectionTask, 0),
		maxWorkers: 3,
	}
}

func (sc *ScheduledCollector) Start(ctx context.Context) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.running {
		return nil
	}

	sc.running = true
	sc.ticker = time.NewTicker(sc.interval)

	go sc.run(ctx)

	logrus.Info("Scheduled collector started")
	return nil
}

func (sc *ScheduledCollector) Stop() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if !sc.running {
		return
	}

	sc.running = false
	if sc.ticker != nil {
		sc.ticker.Stop()
	}

	logrus.Info("Scheduled collector stopped")
}

func (sc *ScheduledCollector) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			sc.Stop()
			return
		case <-sc.ticker.C:
			sc.executeTasks(ctx)
		}
	}
}

func (sc *ScheduledCollector) executeTasks(ctx context.Context) {
	sc.mu.RLock()
	tasks := make([]CollectionTask, len(sc.taskQueue))
	copy(tasks, sc.taskQueue)
	sc.mu.RUnlock()

	if len(tasks) == 0 {
		logrus.Debug("No collection tasks to execute")
		return
	}

	logrus.Infof("Executing %d collection tasks", len(tasks))

	taskChan := make(chan CollectionTask, len(tasks))
	resultChan := make(chan error, len(tasks))

	for i := 0; i < sc.maxWorkers; i++ {
		go sc.worker(ctx, taskChan, resultChan)
	}

	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	for i := 0; i < len(tasks); i++ {
		if err := <-resultChan; err != nil {
			logrus.Warnf("Task execution failed: %v", err)
		}
	}
}

func (sc *ScheduledCollector) worker(ctx context.Context, tasks <-chan CollectionTask, results chan<- error) {
	for task := range tasks {
		var err error

		switch task.Type {
		case "post":
			_, err = sc.service.CollectPostMetrics(ctx, task.Platform, task.TargetID)
		case "account":
			_, err = sc.service.CollectAccountMetrics(ctx, task.Platform, task.TargetID)
		default:
			err = fmt.Errorf("unknown task type: %s", task.Type)
		}

		if err != nil {
			task.LastError = err.Error()
			task.Retries++
			if task.Retries < task.MaxRetries {
				sc.mu.Lock()
				sc.taskQueue = append(sc.taskQueue, task)
				sc.mu.Unlock()
			}
		}

		results <- err
	}
}

func (sc *ScheduledCollector) AddTask(task CollectionTask) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	task.CreatedAt = time.Now()
	if task.MaxRetries == 0 {
		task.MaxRetries = 3
	}

	sc.taskQueue = append(sc.taskQueue, task)
	logrus.Infof("Collection task added: %s - %s", task.Platform, task.Type)
}

func (sc *ScheduledCollector) RemoveTask(taskID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for i, task := range sc.taskQueue {
		if task.ID == taskID {
			sc.taskQueue = append(sc.taskQueue[:i], sc.taskQueue[i+1:]...)
			break
		}
	}
}

func (sc *ScheduledCollector) GetQueueLength() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return len(sc.taskQueue)
}

func (sc *ScheduledCollector) IsRunning() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.running
}
