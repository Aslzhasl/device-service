// internal/handler/upload_handler.go

package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	_ "mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"device-service/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterUploadHandler регистрирует маршрут POST /api/upload для загрузки файла.
// Ожидается multipart/form-data с полем "file".
func RegisterUploadHandler(r *gin.RouterGroup) {
	r.POST("/upload", func(c *gin.Context) {
		// 1) Получаем файл из формы
		fileHeader, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file field is required"})
			return
		}

		// 2) Открываем файл для чтения
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open uploaded file"})
			return
		}
		defer file.Close()

		// 3) Генерируем уникальное имя объекта в бакете: devices/{uuid}.{ext}
		ext := filepath.Ext(fileHeader.Filename) // ".jpg", ".png" и т.д.
		if ext == "" {
			ext = ".jpg"
		}
		objectName := fmt.Sprintf("devices/%s%s", uuid.New().String(), ext)

		// 4) Пишем файл в Firebase Storage
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()

		wc := config.StorageClient.
			Bucket(config.StorageBucket).
			Object(objectName).
			NewWriter(ctx)

		// Устанавливаем правильный Content-Type
		wc.ContentType = fileHeader.Header.Get("Content-Type")
		if wc.ContentType == "" {
			wc.ContentType = "application/octet-stream"
		}

		if _, err := io.Copy(wc, file); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write file to Firebase Storage"})
			return
		}

		// Поправленный блок, чтобы логировать полную ошибку:
		if err := wc.Close(); err != nil {
			// распечатаем точную причину в консоль, чтобы понять, что не так
			log.Printf("🔥 Firebase upload close error: %v\n", err)

			// вернём ответ клиенту с детальным текстом (для отладки)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Закрываем writer (чтобы файл был окончательно сохранён)
		if err := wc.Close(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to finalize file upload"})
			return
		}

		// 5) Собираем публичный URL: https://storage.googleapis.com/{bucket}/{objectName}
		publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", config.StorageBucket, objectName)

		// 6) Возвращаем JSON с информацией о загруженном файле
		c.JSON(http.StatusOK, gin.H{
			"fileName":  objectName,
			"publicUrl": publicURL,
		})
	})
}
