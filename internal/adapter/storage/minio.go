// Package storage implements app.Storage on MinIO / any S3-compatible store.
// The bucket is private: nothing in it is reachable without going through
// the Go server (public images) or a short presigned link (CVs).
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/stevenwilliam/sunshinefacility_compro/internal/app"
)

type MinIO struct {
	c      *minio.Client
	bucket string
}

var _ app.Storage = (*MinIO)(nil)

func NewMinIO(endpoint, access, secret, bucket string, useSSL bool) (*MinIO, error) {
	c, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(access, secret, ""), Secure: useSSL})
	if err != nil {
		return nil, err
	}
	return &MinIO{c: c, bucket: bucket}, nil
}

// EnsureBucket creates the bucket if it is missing. It never sets a public
// policy.
func (m *MinIO) EnsureBucket(ctx context.Context) error {
	ok, err := m.c.BucketExists(ctx, m.bucket)
	if err != nil {
		return fmt.Errorf("object store unreachable: %w", err)
	}
	if ok {
		return nil
	}
	return m.c.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{})
}

func (m *MinIO) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := m.c.PutObject(ctx, m.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (m *MinIO) Get(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	o, err := m.c.GetObject(ctx, m.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, err
	}
	st, err := o.Stat()
	if err != nil {
		_ = o.Close()
		var er minio.ErrorResponse
		if errors.As(err, &er) && er.Code == "NoSuchKey" {
			return nil, 0, app.ErrNotFound
		}
		return nil, 0, err
	}
	return o, st.Size, nil
}

func (m *MinIO) Delete(ctx context.Context, key string) error {
	return m.c.RemoveObject(ctx, m.bucket, key, minio.RemoveObjectOptions{})
}

// PresignGet returns a short-lived link that downloads under downloadName.
func (m *MinIO) PresignGet(ctx context.Context, key string, ttl time.Duration, downloadName string) (string, error) {
	params := url.Values{}
	params.Set("response-content-disposition", fmt.Sprintf("attachment; filename=%q", downloadName))
	u, err := m.c.PresignedGetObject(ctx, m.bucket, key, ttl, params)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (m *MinIO) Ping(ctx context.Context) error {
	_, err := m.c.BucketExists(ctx, m.bucket)
	return err
}
