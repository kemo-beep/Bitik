package mediasvc

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/authsvc"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/pgxutil"
	catalogstore "github.com/bitik/backend/internal/store/catalog"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) RegisterRoutes(rg *gin.RouterGroup, auth *authsvc.Service) {
	media := rg.Group("/media")
	if auth != nil {
		media.Use(middleware.RequireBearerJWT(s.cfg), auth.RequireActiveUser())
	}
	media.POST("/upload", s.requireMarketplaceUser(), s.HandleUpload)
	media.POST("/upload/presigned-url", s.requireMarketplaceUser(), s.HandlePresignedURL)
	media.POST("/upload/presigned-complete", s.requireMarketplaceUser(), s.HandlePresignedComplete)
	media.GET("/files", s.HandleListFiles)
	media.GET("/files/:file_id", s.HandleGetFile)
	media.DELETE("/files/:file_id", s.HandleDeleteFile)
}

func (s *Service) HandleUpload(c *gin.Context) {
	if s.storage == nil {
		apiresponse.Error(c, http.StatusServiceUnavailable, "storage_unavailable", "Storage is not configured.")
		return
	}
	owner, ok := s.ownerID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "file_required", "Upload file is required.")
		return
	}
	f, err := file.Open()
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "file_open_failed", "Could not read upload.")
		return
	}
	defer f.Close()
	body, err := readAllLimited(f, s.maxUploadBytes())
	if err != nil {
		writeMediaError(c, err)
		return
	}
	meta, err := s.validateFile(c.Request.Context(), file.Filename, file.Header.Get("Content-Type"), body)
	if err != nil {
		writeMediaError(c, err)
		return
	}
	key, err := s.objectKey(owner, meta.Extension)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not allocate object key.")
		return
	}
	url, err := s.storage.Put(c.Request.Context(), key, meta.ContentType, int64(len(body)), bytes.NewReader(body))
	if err != nil {
		apiresponse.Error(c, http.StatusBadGateway, "storage_error", "Could not upload file.")
		return
	}
	created, err := s.createMedia(c.Request.Context(), owner, url, s.storage.Bucket(), key, meta, int64(len(body)))
	if err != nil {
		_ = s.storage.Delete(c.Request.Context(), key)
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not record media file.")
		return
	}
	if s.processor != nil {
		if id, ok := uuidFromPG(created.ID); ok {
			_ = s.processor.Enqueue(c.Request.Context(), id)
		}
	}
	apiresponse.Respond(c, http.StatusCreated, mediaJSON(created), nil)
}

func (s *Service) HandlePresignedURL(c *gin.Context) {
	if s.storage == nil {
		apiresponse.Error(c, http.StatusServiceUnavailable, "storage_unavailable", "Storage is not configured.")
		return
	}
	owner, ok := s.ownerID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	var req struct {
		Filename    string `json:"filename" binding:"required"`
		ContentType string `json:"content_type" binding:"required"`
		SizeBytes   int64  `json:"size_bytes" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid presigned upload request.")
		return
	}
	if req.SizeBytes > s.maxUploadBytes() {
		apiresponse.Error(c, http.StatusBadRequest, "file_too_large", "File exceeds maximum upload size.")
		return
	}
	meta, err := s.validatePresigned(req.Filename, req.ContentType)
	if err != nil {
		writeMediaError(c, err)
		return
	}
	key, err := s.objectKey(owner, meta.Extension)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not allocate object key.")
		return
	}
	uploadURL, publicURL, err := s.storage.PresignPut(c.Request.Context(), key, meta.ContentType, defaultPresignExpiry)
	if err != nil {
		apiresponse.Error(c, http.StatusBadGateway, "storage_error", "Could not create presigned upload URL.")
		return
	}
	meta.Source = "presigned"
	meta.Status = "pending"
	created, err := s.createMedia(c.Request.Context(), owner, publicURL, s.storage.Bucket(), key, meta, req.SizeBytes)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not record media file.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, gin.H{
		"file":       mediaJSON(created),
		"upload_url": uploadURL,
		"expires_in": int(defaultPresignExpiry.Seconds()),
	}, nil)
}

func (s *Service) HandlePresignedComplete(c *gin.Context) {
	if s.storage == nil {
		apiresponse.Error(c, http.StatusServiceUnavailable, "storage_unavailable", "Storage is not configured.")
		return
	}
	owner, ok := s.ownerID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	var req struct {
		FileID string `json:"file_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid presigned completion request.")
		return
	}
	fileID, err := uuid.Parse(req.FileID)
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_file_id", "Invalid file id.")
		return
	}
	file, err := s.queries.GetMediaFileByID(c.Request.Context(), pgxutil.UUID(fileID))
	if err != nil {
		writeMediaNotFound(c, err)
		return
	}
	if uuidString(file.OwnerUserID) != owner.String() {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "You cannot complete this media file.")
		return
	}
	if !file.ObjectKey.Valid {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_media_file", "Media file has no object key.")
		return
	}
	existingMeta := decodeFileMeta(file.Metadata)
	body, objectContentType, objectSize, err := s.storage.Get(c.Request.Context(), file.ObjectKey.String, s.maxUploadBytes())
	if err != nil {
		writeMediaError(c, err)
		return
	}
	contentType := existingMeta.ContentType
	if objectContentType != "" && objectContentType != "application/octet-stream" {
		contentType = objectContentType
	}
	meta, err := s.validateFile(c.Request.Context(), existingMeta.Filename, contentType, body)
	if err != nil {
		writeMediaError(c, err)
		return
	}
	meta.Source = "presigned"
	meta.Status = "ready"
	url := s.storage.PublicURL(file.ObjectKey.String)
	metadata, err := json.Marshal(meta)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not encode media metadata.")
		return
	}
	updated, err := s.queries.UpdateMediaFileMetadata(c.Request.Context(), catalogstore.UpdateMediaFileMetadataParams{
		ID:          pgxutil.UUID(fileID),
		Url:         url,
		MimeType:    pgtype.Text{String: meta.ContentType, Valid: true},
		SizeBytes:   pgtype.Int8{Int64: objectSize, Valid: true},
		Metadata:    metadata,
		OwnerUserID: pgxutil.UUID(owner),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update media metadata.")
		return
	}
	if s.processor != nil {
		_ = s.processor.Enqueue(c.Request.Context(), fileID)
	}
	apiresponse.OK(c, mediaJSON(updated))
}

