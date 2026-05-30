package ticket

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/husky/husky/internal/model"
)

func (s *service) UploadAttachment(ctx context.Context, ticketID, userID uint, fileName string, file io.Reader) (*model.Attachment, error) {
	if _, err := s.ticketRepo.GetByID(ctx, ticketID); err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	ext := filepath.Ext(fileName)
	storageName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload dir: %w", err)
	}
	filePath := filepath.Join(uploadDir, storageName)

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if err := dst.Close(); err != nil {
			log.Printf("failed to close attachment file: %v", err)
		}
	}()

	written, err := io.Copy(dst, file)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	contentType := detectContentType(fileName)

	att := &model.Attachment{
		TicketID:   ticketID,
		FileName:   fileName,
		FileSize:   written,
		FileType:   contentType,
		FileURL:    filePath,
		UploadedBy: userID,
	}

	if err := s.ticketRepo.CreateAttachment(ctx, att); err != nil {
		if rmErr := os.Remove(filePath); rmErr != nil {
			log.Printf("failed to remove attachment file after DB error: %v", rmErr)
		}
		return nil, fmt.Errorf("failed to save attachment record: %w", err)
	}

	return att, nil
}

func (s *service) ListAttachments(ctx context.Context, ticketID uint) ([]model.Attachment, error) {
	return s.ticketRepo.ListAttachments(ctx, ticketID)
}

func (s *service) DeleteAttachment(ctx context.Context, id uint) error {
	att, err := s.ticketRepo.GetAttachment(ctx, id)
	if err != nil {
		return err
	}
	if err := s.ticketRepo.DeleteAttachment(ctx, id); err != nil {
		return err
	}
	if err := os.Remove(att.FileURL); err != nil {
		log.Printf("failed to remove attachment file: %v", err)
	}
	return nil
}

func detectContentType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".doc", ".docx":
		return "application/msword"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	case ".zip":
		return "application/zip"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
