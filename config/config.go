package config

import (
	"context"
	"os"

	firebase "firebase.google.com/go/v4"
	firebaseStorage "firebase.google.com/go/v4/storage"
	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var (
	StorageClient *firebaseStorage.Client
)

// InitFirebase menginisialisasi Firebase App dan Storage Client
func InitFirebase() (*firebase.App, error) {
	ctx := context.Background()

	// Path ke Firebase service account key
	serviceAccountKey := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY")
	if serviceAccountKey == "" {
		serviceAccountKey = "./firebase-service-account.json"
	}

	// Inisialisasi Firebase App
	opt := option.WithCredentialsFile(serviceAccountKey)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	// Inisialisasi Storage Client
	StorageClient, err = app.Storage(ctx)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// GetStorageBucket mengembalikan storage bucket
func GetStorageBucket() (*storage.BucketHandle, error) {
	bucketName := os.Getenv("FIREBASE_STORAGE_BUCKET")
	if bucketName == "" {
		return nil, nil
	}

	bucket, err := StorageClient.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

