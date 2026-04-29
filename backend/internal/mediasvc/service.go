package mediasvc

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/pgxutil"
	catalogstore "github.com/bitik/backend/internal/store/catalog"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	_ "golang.org/x/image/webp"
)

const (
	defaultPresignExpiry = 15 * time.Minute
)

var (
	errUnsupportedMedia = errors.New("unsupported media type")
	errTooLarge         = errors.New("file too large")
	errInvalidImage     = errors.New("invalid image dimensions")
)

type MalwareScanner interface {
	Scan(ctx context.Context, filename string, contentType string, body []byte) error
}

type ImageProcessor interface {
	Enqueue(ctx context.Context, fileID uuid.UUID) error
}

type Option func(*Service)

func WithMalwareScanner(scanner MalwareScanner) Option {
	return func(s *Service) { s.scanner = scanner }
}

func WithImageProcessor(processor ImageProcessor) Option {
	return func(s *Service) { s.processor = processor }
}

type Service struct {
	cfg       config.Config
	log       *zap.Logger
	queries   *catalogstore.Queries
	storage   Storage
	scanner   MalwareScanner
	processor ImageProcessor
}

func NewService(cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool, storage Storage, opts ...Option) *Service {
	s := &Service{
		cfg:     cfg,
		log:     logger,
		queries: catalogstore.New(pool),
		storage: storage,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type fileMeta struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Extension   string `json:"extension"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	Source      string `json:"source"`
	Status      string `json:"status"`
}

func (s *Service) ownerID(c *gin.Context) (uuid.UUID, bool) {
	raw, ok := c.Get(middleware.AuthUserIDKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := raw.(uuid.UUID)
	return id, ok
}

func validateRoles(c *gin.Context) bool {
	raw, _ := c.Get(middleware.AuthRolesKey)
	roles, _ := raw.([]string)
	for _, role := range roles {
		if role == "admin" || role == "seller" || role == "buyer" {
			return true
		}
	}
	return false
}

func (s *Service) validateFile(ctx context.Context, filename, contentType string, body []byte) (fileMeta, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	detectedType := http.DetectContentType(body)
	allowed := map[string][]string{
		"image/jpeg": {".jpg", ".jpeg"},
		"image/png":  {".png"},
		"image/webp": {".webp"},
		"image/gif":  {".gif"},
	}
	exts, ok := allowed[detectedType]
	if !ok {
		return fileMeta{}, errUnsupportedMedia
	}
	if !contains(exts, ext) {
		return fileMeta{}, errUnsupportedMedia
	}
	if int64(len(body)) > s.maxUploadBytes() {
		return fileMeta{}, errTooLarge
	}
	if contentType != "" && contentType != detectedType {
		return fileMeta{}, errUnsupportedMedia
	}
	meta := fileMeta{Filename: filename, ContentType: detectedType, Extension: ext, Source: "upload", Status: "ready"}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil {
		return fileMeta{}, errInvalidImage
	}
	if cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 8000 || cfg.Height > 8000 {
		return fileMeta{}, errInvalidImage
	}
	meta.Width = cfg.Width
	meta.Height = cfg.Height
	if s.scanner != nil {
		if err := s.scanner.Scan(ctx, filename, detectedType, body); err != nil {
			return fileMeta{}, err
		}
	}
	return meta, nil
}

func (s *Service) objectKey(owner uuid.UUID, ext string) (string, error) {
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("media/%s/%s%s", owner.String(), hex.EncodeToString(random[:]), ext), nil
}

func (s *Service) createMedia(ctx context.Context, owner uuid.UUID, url, bucket, objectKey string, meta fileMeta, size int64) (catalogstore.MediaFile, error) {
	metadata, err := json.Marshal(meta)
	if err != nil {
		return catalogstore.MediaFile{}, err
	}
	return s.queries.CreateMediaFile(ctx, catalogstore.CreateMediaFileParams{
		OwnerUserID: pgxutil.UUID(owner),
		Url:         url,
		Bucket:      pgtype.Text{String: bucket, Valid: bucket != ""},
		ObjectKey:   pgtype.Text{String: objectKey, Valid: objectKey != ""},
		MimeType:    pgtype.Text{String: meta.ContentType, Valid: meta.ContentType != ""},
		SizeBytes:   pgtype.Int8{Int64: size, Valid: size >= 0},
		Metadata:    metadata,
	})
}

func mediaJSON(file catalogstore.MediaFile) gin.H {
	return gin.H{
		"id":            uuidString(file.ID),
		"owner_user_id": uuidString(file.OwnerUserID),
		"url":           file.Url,
		"bucket":        textValue(file.Bucket),
		"object_key":    textValue(file.ObjectKey),
		"mime_type":     textValue(file.MimeType),
		"size_bytes":    int8Value(file.SizeBytes),
		"metadata":      json.RawMessage(file.Metadata),
		"created_at":    file.CreatedAt.Time,
	}
}

func uuidString(id pgtype.UUID) string {
	if v, ok := pgxutil.ToUUID(id); ok {
		return v.String()
	}
	return ""
}

func textValue(t pgtype.Text) any {
	if !t.Valid {
		return nil
	}
	return t.String
}

func int8Value(v pgtype.Int8) any {
	if !v.Valid {
		return nil
	}
	return v.Int64
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func readAllLimited(r io.Reader, maxUploadBytes int64) ([]byte, error) {
	if maxUploadBytes <= 0 {
		maxUploadBytes = 10 << 20
	}
	limited := io.LimitReader(r, maxUploadBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxUploadBytes {
		return nil, errTooLarge
	}
	return body, nil
}

func (s *Service) maxUploadBytes() int64 {
	if s.cfg.Security.MaxUploadBytes > 0 {
		return s.cfg.Security.MaxUploadBytes
	}
	return 10 << 20
}
