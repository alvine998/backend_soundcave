package handlers

import (
	"backend_soundcave/models"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/zishang520/socket.io/v2/socket"
	"gorm.io/gorm"
)

var (
	ioServer *socket.Server
	// Map to track which stream each socket is currently watching
	// socketID -> streamID
	viewerMap = sync.Map{}
)

// InitSocketServer initializes the Socket.IO server
func InitSocketServer(db *gorm.DB) *socket.Server {
	io := socket.NewServer(nil, nil)

	io.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		log.Printf("Socket connected: %s", client.Id())

		// Event: Join a stream room
		client.On("join_stream", func(args ...any) {
			if len(args) == 0 {
				return
			}

			// Expecting stream_id as the first argument
			streamIDStr := fmt.Sprintf("%v", args[0])
			streamID, err := strconv.Atoi(streamIDStr)
			if err != nil {
				log.Printf("Invalid stream ID: %v", args[0])
				return
			}

			roomName := fmt.Sprintf("stream:%d", streamID)
			client.Join(socket.Room(roomName))
			
			// Store current stream for cleanup on disconnect
			viewerMap.Store(client.Id(), streamID)

			log.Printf("User %s joined stream %d", client.Id(), streamID)

			// Increment viewer count in DB
			updateViewerCount(db, int32(streamID), 1)

			// Broadcast new count to all in this stream's room
			broadcastViewerCount(io, int32(streamID), db)
		})

		// Event: Leave a stream room
		client.On("leave_stream", func(args ...any) {
			if len(args) == 0 {
				return
			}

			streamIDStr := fmt.Sprintf("%v", args[0])
			streamID, err := strconv.Atoi(streamIDStr)
			if err != nil {
				return
			}

			roomName := fmt.Sprintf("stream:%d", streamID)
			client.Leave(socket.Room(roomName))
			viewerMap.Delete(client.Id())

			log.Printf("User %s left stream %d", client.Id(), streamID)

			// Decrement viewer count in DB
			updateViewerCount(db, int32(streamID), -1)

			// Broadcast new count
			broadcastViewerCount(io, int32(streamID), db)
		})

		// Cleanup on disconnect
		client.On("disconnecting", func(args ...any) {
			if streamIDVal, ok := viewerMap.Load(client.Id()); ok {
				streamID := streamIDVal.(int)
				log.Printf("User %s disconnecting from stream %d", client.Id(), streamID)

				// Decrement viewer count
				updateViewerCount(db, int32(streamID), -1)

				// Broadcast new count
				broadcastViewerCount(io, int32(streamID), db)
				
				viewerMap.Delete(client.Id())
			}
		})
	})

	ioServer = io
	return io
}

// updateViewerCount updates the database viewer count for a stream
func updateViewerCount(db *gorm.DB, streamID int32, delta int) {
	// Use GORM's Update with raw SQL for atomic increment/decrement
	// This prevents race conditions
	err := db.Model(&models.ArtistStream{}).Where("id = ?", streamID).
		UpdateColumn("viewer_count", gorm.Expr("viewer_count + ?", delta)).Error
	
	if err != nil {
		log.Printf("Error updating viewer count for stream %d: %v", streamID, err)
	}

	// Ensure viewer_count never goes below 0 (manual safety check)
	db.Model(&models.ArtistStream{}).
		Where("id = ? AND viewer_count < 0", streamID).
		Update("viewer_count", 0)
}

// broadcastViewerCount fetches the latest count from DB and broadcasts to everyone in the room
func broadcastViewerCount(io *socket.Server, streamID int32, db *gorm.DB) {
	var stream models.ArtistStream
	if err := db.Select("viewer_count").First(&stream, streamID).Error; err != nil {
		return
	}

	roomName := fmt.Sprintf("stream:%d", streamID)
	io.To(socket.Room(roomName)).Emit("viewer_count_update", map[string]any{
		"stream_id":    streamID,
		"viewer_count": stream.ViewerCount,
	})
}
