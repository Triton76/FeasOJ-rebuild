package router

import (
	"FeasOJ/app/backend-rebuild/internal/config"
	"FeasOJ/app/backend-rebuild/internal/http/handler"
	"FeasOJ/app/backend-rebuild/internal/ports"
	avatarstorage "FeasOJ/app/backend-rebuild/internal/storage/avatar"
	usersusecase "FeasOJ/app/backend-rebuild/internal/usecase/users"
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type avatarTestRepo struct {
	user usersusecase.User
}

func (r *avatarTestRepo) GetByID(ctx context.Context, userID string) (usersusecase.User, error) {
	if r.user.ID != userID {
		return usersusecase.User{}, ports.ErrNotFound
	}
	return r.user, nil
}

func (r *avatarTestRepo) UpdateProfile(ctx context.Context, userID string, avatar, synopsis *string, updatedAt time.Time) (bool, error) {
	if r.user.ID != userID {
		return false, ports.ErrNotFound
	}
	if avatar != nil {
		r.user.Avatar = *avatar
	}
	if synopsis != nil {
		r.user.Synopsis = *synopsis
	}
	r.user.UpdatedAt = updatedAt
	return true, nil
}

func (r *avatarTestRepo) ListRanking(ctx context.Context, offset, limit int) ([]usersusecase.User, error) {
	return []usersusecase.User{r.user}, nil
}

func TestAvatarUploadValidationAndProfileUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	avatarDir := t.TempDir()
	repo := &avatarTestRepo{
		user: usersusecase.User{
			ID:       "u-1",
			Username: "alice",
			Email:    "alice@example.com",
			Avatar:   "existing.png",
			Synopsis: "before",
			Role:     "student",
			Status:   "active",
		},
	}
	store, err := avatarstorage.NewLocalStorage(avatarDir)
	if err != nil {
		t.Fatalf("new local storage failed: %v", err)
	}

	usersSvc := usersusecase.NewService(repo, usersusecase.Options{
		AvatarStorage:        store,
		AvatarUploadEnabled:  true,
		AvatarUploadMaxBytes: 256,
	})

	h := handler.New(handler.Handlers{
		Auth:                 fakeAuthService{},
		Users:                usersSvc,
		AvatarUploadMaxBytes: 256,
	})
	r := gin.New()
	r.Use(gin.Recovery())
	Register(r, h, config.Config{
		JWTSecret:            "test-secret",
		EnableAvatarUpload:   true,
		AvatarUploadDir:      avatarDir,
		AvatarUploadMaxBytes: 256,
	})

	token := loginToken(t, r, "alice")

	t.Run("rejects invalid content type", func(t *testing.T) {
		req := newAvatarUploadRequest(t, []byte("plain text avatar"), "avatar.txt", token)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "not allowed") {
			t.Fatalf("expected invalid type message, got %s", rec.Body.String())
		}
		if repo.user.Avatar != "existing.png" {
			t.Fatalf("avatar should not mutate on invalid upload, got %s", repo.user.Avatar)
		}
	})

	t.Run("rejects oversized payload", func(t *testing.T) {
		req := newAvatarUploadRequest(t, bytes.Repeat([]byte("a"), 300), "avatar.png", token)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "exceeds max size") {
			t.Fatalf("expected max size message, got %s", rec.Body.String())
		}
		if repo.user.Avatar != "existing.png" {
			t.Fatalf("avatar should not mutate on oversized upload, got %s", repo.user.Avatar)
		}
	})

	t.Run("uploads avatar and patches profile reference", func(t *testing.T) {
		req := newAvatarUploadRequest(t, tinyPNG, "avatar.png", token)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
		}

		var uploadResp struct {
			Data ports.AvatarUploadResponse `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &uploadResp); err != nil {
			t.Fatalf("decode upload response failed: %v", err)
		}
		if uploadResp.Data.Avatar == "" {
			t.Fatal("expected non-empty avatar reference")
		}
		if uploadResp.Data.ContentType != "image/png" {
			t.Fatalf("unexpected content type: %s", uploadResp.Data.ContentType)
		}
		if _, err := os.Stat(filepath.Join(avatarDir, uploadResp.Data.Avatar)); err != nil {
			t.Fatalf("expected stored avatar file, got error: %v", err)
		}

		patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/profile", strings.NewReader(`{"avatar":"`+uploadResp.Data.Avatar+`"}`))
		patchReq.Header.Set("Content-Type", "application/json")
		patchReq.Header.Set("Authorization", "Bearer "+token)
		patchRec := httptest.NewRecorder()
		r.ServeHTTP(patchRec, patchReq)
		if patchRec.Code != http.StatusOK {
			t.Fatalf("patch avatar failed status=%d body=%s", patchRec.Code, patchRec.Body.String())
		}
		if repo.user.Avatar != uploadResp.Data.Avatar {
			t.Fatalf("expected profile avatar %s, got %s", uploadResp.Data.Avatar, repo.user.Avatar)
		}

		synopsisReq := httptest.NewRequest(http.MethodPatch, "/api/v1/profile", strings.NewReader(`{"synopsis":"after"}`))
		synopsisReq.Header.Set("Content-Type", "application/json")
		synopsisReq.Header.Set("Authorization", "Bearer "+token)
		synopsisRec := httptest.NewRecorder()
		r.ServeHTTP(synopsisRec, synopsisReq)
		if synopsisRec.Code != http.StatusOK {
			t.Fatalf("patch synopsis failed status=%d body=%s", synopsisRec.Code, synopsisRec.Body.String())
		}
		if repo.user.Avatar != uploadResp.Data.Avatar {
			t.Fatalf("avatar should be preserved when patching synopsis, got %s", repo.user.Avatar)
		}

		getReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/u-1", nil)
		getReq.Header.Set("Authorization", "Bearer "+token)
		getRec := httptest.NewRecorder()
		r.ServeHTTP(getRec, getReq)
		if getRec.Code != http.StatusOK {
			t.Fatalf("get profile failed status=%d body=%s", getRec.Code, getRec.Body.String())
		}
		if !strings.Contains(getRec.Body.String(), uploadResp.Data.Avatar) || !strings.Contains(getRec.Body.String(), `"synopsis":"after"`) {
			t.Fatalf("unexpected profile payload: %s", getRec.Body.String())
		}
	})
}

func newAvatarUploadRequest(t *testing.T, content []byte, filename, token string) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("avatar", filename)
	if err != nil {
		t.Fatalf("create form file failed: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write form file failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/profile/avatar/upload", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

var tinyPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
	0x00, 0x03, 0x01, 0x01, 0x00, 0xc9, 0xfe, 0x92,
	0xef, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
	0x44, 0xae, 0x42, 0x60, 0x82,
}
