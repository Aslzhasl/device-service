package handler

import (
	"device-service/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterMetaRoutes(r *gin.RouterGroup, repo *repository.DeviceRepository) {
	// GET /api/categories
	r.GET("/categories", func(c *gin.Context) {
		cats, err := repo.GetCategories(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, cats)
	})

	// GET /api/cities
	r.GET("/cities", func(c *gin.Context) {
		cities, err := repo.GetCities(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, cities)
	})

	// GET /api/regions
	r.GET("/regions", func(c *gin.Context) {
		regions, err := repo.GetRegions(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, regions)
	})

	// GET /api/devices/trending?limit=10
	r.GET("/devices/trending", func(c *gin.Context) {
		// optional limit param
		limitStr := c.DefaultQuery("limit", "10")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 10
		}
		devices, err := repo.GetTrendingDevices(c.Request.Context(), limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, devices)
	})
}
