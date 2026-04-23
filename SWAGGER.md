# Swagger Documentation

Dokumentasi API menggunakan Swagger/OpenAPI untuk SoundCave Backend.

## Akses Swagger UI

Setelah server berjalan, akses Swagger UI di:
- **URL**: `http://localhost:6002/swagger/index.html`

## Generate Swagger Docs

Untuk menggenerate ulang dokumentasi Swagger setelah menambahkan atau mengubah annotations:

```bash
go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o docs
```

Atau jika `swag` sudah terinstall secara global:

```bash
swag init -g main.go -o docs
```

## Menambahkan Swagger Annotations

### Contoh untuk Handler

```go
// HandlerName description
// @Summary      Short summary
// @Description  Detailed description
// @Tags         TagName
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Parameter description"
// @Param        request  body      RequestStruct  true  "Request body"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /endpoint [method]
func HandlerName(c *fiber.Ctx, db *gorm.DB) error {
    // Handler implementation
}
```

### Tags yang Tersedia

- `Auth` - Authentication endpoints
- `Users` - User management
- `Albums` - Album CRUD
- `Artists` - Artist CRUD
- `Genres` - Genre CRUD
- `Musics` - Music CRUD
- `MusicVideos` - Music Video CRUD
- `Podcasts` - Podcast CRUD
- `Playlists` - Playlist CRUD
- `Notifications` - Notification CRUD
- `News` - News CRUD
- `Images` - Image upload
- `Dashboard` - Dashboard endpoints
- `ArtistStreams` - Live streaming endpoints
- `SRS Webhooks` - SRS media server callbacks

### Security

Untuk endpoint yang memerlukan authentication, tambahkan:
```go
// @Security     BearerAuth
```

## Stream API Flow

Dokumentasi lengkap untuk live streaming dengan SRS (Simple RTMP Server).

### Status Stream

Stream memiliki 4 status:
- `pending` - Stream dibuat, menunggu publisher connect ke SRS
- `scheduled` - Stream dijadwalkan untuk waktu mendatang
- `live` - SRS confirm on_publish, stream sedang berlangsung
- `ended` - Stream berakhir (on_unpublish atau manual end)

### Flow Streaming

```
1. Artist POST /api/artist-streams/start
   ↓
2. Backend returns: streamKey, webrtc_url, ingest_url, playback_url
   (status = "pending")
   ↓
3. Client starts WebRTC push to: webrtc_url
   ↓
4. SRS POST /api/srs/on_publish (webhook)
   ↓
5. Backend updates status → "live", sets started_at
   ↓
6. Viewers play HLS from: playback_url
   ↓
7. Stream ends (publisher disconnects)
   ↓
8. SRS POST /api/srs/on_unpublish (webhook)
   ↓
9. Backend updates status → "ended", sets ended_at
```

### Environment Variables

Untuk konfigurasi streaming, set env vars berikut:

```bash
# SRS WHIP endpoint (WebRTC push)
SRS_WEBRTC_URL=http://localhost:1985/rtc/v1/whip

# HLS playback endpoint
HLS_SERVER_URL=http://localhost:8080/hls

# RTMP ingest (legacy, untuk kompatibilitas)
RTMP_SERVER_URL=rtmp://localhost/live

# Secret untuk validasi SRS callbacks (optional)
SRS_CALLBACK_SECRET=your-secret-key
```

### SRS Configuration

Konfigurasi SRS untuk memanggil webhook backend:

```json
{
  "vhost": "__defaultVhost__": {
    "http_hooks": {
      "enabled": true,
      "on_publish": "http://your-backend:6002/api/srs/on_publish?secret=your-secret-key",
      "on_unpublish": "http://your-backend:6002/api/srs/on_unpublish?secret=your-secret-key"
    }
  }
}
```

### Endpoints

#### POST /api/artist-streams/start
- **Status**: `pending` (untuk live) atau `scheduled` (jika scheduled_at di masa depan)
- **Returns**: streamKey, webrtc_url, ingest_url, playback_url
- **Auth**: Required (Bearer Token)

#### POST /api/srs/on_publish
- **Called by**: SRS media server
- **Body**: JSON dengan stream key
- **Updates**: status → live, started_at = now
- **Auth**: Optional secret validation
- **Response**: `{"code": 0, "msg": "ok"}`

#### POST /api/srs/on_unpublish
- **Called by**: SRS media server
- **Body**: JSON dengan stream key
- **Updates**: status → ended, ended_at = now
- **Auth**: Optional secret validation
- **Response**: `{"code": 0, "msg": "ok"}`

## File yang Ter-generate

Setelah menjalankan `swag init`, file berikut akan dibuat di folder `docs/`:

- `docs.go` - Generated Go code
- `swagger.json` - OpenAPI specification (JSON)
- `swagger.yaml` - OpenAPI specification (YAML)

## Catatan

- Pastikan untuk menjalankan `swag init` setiap kali menambahkan atau mengubah Swagger annotations
- File di folder `docs/` adalah auto-generated, jangan edit manual
- Swagger UI akan otomatis membaca dari file `swagger.json` atau `swagger.yaml`
- **Untuk stream API**: jalankan `swag init -g main.go -o docs` untuk generate dokumentasi terbaru dengan SRS webhook endpoints
