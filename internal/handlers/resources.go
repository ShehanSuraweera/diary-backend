package handlers

import (
	"fmt"
	"net/http"

	"diary-backend/internal/database"
	"diary-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ResourceRequest struct {
	Title         string   `json:"title" binding:"required"`
	URL           string   `json:"url,omitempty"`
	Description   string   `json:"description,omitempty"`
	Technology    string   `json:"technology" binding:"required"`
	Type          string   `json:"type" binding:"required"`
	Status        string   `json:"status,omitempty"`
	Priority      string   `json:"priority,omitempty"`
	Rating        *int     `json:"rating,omitempty"`
	EstimatedTime *int     `json:"estimated_time,omitempty"`
	Progress      *int     `json:"progress,omitempty"`
	Notes         string   `json:"notes,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

func GetResources(c *gin.Context) {
	// Query parameters for filtering
	technology := c.Query("technology")
	resourceType := c.Query("type")
	status := c.Query("status")
	priority := c.Query("priority")
	search := c.Query("search")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")

	// Convert page and limit to integers
	pageInt := 1
	limitInt := 20
	fmt.Sscanf(page, "%d", &pageInt)
	fmt.Sscanf(limit, "%d", &limitInt)
	if pageInt < 1 {
		pageInt = 1
	}
	if limitInt < 1 {
		limitInt = 20
	}

	// Build query
	db := database.GetDB()
	query := db.Model(&models.Resource{})
	if technology != "" {
		query = query.Where("technology = ?", technology)
	}
	if resourceType != "" {
		query = query.Where("type = ?", resourceType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Pagination
	offset := (pageInt - 1) * limitInt
	var resources []models.Resource
	err := query.Offset(offset).Limit(limitInt).Find(&resources).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch resources"})
		return
	}

	totalPages := (total + int64(limitInt) - 1) / int64(limitInt)

	c.JSON(http.StatusOK, gin.H{
		"resources": resources,
		"pagination": gin.H{
			"page":        pageInt,
			"limit":       limitInt,
			"total":       total,
			"total_pages": totalPages,
		},
		"filters": gin.H{
			"technology": technology,
			"type":       resourceType,
			"status":     status,
			"priority":   priority,
			"search":     search,
		},
	})
}

func GetResourceByID(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID
	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}

	// Query the database for the resource
	db := database.GetDB()
	var resource models.Resource
	err = db.Where("id = ?", uid).First(&resource).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resource": resource})
}

func CreateResource(c *gin.Context) {
	var req ResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a new UUID for the resource
	resourceID := uuid.New()

	// Build the resource model
	resource := models.Resource{
		ID:            resourceID,
		Title:         req.Title,
		URL:           req.URL,
		Description:   req.Description,
		Technology:    req.Technology,
		Type:          req.Type,
		Status:        req.Status,
		Priority:      req.Priority,
		Rating:        req.Rating,
		EstimatedTime: req.EstimatedTime,
		Progress:      req.Progress,
		Notes:         req.Notes,
		Tags:          req.Tags,
	}

	// Insert into database
	err := database.CreateResource(c.Request.Context(), &resource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create resource"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Resource created successfully",
		"resource": resource,
	})
}

func UpdateResource(c *gin.Context) {
	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}

	var req ResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var resource models.Resource
	// Find the resource first
	err = db.Where("id = ?", uid).First(&resource).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
		return
	}

	// Update fields
	resource.Title = req.Title
	resource.URL = req.URL
	resource.Description = req.Description
	resource.Technology = req.Technology
	resource.Type = req.Type
	resource.Status = req.Status
	resource.Priority = req.Priority
	resource.Rating = req.Rating
	resource.EstimatedTime = req.EstimatedTime
	resource.Progress = req.Progress
	resource.Notes = req.Notes
	resource.Tags = req.Tags

	// Save the updated resource
	err = db.Save(&resource).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update resource"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Resource updated successfully",
		"resource": resource,
	})
}

func DeleteResource(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}

	// TODO: Implement database deletion
	c.JSON(http.StatusOK, gin.H{
		"message": "Resource deleted successfully",
	})
}

func UpdateResourceStatus(c *gin.Context) {
	id := c.Param("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var resource models.Resource
	err = db.Where("id = ?", uid).First(&resource).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
		return
	}

	resource.Status = req.Status
	err = db.Save(&resource).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update resource status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Resource status updated successfully",
		"resource": resource,
	})
}

func UpdateResourceRating(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid resource ID"})
		return
	}

	var req struct {
		Rating int `json:"rating" binding:"required,min=1,max=5"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement rating update
	c.JSON(http.StatusOK, gin.H{
		"message": "Resource rating updated successfully",
	})
}

func GetResourceStats(c *gin.Context) {
	// TODO: Implement stats calculation
	c.JSON(http.StatusOK, gin.H{
		"total_resources":      0,
		"completed_count":      0,
		"in_progress_count":    0,
		"to_read_count":        0,
		"bookmarked_count":     0,
		"weekly_hours":         0,
		"avg_rating":           0,
		"technology_breakdown": map[string]int{},
		"type_breakdown":       map[string]int{},
	})
}

func GetTechnologies(c *gin.Context) {
	// TODO: Get unique technologies from database
	c.JSON(http.StatusOK, gin.H{
		"technologies": []string{},
	})
}

func ImportFromURL(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required,url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement URL metadata extraction
	c.JSON(http.StatusOK, gin.H{
		"message": "Resource imported successfully",
		"resource": gin.H{
			"title":       "Extracted Title",
			"description": "Extracted Description",
			"type":        "article",
		},
	})
}
