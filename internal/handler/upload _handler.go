package handler

import (
	"device-service/config"
	"github.com/gin-gonic/gin"
	"net/http"
)

// RegisterUploadURLRoute wires up GET /upload-url
func RegisterUploadURLRoute(r *gin.RouterGroup) {
	r.GET("/upload-url", func(c *gin.Context) {
		fileName := c.Query("fileName")
		if fileName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "fileName query param required"})
			return
		}

		uploadURL, publicURL, err := config.GenerateUploadURL(fileName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"uploadUrl": uploadURL,
			"publicUrl": publicURL,
		})
	})
}
