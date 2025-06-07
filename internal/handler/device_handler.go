// internal/handler/device_handler.go

package handler

import (
	"database/sql"
	"device-service/internal/middleware"
	"device-service/internal/model"
	"device-service/internal/repository"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

// RegisterDeviceRoutes регистрирует маршруты для CRUD операций над устройствами.
// Ожидается, что поле image_url передаётся уже готовым (публичным) URL из Firebase Storage.
func RegisterDeviceRoutes(r *gin.RouterGroup, repo *repository.DeviceRepository) {
	// POST /api/devices — создаёт новое устройство
	r.POST("/devices", func(c *gin.Context) {
		var device model.Device

		// 1) Считываем JSON из тела
		if err := c.BindJSON(&device); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// 2) Получаем owner_id из JWT
		userID, _ := middleware.GetUserID(c)
		device.OwnerID = userID

		// 3) Создаём устройство в БД (включая поле ImageURL)
		err := repo.CreateDevice(c.Request.Context(), &device)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 4) Возвращаем созданный объект (включая сгенерированные id, timestamps)
		c.JSON(http.StatusCreated, device)
	})

	// GET /api/devices — список устройств (с фильтрами)
	r.GET("/devices", func(c *gin.Context) {
		filter := model.ParseDeviceFilter(c)

		devices, err := repo.GetAllDevices(c.Request.Context(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, devices)
	})

	// GET /api/devices/:id — получить устройство по ID
	r.GET("/devices/:id", func(c *gin.Context) {
		id := c.Param("id")

		device, err := repo.GetDeviceByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
			return
		}
		c.JSON(http.StatusOK, device)
	})

	// PUT /api/devices/:id — обновить устройство (в том числе можно обновить image_url)
	r.PUT("/devices/:id", func(c *gin.Context) {
		id := c.Param("id")
		userID, _ := middleware.GetUserID(c)

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

	// DELETE /api/devices/:id — удалить устройство
	r.DELETE("/devices/:id", func(c *gin.Context) {
		id := c.Param("id")
		userID, _ := middleware.GetUserID(c)

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

	// PATCH /api/devices/:id/availability — обновить доступность
	type AvailabilityUpdate struct {
		Available bool `json:"available"`
	}
	r.PATCH("/devices/:id/availability", func(c *gin.Context) {
		id := c.Param("id")
		userID, _ := middleware.GetUserID(c)

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

	// GET /api/devices/:id/availability — получить доступность
	r.GET("/devices/:id/availability", func(c *gin.Context) {
		deviceID := c.Param("id")
		device, err := repo.GetDeviceByID(c.Request.Context(), deviceID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"available": device.Available})
	})
}
