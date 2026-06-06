package s3

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	libconfig "github.com/even-app/even-app/libs/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const defaultRegion = "us-east-1"

// Client wraps MinIO for presigned uploads/downloads.
type Client struct {
	mc      *minio.Client // internal endpoint for service-side object ops
	presign *minio.Client // public host for presigned URLs (SigV4 includes Host)
	bucket  string
}

func New(cfg libconfig.S3) (*Client, error) {
	creds := credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, "")

	intHost, intSecure, err := parseEndpoint(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("S3_ENDPOINT: %w", err)
	}
	mc, err := minio.New(intHost, &minio.Options{
		Creds:  creds,
		Secure: intSecure,
		Region: defaultRegion,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: %w", err)
	}

	pubHost, pubSecure, err := parseEndpoint(cfg.PublicEndpoint)
	if err != nil {
		return nil, fmt.Errorf("S3_PUBLIC_ENDPOINT: %w", err)
	}
	if pubHost == "" {
		pubHost = intHost
		pubSecure = intSecure
	}

	presign := mc
	if pubHost != intHost || pubSecure != intSecure {
		opts := &minio.Options{
			Creds:     creds,
			Secure:    pubSecure,
			Region:    defaultRegion,
			Transport: dialRedirectTransport(intHost),
		}
		presign, err = minio.New(pubHost, &minio.Options{
			Creds:     opts.Creds,
			Secure:    opts.Secure,
			Region:    opts.Region,
			Transport: opts.Transport,
		})
		if err != nil {
			return nil, fmt.Errorf("minio presign: %w", err)
		}
	}

	return &Client{mc: mc, presign: presign, bucket: cfg.Bucket}, nil
}

func parseEndpoint(raw string) (host string, secure bool, err error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", false, fmt.Errorf("invalid endpoint %q", raw)
	}
	return u.Host, u.Scheme == "https", nil
}

// dialRedirectTransport connects to internalHost while keeping the public URL host for signing.
func dialRedirectTransport(internalHost string) *http.Transport {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		base = &http.Transport{}
	}
	t := base.Clone()
	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	t.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, internalHost)
	}
	return t
}

func (c *Client) PresignPut(ctx context.Context, objectKey, contentType string) (string, error) {
	u, err := c.presign.PresignedPutObject(ctx, c.bucket, objectKey, 15*time.Minute)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (c *Client) PresignGet(ctx context.Context, objectKey string) (string, error) {
	u, err := c.presign.PresignedGetObject(ctx, c.bucket, objectKey, time.Hour, url.Values{})
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
