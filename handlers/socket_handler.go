package handlers

import (
	"backend_soundcave/models"
	"fmt"
	stdio "io"
	"log"
	"os"
	"os/exec"
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

	// Map to track active ffmpeg stream stdin pipes for web broadcasting
	// streamKey -> io.WriteCloser (stdin of ffmpeg process)
	ffmpegPipes = sync.Map{}
	// socketID -> streamKey (to clean up if the broadcaster disconnects abruptly)
	broadcasterMap = sync.Map{}
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

		// Event: Send a live comment
		client.On("send_comment", func(args ...any) {
			if len(args) == 0 {
				return
			}

			// Expecting a map with stream_id, username, message, and optionally profile_image
			data, ok := args[0].(map[string]interface{})
			if !ok {
				return
			}

			streamIDStr := fmt.Sprintf("%v", data["stream_id"])
			streamID, err := strconv.Atoi(streamIDStr)
			if err != nil {
				return
			}

			username, _ := data["username"].(string)
			message, _ := data["message"].(string)
			profileImage, _ := data["profile_image"].(string)

			if message == "" {
				return
			}

			roomName := fmt.Sprintf("stream:%d", streamID)

			// Broadcast to everyone in the room (including sender)
			io.To(socket.Room(roomName)).Emit("new_comment", map[string]any{
				"stream_id":     streamID,
				"username":      username,
				"message":       message,
				"profile_image": profileImage,
			})

			log.Printf("Comment in stream %d by %s: %s", streamID, username, message)
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

			// Clean up broadcaster ffmpeg pipe if they disconnect
			if streamKeyVal, ok := broadcasterMap.Load(client.Id()); ok {
				streamKey := streamKeyVal.(string)
				if pipeVal, exists := ffmpegPipes.Load(streamKey); exists {
					if stdin, ok := pipeVal.(stdio.WriteCloser); ok {
						log.Printf("Closing ffmpeg pipe for disconnected broadcaster %s", streamKey)
						stdin.Close()
					}
					ffmpegPipes.Delete(streamKey)
				}
				broadcasterMap.Delete(client.Id())
			}
		})

		// Event: Start Web Broadcast
		client.On("start_web_broadcast", func(args ...any) {
			if len(args) == 0 {
				return
			}
			data, ok := args[0].(map[string]interface{})
			if !ok {
				return
			}
			streamKey, _ := data["streamKey"].(string)
			if streamKey == "" {
				return
			}

			rtmpBaseURL := os.Getenv("RTMP_SERVER_URL")
			if rtmpBaseURL == "" {
				rtmpBaseURL = "rtmp://localhost/live"
			}
			destURL := fmt.Sprintf("%s/%s", rtmpBaseURL, streamKey)

			cmd := exec.Command("ffmpeg",
                "-f", "webm",    // Tell ffmpeg the incoming pipe is WebM format
				"-i", "pipe:0", // Read from stdin
				"-c:v", "libx264", // Re-encode video to H.264
				"-preset", "veryfast", // Fast encoding for live streams
				"-tune", "zerolatency",
				"-b:v", "2000k",
				"-c:a", "aac", // Re-encode audio to AAC expected by RTMP
				"-ar", "44100",
				"-b:a", "128k",
				"-f", "flv", // Output to FLV container for RTMP
				destURL,
			)

			stdin, err := cmd.StdinPipe()
			if err != nil {
				log.Printf("Failed to get ffmpeg stdin pipe: %v", err)
				client.Emit("web_broadcast_error", map[string]string{"message": "Failed to start ffmpeg pipe"})
				return
			}

			// Capture ffmpeg output for debugging
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Start(); err != nil {
				log.Printf("Failed to start ffmpeg: %v", err)
				client.Emit("web_broadcast_error", map[string]string{"message": "Failed to start ffmpeg process"})
				return
			}

			log.Printf("Started ffmpeg processing for stream %s to %s", streamKey, destURL)
			ffmpegPipes.Store(streamKey, stdin)
			broadcasterMap.Store(client.Id(), streamKey)

			// Notify client we're ready
			client.Emit("web_broadcast_ready", map[string]string{"streamKey": streamKey})

			// Goroutine to wait for ffmpeg to exit
			go func() {
				err := cmd.Wait()
				log.Printf("ffmpeg for stream %s exited with error: %v", streamKey, err)
				ffmpegPipes.Delete(streamKey)
			}()
		})

		// Event: Receive Video Chunk
		client.On("web_broadcast_chunk", func(args ...any) {
			if len(args) == 0 {
				return
			}
			data, ok := args[0].(map[string]interface{})
			if !ok {
				return
			}
			streamKey, _ := data["streamKey"].(string)
			chunk, _ := data["chunk"].([]byte)

			if streamKey == "" || len(chunk) == 0 {
				return
			}

			if pipeVal, exists := ffmpegPipes.Load(streamKey); exists {
				if stdin, ok := pipeVal.(stdio.WriteCloser); ok {
					_, err := stdin.Write(chunk)
					if err != nil {
						log.Printf("Error writing chunk to ffmpeg for %s: %v", streamKey, err)
					}
				}
			}
		})

		// Event: Stop Web Broadcast
		client.On("stop_web_broadcast", func(args ...any) {
			if len(args) == 0 {
				return
			}
			data, ok := args[0].(map[string]interface{})
			if !ok {
				return
			}
			streamKey, _ := data["streamKey"].(string)
			if streamKey == "" {
				return
			}

			if pipeVal, exists := ffmpegPipes.Load(streamKey); exists {
				if stdin, ok := pipeVal.(stdio.WriteCloser); ok {
					log.Printf("Stopping web broadcast, closing pipe for %s", streamKey)
					stdin.Close()
				}
				ffmpegPipes.Delete(streamKey)
			}
			broadcasterMap.Delete(client.Id())
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
