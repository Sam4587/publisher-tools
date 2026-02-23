package api

import (
	"encoding/json"
	"net/http"

	"publisher-core/task"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// SchedulerHandlers 调度器处理器
type SchedulerHandlers struct {
	scheduler *task.SchedulerService
}

// NewSchedulerHandlers 创建调度器处理器
func NewSchedulerHandlers(scheduler *task.SchedulerService) *SchedulerHandlers {
	return &SchedulerHandlers{
		scheduler: scheduler,
	}
}

// RegisterRoutes 注册路由
func (h *SchedulerHandlers) RegisterRoutes(router *mux.Router) {
	schedulerRouter := router.PathPrefix("/api/v1/scheduler").Subrouter()

	// 定时任务管理
	schedulerRouter.HandleFunc("/tasks", h.listScheduledTasks).Methods("GET")
	schedulerRouter.HandleFunc("/tasks", h.createScheduledTask).Methods("POST")
	schedulerRouter.HandleFunc("/tasks/{name}", h.getScheduledTask).Methods("GET")
	schedulerRouter.HandleFunc("/tasks/{name}", h.updateScheduledTask).Methods("PUT")
	schedulerRouter.HandleFunc("/tasks/{name}", h.deleteScheduledTask).Methods("DELETE")
	schedulerRouter.HandleFunc("/tasks/{name}/pause", h.pauseScheduledTask).Methods("POST")
	schedulerRouter.HandleFunc("/tasks/{name}/resume", h.resumeScheduledTask).Methods("POST")
	schedulerRouter.HandleFunc("/tasks/{name}/run", h.runScheduledTaskNow).Methods("POST")

	// 调度器状态
	schedulerRouter.HandleFunc("/stats", h.getSchedulerStats).Methods("GET")
}

// createScheduledTask 创建定时任务
func (h *SchedulerHandlers) createScheduledTask(w http.ResponseWriter, r *http.Request) {
	var req task.ScheduledTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	task, err := h.scheduler.CreateScheduledTask(&req)
	if err != nil {
		jsonError(w, "CREATE_TASK_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, task)
}

// listScheduledTasks 列出定时任务
func (h *SchedulerHandlers) listScheduledTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.scheduler.ListScheduledTasks()
	if err != nil {
		jsonError(w, "LIST_TASKS_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"tasks": tasks,
		"total": len(tasks),
	})
}

// getScheduledTask 获取定时任务
func (h *SchedulerHandlers) getScheduledTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	task, err := h.scheduler.GetScheduledTask(name)
	if err != nil {
		jsonError(w, "TASK_NOT_FOUND", err.Error(), http.StatusNotFound)
		return
	}

	jsonSuccess(w, task)
}

// updateScheduledTask 更新定时任务
func (h *SchedulerHandlers) updateScheduledTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	var req task.ScheduledTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.scheduler.UpdateScheduledTask(name, &req); err != nil {
		jsonError(w, "UPDATE_TASK_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	task, err := h.scheduler.GetScheduledTask(name)
	if err != nil {
		jsonError(w, "TASK_NOT_FOUND", err.Error(), http.StatusNotFound)
		return
	}

	jsonSuccess(w, task)
}

// deleteScheduledTask 删除定时任务
func (h *SchedulerHandlers) deleteScheduledTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if err := h.scheduler.DeleteScheduledTask(name); err != nil {
		jsonError(w, "DELETE_TASK_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"message": "Scheduled task deleted successfully",
		"name":    name,
	})
}

// pauseScheduledTask 暂停定时任务
func (h *SchedulerHandlers) pauseScheduledTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if err := h.scheduler.PauseScheduledTask(name); err != nil {
		jsonError(w, "PAUSE_TASK_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"message": "Scheduled task paused successfully",
		"name":    name,
	})
}

// resumeScheduledTask 恢复定时任务
func (h *SchedulerHandlers) resumeScheduledTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if err := h.scheduler.ResumeScheduledTask(name); err != nil {
		jsonError(w, "RESUME_TASK_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"message": "Scheduled task resumed successfully",
		"name":    name,
	})
}

// runScheduledTaskNow 立即执行定时任务
func (h *SchedulerHandlers) runScheduledTaskNow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if err := h.scheduler.RunScheduledTaskNow(name); err != nil {
		jsonError(w, "RUN_TASK_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, map[string]interface{}{
		"message": "Scheduled task triggered successfully",
		"name":    name,
	})
}

// getSchedulerStats 获取调度器统计
func (h *SchedulerHandlers) getSchedulerStats(w http.ResponseWriter, r *http.Request) {
	stats := h.scheduler.GetSchedulerStats()
	jsonSuccess(w, stats)
}
