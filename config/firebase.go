package config

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// global Firebase Storage client + bucket
var (
	StorageClient *storage.Client
	StorageBucket string
	svcEmail      string
	svcKey        []byte
)

// serviceAccount holds only the fields we need from the JSON key
type serviceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
}

// InitFirebase reads your service-account JSON, initializes the Storage client,
// and loads the bucket name from the FIREBASE_BUCKET_NAME env var.
func InitFirebase() {
	// 1) Read the JSON key
	credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	data, err := ioutil.ReadFile(credPath)
	if err != nil {
		log.Fatalf("⛔️ cannot read Firebase key file %s: %v", credPath, err)
	}

	// 2) Parse out the email and private key for signed URLs
	var sa serviceAccount
	if err := json.Unmarshal(data, &sa); err != nil {
		log.Fatalf("⛔️ invalid Firebase key format: %v", err)
	}
	svcEmail = sa.ClientEmail
	svcKey = []byte(sa.PrivateKey)

	// 3) Create a Storage client using the same JSON key
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credPath))
	if err != nil {
		log.Fatalf("⛔️ failed to initialize Firebase Storage client: %v", err)
	}
	StorageClient = client

	// 4) Read your bucket name
	StorageBucket = os.Getenv("FIREBASE_BUCKET_NAME")
	if StorageBucket == "" {
		log.Fatal("⛔️ FIREBASE_BUCKET_NAME env var must be set")
	}

	log.Printf("✅ Firebase Storage ready (bucket: %s)\n", StorageBucket)
}

// GenerateUploadURL returns:
//   - uploadURL: a signed PUT URL valid for 15 minutes
//   - publicURL: the public https URL to read the object once uploaded
func GenerateUploadURL(objectName string) (uploadURL, publicURL string, err error) {
	opts := &storage.SignedURLOptions{
		GoogleAccessID: svcEmail,
		PrivateKey:     svcKey,
		Method:         "PUT",
		Expires:        time.Now().Add(15 * time.Minute),
		ContentType:    "image/jpeg",
	}

	uploadURL, err = storage.SignedURL(StorageBucket, objectName, opts)
	if err != nil {
		return "", "", err
	}

	publicURL = "https://storage.googleapis.com/" + StorageBucket + "/" + objectName
	return uploadURL, publicURL, nil
}
