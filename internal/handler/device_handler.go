package handler

import (
	"database/sql"
	"device-service/internal/middleware"
	"device-service/internal/model"
	"device-service/internal/repository"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func RegisterDeviceRoutes(r *gin.RouterGroup, repo *repository.DeviceRepository) {
	r.POST("/devices", func(c *gin.Context) {
		var device model.Device

		if err := c.BindJSON(&device); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// TEMPORARY — replace with real JWT extraction later
		device.OwnerID = middleware.GetUserID(c)

		err := repo.CreateDevice(c.Request.Context(), &device)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, device)
	})
	r.GET("/devices", func(c *gin.Context) {
		// Parse query params
		category := c.Query("category")
		availableStr := c.Query("available")
		minPriceStr := c.Query("min_price")
		maxPriceStr := c.Query("max_price")
		sort := c.DefaultQuery("sort", "recent")
		limitStr := c.DefaultQuery("limit", "10")
		pageStr := c.DefaultQuery("page", "1")
		city := c.Query("city")
		region := c.Query("region")

		// 🔍 DEBUG:
		fmt.Printf("🔎 Incoming filters — category=%q, available=%q, min_price=%q, max_price=%q, city=%q, region=%q\n",
			category, availableStr, minPriceStr, maxPriceStr, city, region)
		// Convert types
		var available *bool
		if availableStr == "true" {
			v := true
			available = &v
		} else if availableStr == "false" {
			v := false
			available = &v
		}

		minPrice, _ := strconv.ParseFloat(minPriceStr, 64)
		maxPrice, _ := strconv.ParseFloat(maxPriceStr, 64)
		limit, _ := strconv.Atoi(limitStr)
		page, _ := strconv.Atoi(pageStr)

		filter := map[string]interface{}{
			"category":  category,
			"available": available,
			"min_price": minPrice,
			"max_price": maxPrice,
			"sort":      sort,
			"limit":     limit,
			"page":      page,
			"city":      city,
			"region":    region,
		}

		devices, err := repo.GetAllDevices(c.Request.Context(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, devices)
	})

	r.GET("/devices/:id", func(c *gin.Context) {
		id := c.Param("id")

		device, err := repo.GetDeviceByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
			return
		}

		c.JSON(http.StatusOK, device)
	})
	r.PUT("/devices/:id", func(c *gin.Context) {
		id := c.Param("id")
		userID := middleware.GetUserID(c)

		var device model.Device
		if err := c.ShouldBindJSON(&device); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		device.ID = id
		device.OwnerID = userID

		err := repo.UpdateDevice(c.Request.Context(), &device)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Not found or no permission"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Device updated successfully"})
	})
	r.DELETE("/devices/:id", func(c *gin.Context) {
		id := c.Param("id")
		userID := middleware.GetUserID(c)

		err := repo.DeleteDevice(c.Request.Context(), id, userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Not found or no permission"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Device deleted successfully"})
	})
	type AvailabilityUpdate struct {
		Available bool `json:"available"`
	}

	r.PATCH("/devices/:id/availability", func(c *gin.Context) {
		id := c.Param("id")
		userID := middleware.GetUserID(c)

		var input AvailabilityUpdate
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		err := repo.UpdateAvailability(c.Request.Context(), id, userID, input.Available)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusForbidden, gin.H{"error": "Not found or no permission"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Availability updated"})
	})

}
