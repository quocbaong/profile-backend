package handler

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quocbaong/profile-backend/internal/config"
	"github.com/quocbaong/profile-backend/pkg/response"
)

type UploadHandler struct {
	cfg *config.Config
}

func NewUploadHandler(cfg *config.Config) *UploadHandler {
	return &UploadHandler{
		cfg: cfg,
	}
}

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
	".svg":  true,
	".pdf":  true,
	".doc":  true,
	".docx": true,
}
func (h *UploadHandler) UploadFile(c *gin.Context) {
	// 1. Kiểm tra cấu hình Supabase Storage
	if h.cfg.SupabaseURL == "" || h.cfg.SupabaseKey == "" {
		response.InternalServerError(c, "Supabase Storage chưa được cấu hình. Vui lòng kiểm tra SUPABASE_URL và SUPABASE_KEY trong .env", nil)
		return
	}

	// 2. Lấy file từ form-data (hỗ trợ cả tên field "file" và "image")
	file, err := c.FormFile("file")
	if err != nil {
		file, err = c.FormFile("image")
		if err != nil {
			response.BadRequest(c, "Vui lòng chọn file tải lên (field 'file' hoặc 'image')", err)
			return
		}
	}

	// 3. Kiểm tra dung lượng (tối đa 10MB)
	if file.Size > 10*1024*1024 {
		response.BadRequest(c, "Dung lượng ảnh vượt quá giới hạn 10MB", nil)
		return
	}

	// 4. Kiểm tra định dạng đuôi file
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		response.BadRequest(c, "Định dạng file không hỗ trợ. Cho phép: ảnh (.jpg, .png, .webp, .svg) hoặc tài liệu (.pdf, .doc, .docx)", nil)
		return
	}

	// 5. Tạo tên file duy nhất tránh trùng lặp
	timestamp := time.Now().UnixNano()
	baseName := strings.TrimSuffix(filepath.Base(file.Filename), ext)
	cleanName := strings.ReplaceAll(baseName, " ", "_")
	if len(cleanName) > 30 {
		cleanName = cleanName[:30]
	}
	newFilename := fmt.Sprintf("%s_%d%s", cleanName, timestamp, ext)

	// 6. Đẩy file trực tiếp lên Supabase Storage
	publicURL, err := h.uploadToSupabase(c.Request.Context(), file, newFilename)
	if err != nil {
		response.InternalServerError(c, fmt.Sprintf("Upload lên Supabase Storage thất bại: %v", err), err)
		return
	}

	// 7. Trả về URL CDN công khai vĩnh viễn trên Supabase
	response.Success(c, http.StatusOK, "Upload lên Supabase Storage thành công", gin.H{
		"url":       publicURL,
		"filename":  newFilename,
		"size":      file.Size,
		"provider":  "supabase",
		"mime_type": file.Header.Get("Content-Type"),
	})
}

// uploadToSupabase: Đẩy ảnh trực tiếp lên Supabase Storage bucket qua REST API
func (h *UploadHandler) uploadToSupabase(ctx context.Context, file *multipart.FileHeader, filename string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open file error: %w", err)
	}
	defer src.Close()

	bucket := h.cfg.SupabaseBucket
	if bucket == "" {
		bucket = "portfolio"
	}

	baseURL := strings.TrimRight(h.cfg.SupabaseURL, "/")
	uploadEndpoint := fmt.Sprintf("%s/storage/v1/object/%s/%s", baseURL, bucket, filename)

	req, err := http.NewRequestWithContext(ctx, "POST", uploadEndpoint, src)
	if err != nil {
		return "", fmt.Errorf("create supabase request error: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+h.cfg.SupabaseKey)
	req.Header.Set("apikey", h.cfg.SupabaseKey)
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute supabase upload error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("supabase returned HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Link công khai (Public URL) từ Supabase Storage
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", baseURL, bucket, filename)
	return publicURL, nil
}
