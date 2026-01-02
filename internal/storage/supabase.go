package storage

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"path"
	"time"

	"papertrail/internal/config"
)

type SupabaseClient struct {
	url    string
	key    string
	bucket string
	client *http.Client
}

func NewSupabaseClient(cfg *config.Config) *SupabaseClient {
	return &SupabaseClient{
		url:    cfg.SupabaseURL,
		key:    cfg.SupabaseKey,
		bucket: cfg.SupabaseBucket,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// UploadFile pushes a PDF to Supabase Storage and returns the public path. Minimal, non-streaming example.
func (s *SupabaseClient) UploadFile(filename string, content []byte) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(content); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.url, s.bucket, filename)

	req, err := http.NewRequest(http.MethodPost, endpoint, &buf)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.key))
	req.Header.Set("x-upsert", "true")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("supabase upload failed: %s", resp.Status)
	}

	return path.Join("/storage/v1/object", s.bucket, filename), nil
}
