package handlers

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"habibi-go/internal/services"
)

type FileHandlers struct {
	sessionService *services.SessionService
	fileService    *services.FileService
}

func NewFileHandlers(sessionService *services.SessionService, fileService *services.FileService) *FileHandlers {
	return &FileHandlers{
		sessionService: sessionService,
		fileService:    fileService,
	}
}

// SearchFiles searches for files in the session's worktree
func (h *FileHandlers) SearchFiles(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	// Get session to find worktree path
	session, err := h.sessionService.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Search files
	files, err := h.fileService.SearchFiles(session.WorktreePath, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, files)
}

// ListFiles lists files in a directory
func (h *FileHandlers) ListFiles(c *gin.Context) {
	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	path := c.Query("path")
	
	// Get session to find worktree path
	session, err := h.sessionService.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Resolve full path
	fullPath := session.WorktreePath
	if path != "" {
		fullPath = filepath.Join(session.WorktreePath, path)
	}

	// Ensure path is within worktree
	if !strings.HasPrefix(fullPath, session.WorktreePath) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path outside worktree"})
		return
	}

	// List files
	files, err := h.fileService.ListFiles(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, files)
}