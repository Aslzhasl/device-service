package handler

import (
	"database/sql"
	"device-service/internal/middleware"
	"device-service/internal/model"
	"device-service/internal/repository"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func RegisterDeviceRoutes(r *gin.RouterGroup, repo *repository.DeviceRepository) {
	r.POST("/devices", func(c *gin.Context) {
		var device model.Device

		if err := c.BindJSON(&device); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// TEMPORARY — replace with real JWT extraction later
		device.OwnerID, _ = middleware.GetUserID(c)

		err := repo.CreateDevice(c.Request.Context(), &device)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, device)
	})
	r.GET("/devices", func(c *gin.Context) {
		filter := model.ParseDeviceFilter(c)
		fmt.Printf("🔎 Filter: %+v\n", filter) // debug

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
	//
	////PHOTO
	//r.GET("/upload-url", func(c *gin.Context) {
	//	fileName := c.Query("fileName")
	//	if fileName == "" {
	//		c.JSON(400, gin.H{"error": "fileName query param required"})
	//		return
	//	}
	//	uploadURL, publicURL, err := config.GenerateUploadURL(fileName)
	//	if err != nil {
	//		c.JSON(500, gin.H{"error": err.Error()})
	//		return
	//	}
	//	c.JSON(200, gin.H{
	//		"uploadUrl": uploadURL,
	//		"publicUrl": publicURL,
	//	})
	//})

}
func RegisterFavoriteRoutes(r *gin.RouterGroup, favRepo *repository.FavoriteRepository) {
	// Add to favorites
	r.POST("/devices/:id/favorite", func(c *gin.Context) {
		// 1) Extract the userID and check it exists
		userID, ok := middleware.GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}
		deviceID := c.Param("id")

		// 2) Debug log so you see what’s happening
		log.Printf("➕ AddFavorite called by user=%s for device=%s", userID, deviceID)

		// 3) Insert into DB
		if err := favRepo.AddFavorite(c.Request.Context(), userID, deviceID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	r.GET("/devices/favorites", func(c *gin.Context) {
		userID, ok := middleware.GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		log.Printf("🔎 GetFavorites called by user=%s", userID)
		devices, err := favRepo.GetFavorites(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, devices)
	})

	// Remove from favorites
	r.DELETE("/devices/:id/favorite", func(c *gin.Context) {
		userID, _ := middleware.GetUserID(c)
		deviceID := c.Param("id")

		if err := favRepo.RemoveFavorite(c.Request.Context(), userID, deviceID); err != nil {
			if err.Error() == "not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "Favorite not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.Status(http.StatusNoContent)
	})

}
