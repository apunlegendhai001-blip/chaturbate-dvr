package router

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/teacat/chaturbate-dvr/entity"
	"github.com/teacat/chaturbate-dvr/server"
)

// IndexData represents the data structure for the index page.
type IndexData struct {
	Config     *entity.Config
	Channels   []*entity.ChannelInfo
	DiskUsage  string
}

// getDiskUsage calculates the total size of files in the given directory.
func getDiskUsage(dir string) string {
	var totalSize int64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	// Format size
	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := float64(totalSize)
	unitIndex := 0
	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}
	return fmt.Sprintf("%.1f %s", size, units[unitIndex])
}

// Index renders the index page with channel information.
func Index(c *gin.Context) {
	dir := "videos"
	if server.Config.OutputDir != "" {
		dir = server.Config.OutputDir
	}
	diskUsage := getDiskUsage(dir)
	c.HTML(200, "index.html", &IndexData{
		Config:    server.Config,
		Channels:  server.Manager.ChannelInfo(),
		DiskUsage: diskUsage,
	})
}

// CreateChannelRequest represents the request body for creating a channel.
type CreateChannelRequest struct {
	Username    string `form:"username" binding:"required"`
	Framerate   int    `form:"framerate" binding:"required"`
	Resolution  int    `form:"resolution" binding:"required"`
	Pattern     string `form:"pattern" binding:"required"`
	MaxDuration int    `form:"max_duration"`
	MaxFilesize int    `form:"max_filesize"`
	Compress    bool   `form:"compress"`
}

// CreateChannel creates a new channel.
func CreateChannel(c *gin.Context) {
	var req *CreateChannelRequest
	if err := c.Bind(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("bind: %w", err))
		return
	}

	for _, username := range strings.Split(req.Username, ",") {
		server.Manager.CreateChannel(&entity.ChannelConfig{
			IsPaused:    false,
			Username:    username,
			Framerate:   req.Framerate,
			Resolution:  req.Resolution,
			Pattern:     req.Pattern,
			MaxDuration: req.MaxDuration,
			MaxFilesize: req.MaxFilesize,
			Compress:    req.Compress,
			CreatedAt:   time.Now().Unix(),
		}, true)
	}
	c.Redirect(http.StatusFound, "/")
}

// StopChannel stops a channel.
func StopChannel(c *gin.Context) {
	server.Manager.StopChannel(c.Param("username"))

	c.Redirect(http.StatusFound, "/")
}

// PauseChannel pauses a channel.
func PauseChannel(c *gin.Context) {
	server.Manager.PauseChannel(c.Param("username"))

	c.Redirect(http.StatusFound, "/")
}

// ResumeChannel resumes a paused channel.
func ResumeChannel(c *gin.Context) {
	server.Manager.ResumeChannel(c.Param("username"))

	c.Redirect(http.StatusFound, "/")
}

// Updates handles the SSE connection for updates.
func Updates(c *gin.Context) {
	server.Manager.Subscriber(c.Writer, c.Request)
}

// UpdateConfigRequest represents the request body for updating configuration.
type UpdateConfigRequest struct {
	Cookies   string `form:"cookies"`
	UserAgent string `form:"user_agent"`
}

// UpdateConfig updates the server configuration.
func UpdateConfig(c *gin.Context) {
	var req *UpdateConfigRequest
	if err := c.Bind(&req); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("bind: %w", err))
		return
	}

	server.Config.Cookies = req.Cookies
	server.Config.UserAgent = req.UserAgent
	c.Redirect(http.StatusFound, "/")
}
