package handler

import (
	"net/http"

	"checkopen/internal/model"
	"checkopen/internal/scraper"

	"github.com/gin-gonic/gin"
)

// StatusHandler handles restaurant status requests
type StatusHandler struct {
	scraper *scraper.WongnaiScraper
}

// NewStatusHandler creates a new status handler
func NewStatusHandler() *StatusHandler {
	return &StatusHandler{
		scraper: scraper.NewWongnaiScraper(),
	}
}

// GetStatus handles GET /api/v1/status/:restaurantId
func (h *StatusHandler) GetStatus(c *gin.Context) {
	restaurantID := c.Param("restaurantId")

	if restaurantID == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "bad_request",
			Message: "Restaurant ID is required",
		})
		return
	}

	status, err := h.scraper.GetRestaurantStatus(restaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// HealthCheck handles GET /health
func (h *StatusHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Restaurant Status Checker API is running",
	})
}
