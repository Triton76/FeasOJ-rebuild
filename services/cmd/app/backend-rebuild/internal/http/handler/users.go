// Layer: Handler (HTTP接入层) - Users模块
// Responsibility: 用户相关HTTP端点处理(获取资料/更新资料/排行榜)
package handler

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const avatarMultipartOverheadBytes int64 = 64 * 1024

func (h Handlers) GetProfile(c *gin.Context) {
	resp, err := h.Users.GetProfile(c.Request.Context(), c.Param("user_id"))
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) UpdateProfile(c *gin.Context) {
	var req ports.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("auth_user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	resp, err := h.Users.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) UploadAvatar(c *gin.Context) {
	userID := c.GetString("auth_user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	limit := h.AvatarUploadMaxBytes + avatarMultipartOverheadBytes
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		msg := "avatar file is required"
		if strings.Contains(err.Error(), "request body too large") {
			msg = fmt.Sprintf("avatar file exceeds max size %d bytes", h.AvatarUploadMaxBytes)
		}
		fail(c, fmt.Errorf("%s: %w", msg, ports.ErrInvalidArgument))
		return
	}
	if fileHeader.Size > 0 && fileHeader.Size > h.AvatarUploadMaxBytes {
		fail(c, fmt.Errorf("avatar file exceeds max size %d bytes: %w", h.AvatarUploadMaxBytes, ports.ErrInvalidArgument))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		fail(c, fmt.Errorf("open avatar file failed: %w", err))
		return
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, h.AvatarUploadMaxBytes+1))
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			fail(c, fmt.Errorf("avatar file exceeds max size %d bytes: %w", h.AvatarUploadMaxBytes, ports.ErrInvalidArgument))
			return
		}
		fail(c, fmt.Errorf("read avatar file failed: %w", err))
		return
	}
	if int64(len(content)) > h.AvatarUploadMaxBytes {
		fail(c, fmt.Errorf("avatar file exceeds max size %d bytes: %w", h.AvatarUploadMaxBytes, ports.ErrInvalidArgument))
		return
	}

	resp, err := h.Users.UploadAvatar(c.Request.Context(), userID, ports.UploadAvatarRequest{
		Filename:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Content:     content,
	})
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) ListRanking(c *gin.Context) {
	var req ports.RankingQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Users.ListRanking(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}
