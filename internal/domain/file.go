package domain

import "time"

type FileRecord struct {
	ID          int
	Size        int64
	SenderID    int
	RecipientID int
	StorageKey  string
	FileName    string
	MimeType    string
	CreatedAt   time.Time
}

type StoreContainerInput struct {
	SenderID       int
	RecipientID    int
	ContainerBytes []byte
	FileName       string
	MimeType       string
	Size           int64
}

type StoreContainerOutput struct {
	ID int
}

type LoadContainerInput struct {
	ID int
}

type LoadContainerOutput struct {
	ID             int
	SenderID       int
	RecipientID    int
	ContainerBytes []byte
	FileName       string
	MimeType       string
	Size           int64
	CreatedAt      time.Time
}
