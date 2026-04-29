package mediasvc

import (
	"context"
	"encoding/base64"
	"testing"
)

func TestValidateFileAcceptsPNGAndDimensions(t *testing.T) {
	body, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	svc := &Service{}

	meta, err := svc.validateFile(context.Background(), "image.png", "image/png", body)
	if err != nil {
		t.Fatalf("validateFile returned error: %v", err)
	}
	if meta.ContentType != "image/png" || meta.Extension != ".png" || meta.Width != 1 || meta.Height != 1 {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
}

func TestValidateFileRejectsUnsupportedExtension(t *testing.T) {
	body := []byte("not an image")
	svc := &Service{}

	_, err := svc.validateFile(context.Background(), "image.txt", "text/plain", body)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateFileRejectsSpoofedImageHeader(t *testing.T) {
	body := []byte("not an image")
	svc := &Service{}

	_, err := svc.validateFile(context.Background(), "image.webp", "image/webp", body)
	if err == nil {
		t.Fatal("expected spoofed image to be rejected")
	}
}

func TestValidatePresignedRejectsMismatchedExtension(t *testing.T) {
	svc := &Service{}

	_, err := svc.validatePresigned("image.txt", "image/png")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidatePresignedMarksPending(t *testing.T) {
	svc := &Service{}

	meta, err := svc.validatePresigned("image.png", "image/png")
	if err != nil {
		t.Fatalf("validatePresigned returned error: %v", err)
	}
	if meta.Status != "pending" || meta.Source != "presigned" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
}
