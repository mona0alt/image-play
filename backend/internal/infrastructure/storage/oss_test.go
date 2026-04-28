package storage

import (
	"strings"
	"testing"
)

func newTestSigner(t *testing.T) *OSSSigner {
	t.Helper()
	s, err := NewOSSSigner("oss-cn-beijing.aliyuncs.com", "imag-play", "fake-ak", "fake-sk")
	if err != nil {
		t.Fatalf("create signer: %v", err)
	}
	return s
}

func TestOSSSigner_SignImageURL_Empty(t *testing.T) {
	s := newTestSigner(t)
	if got := s.SignImageURL(""); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestOSSSigner_SignImageURL_NilSafe(t *testing.T) {
	var s *OSSSigner
	url := "https://example.com/foo.jpg"
	if got := s.SignImageURL(url); got != url {
		t.Fatalf("expected pass-through, got %q", got)
	}
}

func TestOSSSigner_SignImageURL_ForeignHost(t *testing.T) {
	s := newTestSigner(t)
	url := "https://other.example.com/explore/abc.jpg"
	if got := s.SignImageURL(url); got != url {
		t.Fatalf("expected unchanged URL for foreign host, got %q", got)
	}
}

func TestOSSSigner_SignImageURL_BucketURL(t *testing.T) {
	s := newTestSigner(t)
	raw := "https://imag-play.oss-cn-beijing.aliyuncs.com/explore/vDSOoBjCd6W09o.jpeg"
	signed := s.SignImageURL(raw)

	if !strings.HasPrefix(signed, "https://imag-play.oss-cn-beijing.aliyuncs.com/explore/vDSOoBjCd6W09o.jpeg?") {
		t.Errorf("signed URL should preserve scheme/host/path, got: %s", signed)
	}
	if !strings.Contains(signed, "OSSAccessKeyId=fake-ak") {
		t.Errorf("signed URL missing OSSAccessKeyId: %s", signed)
	}
	if !strings.Contains(signed, "Signature=") {
		t.Errorf("signed URL missing Signature: %s", signed)
	}
	if !strings.Contains(signed, "Expires=") {
		t.Errorf("signed URL missing Expires: %s", signed)
	}
	if strings.Contains(strings.SplitN(signed, "?", 2)[0], "%2F") {
		t.Errorf("path should have decoded slashes: %s", signed)
	}
}

func TestRestorePathSlashes(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://h/a%2Fb%2Fc.jpg", "https://h/a/b/c.jpg"},
		{"https://h/a%2Fb.jpg?Signature=x%2Fy", "https://h/a/b.jpg?Signature=x%2Fy"},
		{"https://h/no-slashes.jpg?x=1", "https://h/no-slashes.jpg?x=1"},
	}
	for _, tt := range tests {
		if got := restorePathSlashes(tt.in); got != tt.want {
			t.Errorf("restorePathSlashes(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestOSSSigner_objectKeyFromURL(t *testing.T) {
	s := newTestSigner(t)

	tests := []struct {
		name string
		url  string
		want string
	}{
		{"normal explore", "https://imag-play.oss-cn-beijing.aliyuncs.com/explore/abc.jpg", "explore/abc.jpg"},
		{"nested key", "https://imag-play.oss-cn-beijing.aliyuncs.com/users/1/foo.png", "users/1/foo.png"},
		{"foreign host", "https://other.example.com/explore/abc.jpg", ""},
		{"malformed", "::not a url", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.objectKeyFromURL(tt.url); got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
			}
		})
	}
}
