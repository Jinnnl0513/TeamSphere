package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/teamsphere/server/internal/config"
	"github.com/google/uuid"
)

var safeExtPattern = regexp.MustCompile(`^\.[a-z0-9]{1,16}$`)

type UploadResult struct {
	URL      string
	Size     int64
	MimeType string
}

type UploadService struct {
	storageCfg *config.StorageConfig
	fileTokenSecret string
}

func NewUploadService(storageCfg *config.StorageConfig, fileTokenSecret string) *UploadService {
	return &UploadService{storageCfg: storageCfg, fileTokenSecret: fileTokenSecret}
}

// SaveFile validates and saves an uploaded file, returning the URL path (e.g. "/uploads/uuid.pdf").
// Fix #3: Uses server-side MIME sniffing (http.DetectContentType) instead of trusting the
// client-supplied Content-Type header, preventing content-type spoofing attacks.
func (s *UploadService) SaveFile(header *multipart.FileHeader) (*UploadResult, error) {
	// Validate size first (cheaper check)
	if header.Size > s.storageCfg.MaxFileSize {
		return nil, ErrFileTooLarge
	}

	// Open source file
	src, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("open uploaded file: %w", err)
	}
	defer src.Close()

	// Read first 512 bytes for MIME sniffing (server-side detection)
	sniffBuf := make([]byte, 512)
	n, err := src.Read(sniffBuf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read file header: %w", err)
	}
	detectedType := http.DetectContentType(sniffBuf[:n])

	// Normalize: DetectContentType may return "image/jpeg" or "image/jpeg; charset=..." style
	if idx := strings.Index(detectedType, ";"); idx != -1 {
		detectedType = strings.TrimSpace(detectedType[:idx])
	}

	// Allow list of safe content types.
	if !allowedMimeTypes[detectedType] {
		return nil, ErrFileTypeNotAllowed
	}

	// Determine extension: prefer user extension (if safe), then MIME-derived extension.
	ext := sanitizeExt(filepath.Ext(header.Filename))
	if ext == "" {
		if exts, err := mime.ExtensionsByType(detectedType); err == nil && len(exts) > 0 {
			ext = sanitizeExt(exts[0])
		}
	}
	if ext == "" {
		ext = ".bin"
	}

	// Generate UUID filename
	filename := uuid.NewString() + ext
	dstPath := filepath.Join(s.storageCfg.UploadDir, filename)

	// Ensure upload directory exists
	if err := os.MkdirAll(s.storageCfg.UploadDir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	// Create destination
	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	// Read full content for optional image re-encode
	content, err := io.ReadAll(io.MultiReader(bytes.NewReader(sniffBuf[:n]), src))
	if err != nil {
		return nil, fmt.Errorf("read file content: %w", err)
	}

	var finalSize int64 = header.Size
	if isImageType(detectedType) {
		encExt, encType, encBytes, encErr := reencodeImage(content, detectedType)
		if encErr != nil {
			return nil, ErrFileTypeNotAllowed
		}
		detectedType = encType
		if encExt != "" {
			ext = encExt
			filename = uuid.NewString() + ext
			dstPath = filepath.Join(s.storageCfg.UploadDir, filename)
			dst.Close()
			dst, err = os.Create(dstPath)
			if err != nil {
				return nil, fmt.Errorf("create file: %w", err)
			}
			defer dst.Close()
		}
		if _, err := dst.Write(encBytes); err != nil {
			return nil, fmt.Errorf("save reencoded image: %w", err)
		}
		finalSize = int64(len(encBytes))
	} else {
		if _, err := dst.Write(content); err != nil {
			return nil, fmt.Errorf("save file: %w", err)
		}
		finalSize = int64(len(content))
	}

	// Return signed URL path with forward slashes
	urlPath := "/uploads/" + filename
	urlPath = strings.ReplaceAll(urlPath, "\\", "/")
	if s.fileTokenSecret != "" {
		urlPath = s.signURL(urlPath)
	}
	return &UploadResult{URL: urlPath, Size: finalSize, MimeType: detectedType}, nil
}

func sanitizeExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if safeExtPattern.MatchString(ext) {
		return ext
	}
	return ""
}

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"application/pdf": true,
	"text/plain": true,
	"audio/mpeg": true,
	"audio/wav":  true,
	"audio/ogg":  true,
	"audio/webm": true,
	"video/mp4":  true,
	"video/webm": true,
}

func isImageType(mt string) bool {
	return mt == "image/jpeg" || mt == "image/png"
}

func reencodeImage(content []byte, mimeType string) (string, string, []byte, error) {
	img, format, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		return "", "", nil, err
	}
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return "", "", nil, err
		}
		return ".jpg", "image/jpeg", buf.Bytes(), nil
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return "", "", nil, err
		}
		return ".png", "image/png", buf.Bytes(), nil
	default:
		return "", "", nil, fmt.Errorf("unsupported image format")
	}
}

func (s *UploadService) signURL(path string) string {
	exp := time.Now().Add(24 * time.Hour).Unix()
	expStr := fmt.Sprintf("%d", exp)
	mac := hmac.New(sha256.New, []byte(s.fileTokenSecret))
	mac.Write([]byte(strings.TrimPrefix(path, "/uploads/")))
	mac.Write([]byte("|"))
	mac.Write([]byte(expStr))
	token := hex.EncodeToString(mac.Sum(nil))
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return fmt.Sprintf("%s%sexp=%s&token=%s", path, sep, expStr, token)
}
