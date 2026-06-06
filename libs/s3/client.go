package s3

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	libconfig "github.com/even-app/even-app/libs/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client wraps MinIO for presigned uploads/downloads.
type Client struct {
	mc     *minio.Client
	bucket string
	public string
}

func New(cfg libconfig.S3) (*Client, error) {
	raw := strings.TrimPrefix(strings.TrimPrefix(cfg.Endpoint, "https://"), "http://")
	secure := strings.HasPrefix(cfg.Endpoint, "https://")
	mc, err := minio.New(raw, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: %w", err)
	}
	return &Client{mc: mc, bucket: cfg.Bucket, public: cfg.PublicEndpoint}, nil
}

func (c *Client) PresignPut(ctx context.Context, objectKey, contentType string) (string, error) {
	u, err := c.mc.PresignedPutObject(ctx, c.bucket, objectKey, 15*time.Minute)
	if err != nil {
		return "", err
	}
	return rewritePublic(u.String(), c.public), nil
}

func (c *Client) PresignGet(ctx context.Context, objectKey string) (string, error) {
	u, err := c.mc.PresignedGetObject(ctx, c.bucket, objectKey, time.Hour, url.Values{})
	if err != nil {
		return "", err
	}
	return rewritePublic(u.String(), c.public), nil
}

func rewritePublic(raw, public string) string {
	if public == "" {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	pub, err := url.Parse(public)
	if err != nil {
		return raw
	}
	u.Scheme = pub.Scheme
	u.Host = pub.Host
	return u.String()
}
