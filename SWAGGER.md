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

### Security

Untuk endpoint yang memerlukan authentication, tambahkan:
```go
// @Security     BearerAuth
```

## File yang Ter-generate

Setelah menjalankan `swag init`, file berikut akan dibuat di folder `docs/`:

- `docs.go` - Generated Go code
- `swagger.json` - OpenAPI specification (JSON)
- `swagger.yaml` - OpenAPI specification (YAML)

## Catatan

- Pastikan untuk menjalankan `swag init` setiap kali menambahkan atau mengubah Swagger annotations
- File di folder `docs/` adalah auto-generated, jangan edit manual
- Swagger UI akan otomatis membaca dari file `swagger.json` atau `swagger.yaml`
