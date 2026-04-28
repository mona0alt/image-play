package storage

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

const defaultSignedURLTTL = time.Hour

type Signer interface {
	SignImageURL(rawURL string) string
}

type OSSSigner struct {
	bucket   *oss.Bucket
	host     string
	ttl      time.Duration
}

func NewOSSSigner(endpoint, bucket, accessKeyID, accessKeySecret string) (*OSSSigner, error) {
	if endpoint == "" || bucket == "" || accessKeyID == "" || accessKeySecret == "" {
		return nil, fmt.Errorf("missing oss configuration")
	}

	bareEndpoint := strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
	clientEndpoint := endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		clientEndpoint = "https://" + bareEndpoint
	}

	client, err := oss.New(clientEndpoint, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("oss new client: %w", err)
	}

	b, err := client.Bucket(bucket)
	if err != nil {
		return nil, fmt.Errorf("oss get bucket: %w", err)
	}

	return &OSSSigner{
		bucket: b,
		host:   fmt.Sprintf("%s.%s", bucket, bareEndpoint),
		ttl:    defaultSignedURLTTL,
	}, nil
}

func (s *OSSSigner) SignImageURL(rawURL string) string {
	if s == nil || rawURL == "" {
		return rawURL
	}

	key := s.objectKeyFromURL(rawURL)
	if key == "" {
		return rawURL
	}

	signed, err := s.bucket.SignURL(key, oss.HTTPGet, int64(s.ttl.Seconds()))
	if err != nil {
		return rawURL
	}
	return restorePathSlashes(signed)
}

func restorePathSlashes(signed string) string {
	q := strings.IndexByte(signed, '?')
	if q < 0 {
		return strings.ReplaceAll(signed, "%2F", "/")
	}
	return strings.ReplaceAll(signed[:q], "%2F", "/") + signed[q:]
}

func (s *OSSSigner) objectKeyFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	if u.Host != s.host {
		return ""
	}
	return strings.TrimPrefix(u.Path, "/")
}

type NoopSigner struct{}

func (NoopSigner) SignImageURL(rawURL string) string { return rawURL }
