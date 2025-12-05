# SoundCave Backend

Backend API menggunakan Golang Fiber dengan MySQL dan Firebase Storage untuk upload gambar.

## Fitur

- ✅ RESTful API dengan Golang Fiber
- ✅ Database MySQL dengan GORM
- ✅ Upload gambar ke Firebase Storage
- ✅ Upload multiple gambar
- ✅ Validasi file (size, type)
- ✅ CRUD operations untuk gambar

## Setup

### 1. Install Dependencies

```bash
go mod download
```

### 2. Setup Database MySQL

Buat database MySQL:

```sql
CREATE DATABASE soundcave;
```

### 3. Setup Firebase

1. Buat project di [Firebase Console](https://console.firebase.google.com/)
2. Enable Firebase Storage
3. Download service account key JSON
4. Simpan sebagai `firebase-service-account.json` di root project
5. Set bucket name di `.env`

### 4. Setup Environment Variables

Copy `.env.example` ke `.env` dan isi dengan konfigurasi Anda:

```bash
cp .env.example .env
```

Edit `.env`:
```
PORT=3000
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=soundcave
FIREBASE_SERVICE_ACCOUNT_KEY=./firebase-service-account.json
FIREBASE_STORAGE_BUCKET=your-project-id.appspot.com
```

### 5. Run Application

```bash
go run main.go
```

Server akan berjalan di `http://localhost:3000`

## API Endpoints

### Health Check
- `GET /` - Health check

### Image Upload
- `POST /api/images/upload` - Upload single image
  - Body: `multipart/form-data` dengan field `image`
  
- `POST /api/images/upload-multiple` - Upload multiple images
  - Body: `multipart/form-data` dengan field `images[]`

- `GET /api/images` - Get all images

- `DELETE /api/images/:id` - Delete image by ID

## Contoh Request

### Upload Single Image

```bash
curl -X POST http://localhost:3000/api/images/upload \
  -F "image=@/path/to/image.jpg"
```

### Upload Multiple Images

```bash
curl -X POST http://localhost:3000/api/images/upload-multiple \
  -F "images[]=@/path/to/image1.jpg" \
  -F "images[]=@/path/to/image2.png"
```

## Struktur Project

```
backend_soundcave/
├── config/
│   └── config.go          # Firebase configuration
├── database/
│   └── database.go         # MySQL connection
├── handlers/
│   └── upload_handler.go   # Image upload handlers
├── models/
│   └── image.go           # Image model
├── routes/
│   └── routes.go          # API routes
├── .env                   # Environment variables
├── .env.example           # Example env file
├── go.mod                 # Go dependencies
├── go.sum                 # Go dependencies checksum
└── main.go               # Application entry point
```

