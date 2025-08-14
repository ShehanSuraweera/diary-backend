package handlers

import (
	"diary-backend/internal/database"
	"diary-backend/internal/models"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Add this helper function at the top of your handlers file
func getColumnName(sortBy string) string {
	columnMap := map[string]string{
		"dueDate":   "due_date",
		"createdAt": "created_at",
		"updatedAt": "updated_at",
		"projectId": "project_id",
		"title":     "title",
		"priority":  "priority",
		"category":  "category",
		"status":    "status",
		"completed": "completed",
	}

	if column, exists := columnMap[sortBy]; exists {
		return column
	}
	return sortBy // fallback to original if not found
}

// GetTasks retrieves tasks with optional filtering
func GetTasks(c *gin.Context) {
	var filters models.TaskFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	// Set default values
	if filters.Limit == 0 {
		filters.Limit = 50
	}
	if filters.SortBy == "" {
		filters.SortBy = "dueDate"
	}
	if filters.SortOrder == "" {
		filters.SortOrder = "asc"
	}

	db := database.GetDB()
	query := db.Model(&models.Task{})

	// Apply filters
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}

	if filters.Priority != "" {
		query = query.Where("priority = ?", filters.Priority)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.Completed != nil {
		query = query.Where("completed = ?", *filters.Completed)
	}

	if filters.ProjectID != "" {
		if projectUUID, err := uuid.Parse(filters.ProjectID); err == nil {
			query = query.Where("project_id = ?", projectUUID)
		}
	}

	// Search in title and description
	if filters.Search != "" {
		searchTerm := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	// Date filtering
	now := time.Now()
	switch filters.DateFilter {
	case "today":
		today := now.Format("2006-01-02")
		query = query.Where("due_date = ?", today)
	case "tomorrow":
		tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
		query = query.Where("due_date = ?", tomorrow)
	case "this-week":
		weekEnd := now.AddDate(0, 0, 7-int(now.Weekday()))
		query = query.Where("due_date BETWEEN ? AND ?", now.Format("2006-01-02"), weekEnd.Format("2006-01-02"))
	case "overdue":
		yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
		query = query.Where("due_date <= ? AND completed = false", yesterday)
	}

	// Tag filtering
	if filters.Tags != "" {
		tags := strings.Split(filters.Tags, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				query = query.Where("? = ANY(tags)", tag)
			}
		}
	}

	sortColumn := getColumnName(filters.SortBy) // Convert camelCase to snake_case
	orderClause := fmt.Sprintf("%s %s", sortColumn, filters.SortOrder)

	if filters.SortBy == "priority" {
		// Custom priority ordering
		orderClause = "CASE priority WHEN 'urgent' THEN 4 WHEN 'high' THEN 3 WHEN 'medium' THEN 2 WHEN 'low' THEN 1 END"
		if filters.SortOrder == "asc" {
			orderClause += " ASC"
		} else {
			orderClause += " DESC"
		}
	}

	var tasks []models.Task
	var total int64

	// Get total count
	query.Count(&total)

	// Apply pagination and sorting
	result := query.Order(orderClause).Limit(filters.Limit).Offset(filters.Offset).Find(&tasks)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"pagination": gin.H{
			"total":  total,
			"limit":  filters.Limit,
			"offset": filters.Offset,
		},
	})
}

// GetTaskByID retrieves a single task by ID
func GetTaskByID(c *gin.Context) {
	taskID := c.Param("id")

	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	var task models.Task
	db := database.GetDB()

	result := db.First(&task, "id = ?", taskUUID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// CreateTask creates a new task
func CreateTask(c *gin.Context) {
	var req models.CreateTaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create task model
	task := models.Task{
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		Priority:    req.Priority,
		Category:    req.Category,
		Status:      req.Status,
		ProjectID:   req.ProjectID,
		Tags:        pq.StringArray(req.Tags),
		Completed:   false,
	}

	// Set defaults if empty
	if task.Priority == "" {
		task.Priority = "medium"
	}
	if task.Category == "" {
		task.Category = "personal"
	}
	if task.Status == "" {
		task.Status = "pending"
	}

	db := database.GetDB()
	result := db.Create(&task)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"task": task})
}

// UpdateTask updates an existing task
func UpdateTask(c *gin.Context) {
	taskID := c.Param("id")

	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var task models.Task

	// Check if task exists
	if result := db.First(&task, "id = ?", taskUUID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Update fields only if provided
	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Completed != nil {
		updates["completed"] = *req.Completed
		// Auto-update status when marked as completed
		if *req.Completed && task.Status != "completed" {
			updates["status"] = "completed"
		}
	}
	if req.DueDate != nil {
		updates["due_date"] = *req.DueDate
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.ProjectID != nil {
		updates["project_id"] = *req.ProjectID
	}
	if req.Tags != nil {
		updates["tags"] = pq.StringArray(*req.Tags)
	}

	// Update timestamp
	updates["updated_at"] = time.Now()

	result := db.Model(&task).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	// Fetch updated task
	db.First(&task, "id = ?", taskUUID)
	c.JSON(http.StatusOK, gin.H{"task": task})
}

// DeleteTask deletes a task by ID
func DeleteTask(c *gin.Context) {
	taskID := c.Param("id")

	taskUUID, err := uuid.Parse(taskID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	db := database.GetDB()
	result := db.Delete(&models.Task{}, "id = ?", taskUUID)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// GetTaskStats returns task statistics
func GetTaskStats(c *gin.Context) {
	db := database.GetDB()

	var stats struct {
		Total          int64 `json:"total"`
		Today          int64 `json:"today"`
		TodayCompleted int64 `json:"todayCompleted"`
		InProgress     int64 `json:"inProgress"`
		Completed      int64 `json:"completed"`
		Overdue        int64 `json:"overdue"`
	}

	// Total tasks
	db.Model(&models.Task{}).Count(&stats.Total)

	// Today's tasks
	today := time.Now().Format("2006-01-02")
	db.Model(&models.Task{}).Where("due_date = ?", today).Count(&stats.Today)

	// Today's completed tasks
	db.Model(&models.Task{}).Where("due_date = ? AND completed = true", today).Count(&stats.TodayCompleted)

	// In progress tasks
	db.Model(&models.Task{}).Where("completed = false AND status != 'cancelled'").Count(&stats.InProgress)

	// Completed tasks
	db.Model(&models.Task{}).Where("completed = true").Count(&stats.Completed)

	// Overdue tasks
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	db.Model(&models.Task{}).Where("due_date <= ? AND completed = false", yesterday).Count(&stats.Overdue)

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}
