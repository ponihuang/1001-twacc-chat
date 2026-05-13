package chat

import (
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFileStoreAcceptsDocx(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "source.docx")
	if err := os.WriteFile(sourcePath, []byte("docx"), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
	file, err := os.Open(sourcePath)
	if err != nil {
		t.Fatalf("Open error: %v", err)
	}
	defer file.Close()

	store := NewLocalFileStore(filepath.Join(tempDir, "uploads"))
	header := &multipart.FileHeader{
		Filename: "sample.docx",
		Header: textproto.MIMEHeader{
			"Content-Type": []string{"application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		},
		Size: 4,
	}

	attachment, messageType, err := store.Save(file, header)
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if messageType != "file" {
		t.Fatalf("messageType = %q, want file", messageType)
	}
	if attachment.OriginalName != "sample.docx" {
		t.Fatalf("OriginalName = %q, want sample.docx", attachment.OriginalName)
	}
}
