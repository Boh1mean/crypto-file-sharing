package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

type StoreContainerInput struct {
	ID             int
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

func (c *Client) StoreContainer(ctx context.Context, input StoreContainerInput) (StoreContainerOutput, error) {
	var out storeContainerResponse
	err := c.doJSON(ctx, http.MethodPost, "/files", storeContainerRequest{
		ID:          input.ID,
		SenderID:    input.SenderID,
		RecipientID: input.RecipientID,
		Container:   base64.StdEncoding.EncodeToString(input.ContainerBytes),
		FileName:    input.FileName,
		MimeType:    input.MimeType,
		Size:        input.Size,
	}, &out)
	if err != nil {
		return StoreContainerOutput{}, err
	}

	return StoreContainerOutput{ID: out.ID}, nil
}

func (c *Client) LoadContainer(ctx context.Context, id int) (LoadContainerOutput, error) {
	var out loadContainerResponse
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/files/%d", id), nil, &out); err != nil {
		return LoadContainerOutput{}, err
	}

	containerBytes, err := base64.StdEncoding.DecodeString(out.Container)
	if err != nil {
		return LoadContainerOutput{}, err
	}

	createdAt, err := time.Parse(http.TimeFormat, out.CreatedAt)
	if err != nil {
		return LoadContainerOutput{}, err
	}

	return LoadContainerOutput{
		ID:             out.ID,
		SenderID:       out.SenderID,
		RecipientID:    out.RecipientID,
		ContainerBytes: containerBytes,
		FileName:       out.FileName,
		MimeType:       out.MimeType,
		Size:           out.Size,
		CreatedAt:      createdAt,
	}, nil
}
