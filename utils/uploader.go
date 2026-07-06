package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UploadResult holds information about a saved file.
type UploadResult struct {
	FileName  string
	FilePath  string
	FileURL   string
	MimeType  string
	SizeBytes int64
}

// SaveFile saves a multipart file header to disk under uploadDir/subDir/
// and returns the URL using baseURL.
func SaveFile(fh *multipart.FileHeader, uploadDir, subDir, baseURL string) (*UploadResult, error) {
	dir := filepath.Join(uploadDir, subDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	uniqueName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	destPath := filepath.Join(dir, uniqueName)

	src, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	fileURL := fmt.Sprintf("%s/uploads/%s/%s", baseURL, subDir, uniqueName)
	return &UploadResult{
		FileName:  fh.Filename,
		FilePath:  destPath,
		FileURL:   fileURL,
		MimeType:  fh.Header.Get("Content-Type"),
		SizeBytes: fh.Size,
	}, nil
}

// DeleteFile removes a file from disk. Non-fatal if missing.
func DeleteFile(path string) {
	if path != "" {
		_ = os.Remove(path)
	}
}

// AllowedImageMime checks if the MIME type is an allowed image type.
func AllowedImageMime(mime string) bool {
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}
	return allowed[strings.ToLower(mime)]
}
