package files

import "time"

type File struct {
	ID          string
	OwnerUserID string
	Filename    string
	MimeType    string
	SizeBytes   int64
	StorageKey  string
	CreatedAt   time.Time
}

type FileResponse struct {
	ID          string    `json:"id"`
	OwnerUserID string    `json:"ownerUserId"`
	Filename    string    `json:"filename"`
	MimeType    string    `json:"mimeType"`
	SizeBytes   int64     `json:"sizeBytes"`
	StorageKey  string    `json:"storageKey"`
	CreatedAt   time.Time `json:"createdAt"`
}

type FilesResponse struct {
	Files []FileResponse `json:"files"`
}

type CreateFileRequest struct {
	Filename   string `json:"filename" binding:"required"`
	MimeType   string `json:"mimeType" binding:"required"`
	SizeBytes  int64  `json:"sizeBytes" binding:"min=0"`
	StorageKey string `json:"storageKey" binding:"required"`
}
