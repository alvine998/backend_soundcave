package config

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	firebaseStorage "firebase.google.com/go/v4/storage"
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
	if StorageClient == nil {
		return nil, fmt.Errorf("Firebase Storage client belum diinisialisasi")
	}

	bucketName := os.Getenv("FIREBASE_STORAGE_BUCKET")
	if bucketName == "" {
		return nil, fmt.Errorf("FIREBASE_STORAGE_BUCKET environment variable tidak di-set")
	}

	bucket, err := StorageClient.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}
