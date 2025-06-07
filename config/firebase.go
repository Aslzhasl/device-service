// config/firebase.go

package config

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// StorageClient — глобальный клиент для Firebase Storage
var StorageClient *storage.Client

// StorageBucket — имя вашего GCS-бакета (“diploma-32c26.appspot.com”)
var StorageBucket string

// serviceAccount держит только те поля, которые нужны для signed URL (record-keeping)
type serviceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
}

// InitFirebase инициализирует клиент Storage и читает имя бакета из окружения
func InitFirebase() {
	// 1) Читаем JSON-ключ
	credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	data, err := ioutil.ReadFile(credPath)
	if err != nil {
		log.Fatalf("⛔️ cannot read Firebase key file %s: %v", credPath, err)
	}

	// 2) Распарсим email/ключ (не обязательно использовать их в дальнейшем напрямую)
	var sa serviceAccount
	if err := json.Unmarshal(data, &sa); err != nil {
		log.Fatalf("⛔️ invalid Firebase key format: %v", err)
	}

	// 3) Создаём клиента Storage
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credPath))
	if err != nil {
		log.Fatalf("⛔️ failed to initialize Firebase Storage client: %v", err)
	}
	StorageClient = client

	// 4) Получаем имя бакета
	StorageBucket = os.Getenv("FIREBASE_BUCKET_NAME")
	if StorageBucket == "" {
		log.Fatal("⛔️ FIREBASE_BUCKET_NAME env var must be set")
	}

	log.Printf("✅ Firebase Storage ready (bucket: %s)\n", StorageBucket)
}