func (s *Service) HandleListFiles(c *gin.Context) {
	owner, ok := s.ownerID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	page, perPage := parseMediaPagination(c)
	files, err := s.queries.ListMediaFiles(c.Request.Context(), catalogstore.ListMediaFilesParams{
		OwnerUserID: pgxutil.UUID(owner),
		Limit:       perPage,
		Offset:      (page - 1) * perPage,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load media files.")
		return
	}
	apiresponse.OK(c, gin.H{"items": mapMedia(files), "pagination": gin.H{"page": page, "per_page": perPage}})
}

func (s *Service) HandleGetFile(c *gin.Context) {
	owner, ok := s.ownerID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	fileID, ok := parseFileID(c)
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_file_id", "Invalid file id.")
		return
	}
	file, err := s.queries.GetMediaFileByID(c.Request.Context(), fileID)
	if err != nil {
		writeMediaNotFound(c, err)
		return
	}
	if uuidString(file.OwnerUserID) != owner.String() && !validateRoles(c) {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "You cannot access this media file.")
		return
	}
	apiresponse.OK(c, mediaJSON(file))
}

func (s *Service) HandleDeleteFile(c *gin.Context) {
	owner, ok := s.ownerID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	fileID, ok := parseFileID(c)
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_file_id", "Invalid file id.")
		return
	}
	file, err := s.queries.GetMediaFileByID(c.Request.Context(), fileID)
	if err != nil {
		writeMediaNotFound(c, err)
		return
	}
	if uuidString(file.OwnerUserID) != owner.String() && !validateRoles(c) {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "You cannot delete this media file.")
		return
	}
	if err := s.queries.DeleteMediaFile(c.Request.Context(), catalogstore.DeleteMediaFileParams{ID: fileID, OwnerUserID: file.OwnerUserID}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete media record.")
		return
	}
	if s.storage != nil && file.ObjectKey.Valid {
		_ = s.storage.Delete(c.Request.Context(), file.ObjectKey.String)
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) validatePresigned(filename, contentType string) (fileMeta, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	allowed := map[string][]string{
		"image/jpeg": {".jpg", ".jpeg"},
		"image/png":  {".png"},
		"image/webp": {".webp"},
		"image/gif":  {".gif"},
	}
	exts, ok := allowed[contentType]
	if !ok || !contains(exts, ext) {
		return fileMeta{}, errUnsupportedMedia
	}
	return fileMeta{Filename: filename, ContentType: contentType, Extension: ext, Source: "presigned", Status: "pending"}, nil
}

func (s *Service) requireMarketplaceUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !validateRoles(c) {
			apiresponse.Error(c, http.StatusForbidden, "forbidden", "Buyer, seller, or admin role is required.")
			c.Abort()
			return
		}
		c.Next()
	}
}

func writeMediaError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errUnsupportedMedia):
		apiresponse.Error(c, http.StatusBadRequest, "unsupported_media_type", "Unsupported file MIME type or extension.")
	case errors.Is(err, errTooLarge):
		apiresponse.Error(c, http.StatusBadRequest, "file_too_large", "File exceeds maximum upload size.")
	case errors.Is(err, errInvalidImage):
		apiresponse.Error(c, http.StatusBadRequest, "invalid_image", "Image dimensions are invalid.")
	default:
		apiresponse.Error(c, http.StatusBadRequest, "invalid_file", "File validation failed.")
	}
}

func writeMediaNotFound(c *gin.Context, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		apiresponse.Error(c, http.StatusNotFound, "media_file_not_found", "Media file not found.")
		return
	}
	apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load media file.")
}

func mapMedia(files []catalogstore.MediaFile) []gin.H {
	out := make([]gin.H, 0, len(files))
	for _, file := range files {
		out = append(out, mediaJSON(file))
	}
	return out
}

func decodeFileMeta(raw []byte) fileMeta {
	var meta fileMeta
	_ = json.Unmarshal(raw, &meta)
	return meta
}

func parseMediaPagination(c *gin.Context) (int32, int32) {
	page := parsePositiveInt32(c.DefaultQuery("page", "1"), 1)
	perPage := parsePositiveInt32(c.DefaultQuery("per_page", "50"), 50)
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}

func parsePositiveInt32(raw string, fallback int32) int32 {
	v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 32)
	if err != nil || v < 1 {
		return fallback
	}
	return int32(v)
}

func parseFileID(c *gin.Context) (pgtype.UUID, bool) {
	id, err := uuid.Parse(c.Param("file_id"))
	if err != nil {
		return pgtype.UUID{}, false
	}
	return pgxutil.UUID(id), true
}

func uuidFromPG(id pgtype.UUID) (uuid.UUID, bool) {
	return pgxutil.ToUUID(id)
}
