// internal/handler/favorite_handler.go

package handler

import (
	"log"
	"net/http"

	"device-service/internal/middleware"
	"device-service/internal/repository"

	"github.com/gin-gonic/gin"
)

// RegisterFavoriteRoutes регистрирует маршруты для работы с избранным (favorites).
func RegisterFavoriteRoutes(r *gin.RouterGroup, favRepo *repository.FavoriteRepository) {
	// 1) Добавить устройство в избранное: POST /api/devices/:id/favorite
	r.POST("/devices/:id/favorite", func(c *gin.Context) {
		// Извлекаем userID из контекста (JWT)
		userID, ok := middleware.GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}
		deviceID := c.Param("id")

		log.Printf("➕ AddFavorite called by user=%s for device=%s", userID, deviceID)

		if err := favRepo.AddFavorite(c.Request.Context(), userID, deviceID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Успешно добавлено, возвращаем 204 No Content
		c.Status(http.StatusNoContent)
	})

	// 2) Получить список избранных устройств пользователя: GET /api/devices/favorite
	r.GET("/devices/favorite", func(c *gin.Context) {
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

	// 3) Удалить устройство из избранного: DELETE /api/devices/:id/favorite
	r.DELETE("/devices/:id/favorite", func(c *gin.Context) {
		userID, _ := middleware.GetUserID(c) // Если токен валиден, userID будет ненулевым
		deviceID := c.Param("id")

		if err := favRepo.RemoveFavorite(c.Request.Context(), userID, deviceID); err != nil {
			// Если ошибки “not found” — возвращаем 404, иначе 500
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
