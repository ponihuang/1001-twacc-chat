package chat

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

const maxAttachmentSize = 40 << 20

var (
	allowedImageExts = map[string]string{
		".jpg":  "image",
		".jpeg": "image",
		".png":  "image",
	}
	allowedFileExts = map[string]string{
		".doc":  "file",
		".docx": "file",
		".pdf":  "file",
		".xlsx": "file",
	}
)

// FileStore persists uploaded message attachments.
type FileStore interface {
	Save(file multipart.File, header *multipart.FileHeader) (*AttachmentInput, string, error)
}

// LocalFileStore writes attachments to a local directory served by the app.
type LocalFileStore struct {
	dir string
}

// NewLocalFileStore builds a local attachment store.
func NewLocalFileStore(dir string) *LocalFileStore {
	return &LocalFileStore{dir: dir}
}

// Save validates and writes the uploaded file.
func (s *LocalFileStore) Save(file multipart.File, header *multipart.FileHeader) (*AttachmentInput, string, error) {
	if s == nil || strings.TrimSpace(s.dir) == "" {
		return nil, "", fmt.Errorf("file store unavailable")
	}

	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(header.Filename)))
	messageType, ok := allowedImageExts[ext]
	if !ok {
		messageType, ok = allowedFileExts[ext]
	}
	if !ok {
		return nil, "", ErrUnsupportedAttachmentType
	}
	if header.Size <= 0 {
		return nil, "", ErrAttachmentRequired
	}
	if header.Size > maxAttachmentSize {
		return nil, "", ErrAttachmentTooLarge
	}

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return nil, "", fmt.Errorf("create upload dir: %w", err)
	}

	storedName, err := randomFileName(ext)
	if err != nil {
		return nil, "", fmt.Errorf("generate upload name: %w", err)
	}

	targetPath := filepath.Join(s.dir, storedName)
	dst, err := os.Create(targetPath)
	if err != nil {
		return nil, "", fmt.Errorf("create upload file: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, io.LimitReader(file, maxAttachmentSize+1))
	if err != nil {
		return nil, "", fmt.Errorf("write upload file: %w", err)
	}
	if written > maxAttachmentSize {
		return nil, "", ErrAttachmentTooLarge
	}

	mimeType := strings.TrimSpace(header.Header.Get("Content-Type"))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return &AttachmentInput{
		OriginalName: header.Filename,
		StoragePath:  "/uploads/" + storedName,
		MIMEType:     mimeType,
		SizeBytes:    written,
	}, messageType, nil
}

func randomFileName(ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf) + ext, nil
}
