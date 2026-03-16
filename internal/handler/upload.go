package handler

import (
	"net/http"

	"github.com/teamsphere/server/internal/service"
	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	uploadService *service.UploadService
}

func NewUploadHandler(uploadService *service.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// Upload handles POST /upload.
func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Error(c, http.StatusBadRequest, 40001, "鏂囦欢涓嶈兘涓虹┖")
		return
	}

	result, err := h.uploadService.SaveFile(file)
	if err != nil {
		Error(c, http.StatusBadRequest, 40001, "鏂囦欢涓婁紶澶辫触锛岃妫€鏌ユ枃浠剁被鍨嬪拰澶у皬")
		return
	}

	Success(c, gin.H{"url": result.URL, "size": result.Size, "mime_type": result.MimeType})
}
