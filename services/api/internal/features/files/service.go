package files

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrFilenameRequired = errors.New("filename required")
	ErrMimeTypeRequired = errors.New("mime type required")
	ErrStorageKeyNeeded = errors.New("storage key required")
	ErrInvalidFileSize  = errors.New("invalid file size")
)

type Service struct {
	repo Repository
}

const (
	defaultFilesLimit = 20
	maxFilesLimit     = 100
)

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListByOwner(ctx context.Context, ownerUserID string, limit, offset int) ([]File, error) {
	normalizedLimit := normalizeFilesLimit(limit)
	normalizedOffset := normalizeFilesOffset(offset)
	files, err := s.repo.ListFilesByOwner(ctx, ownerUserID, normalizedLimit, normalizedOffset)
	if err != nil {
		return nil, err
	}

	payload := make([]File, 0, len(files))
	for _, file := range files {
		payload = append(payload, mapStoredFile(file))
	}
	return payload, nil
}

func (s *Service) Create(ctx context.Context, ownerUserID, filename, mimeType string, sizeBytes int64, storageKey string) (File, error) {
	normalizedFilename := strings.TrimSpace(filename)
	if normalizedFilename == "" {
		return File{}, ErrFilenameRequired
	}
	normalizedMimeType := strings.TrimSpace(mimeType)
	if normalizedMimeType == "" {
		return File{}, ErrMimeTypeRequired
	}
	if sizeBytes < 0 {
		return File{}, ErrInvalidFileSize
	}
	normalizedStorageKey := strings.TrimSpace(storageKey)
	if normalizedStorageKey == "" {
		return File{}, ErrStorageKeyNeeded
	}

	file, err := s.repo.CreateFile(ctx, ownerUserID, normalizedFilename, normalizedMimeType, sizeBytes, normalizedStorageKey)
	if err != nil {
		return File{}, err
	}
	return mapStoredFile(file), nil
}

func (s *Service) GetByID(ctx context.Context, ownerUserID, fileID string) (File, error) {
	file, err := s.repo.GetFileByID(ctx, ownerUserID, fileID)
	if err != nil {
		if errors.Is(err, ErrFileNotFound) {
			return File{}, ErrFileNotFound
		}
		return File{}, err
	}
	return mapStoredFile(file), nil
}

func mapStoredFile(file StoredFile) File {
	return File{
		ID:          file.ID,
		OwnerUserID: file.OwnerUserID,
		Filename:    file.Filename,
		MimeType:    file.MimeType,
		SizeBytes:   file.SizeBytes,
		StorageKey:  file.StorageKey,
		CreatedAt:   file.CreatedAt,
	}
}

func normalizeFilesLimit(limit int) int {
	if limit <= 0 {
		return defaultFilesLimit
	}
	if limit > maxFilesLimit {
		return maxFilesLimit
	}
	return limit
}

func normalizeFilesOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
