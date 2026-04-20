package discussions

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) ListDiscussions(ctx context.Context, req ports.DiscussionsQuery) ([]ports.DiscussionDTO, error) {
	if s.repo == nil {
		return nil, ports.ErrNotImplemented
	}
	page := req.Page
	if page <= 0 {
		page = defaultPage
	}
	limit := req.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	items, err := s.repo.List(ctx, (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}
	resp := make([]ports.DiscussionDTO, 0, len(items))
	for _, item := range items {
		resp = append(resp, toDiscussionDTO(item))
	}
	return resp, nil
}

func (s *Service) GetDiscussion(ctx context.Context, discussionID string) (ports.DiscussionDTO, error) {
	if s.repo == nil {
		return ports.DiscussionDTO{}, ports.ErrNotImplemented
	}
	discussionID = strings.TrimSpace(discussionID)
	if discussionID == "" {
		return ports.DiscussionDTO{}, ports.ErrInvalidArgument
	}
	d, err := s.repo.GetByID(ctx, discussionID)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.DiscussionDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.DiscussionDTO{}, err
	}
	return toDiscussionDTO(d), nil
}

func (s *Service) CreateDiscussion(ctx context.Context, req ports.CreateDiscussionRequest) (ports.DiscussionDTO, error) {
	if s.repo == nil {
		return ports.DiscussionDTO{}, ports.ErrNotImplemented
	}
	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	userID := strings.TrimSpace(req.UserID)
	if title == "" || content == "" || userID == "" {
		return ports.DiscussionDTO{}, ports.ErrInvalidArgument
	}
	now := time.Now().UTC()
	d, err := s.repo.CreateDiscussion(ctx, Discussion{ID: uuid.NewString(), Title: title, Content: content, UserID: userID, CreatedAt: now})
	if err != nil {
		return ports.DiscussionDTO{}, err
	}
	return toDiscussionDTO(d), nil
}

func (s *Service) CreateComment(ctx context.Context, req ports.CreateCommentRequest) (ports.CommentDTO, error) {
	if s.repo == nil {
		return ports.CommentDTO{}, ports.ErrNotImplemented
	}
	discussionID := strings.TrimSpace(req.DiscussionID)
	content := strings.TrimSpace(req.Content)
	userID := strings.TrimSpace(req.UserID)
	if discussionID == "" || content == "" || userID == "" {
		return ports.CommentDTO{}, ports.ErrInvalidArgument
	}
	now := time.Now().UTC()
	c, err := s.repo.CreateComment(ctx, Comment{ID: uuid.NewString(), DiscussionID: discussionID, Content: content, UserID: userID, Profanity: false, CreatedAt: now})
	if errors.Is(err, ports.ErrNotFound) {
		return ports.CommentDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.CommentDTO{}, err
	}
	return toCommentDTO(c), nil
}

func (s *Service) ListComments(ctx context.Context, req ports.CommentsQuery) ([]ports.CommentDTO, error) {
	if s.repo == nil {
		return nil, ports.ErrNotImplemented
	}
	discussionID := strings.TrimSpace(req.DiscussionID)
	if discussionID == "" {
		return nil, ports.ErrInvalidArgument
	}
	page := req.Page
	if page <= 0 {
		page = defaultPage
	}
	limit := req.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	if _, err := s.repo.GetByID(ctx, discussionID); errors.Is(err, ports.ErrNotFound) {
		return nil, ports.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	items, err := s.repo.ListCommentsByDiscussionID(ctx, discussionID, (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}
	resp := make([]ports.CommentDTO, 0, len(items))
	for _, item := range items {
		resp = append(resp, toCommentDTO(item))
	}
	return resp, nil
}

func (s *Service) DeleteDiscussion(ctx context.Context, req ports.DeleteDiscussionRequest) error {
	if s.repo == nil {
		return ports.ErrNotImplemented
	}
	discussionID := strings.TrimSpace(req.DiscussionID)
	actorUserID := strings.TrimSpace(req.ActorUserID)
	actorRole := strings.TrimSpace(req.ActorRole)
	if discussionID == "" || actorUserID == "" {
		return ports.ErrInvalidArgument
	}

	d, err := s.repo.GetByID(ctx, discussionID)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ErrNotFound
	}
	if err != nil {
		return err
	}
	if actorRole != "admin" && strings.TrimSpace(d.UserID) != actorUserID {
		return ports.ErrForbidden
	}
	return s.repo.DeleteDiscussion(ctx, discussionID)
}

func (s *Service) DeleteComment(ctx context.Context, req ports.DeleteCommentRequest) error {
	if s.repo == nil {
		return ports.ErrNotImplemented
	}
	commentID := strings.TrimSpace(req.CommentID)
	actorUserID := strings.TrimSpace(req.ActorUserID)
	actorRole := strings.TrimSpace(req.ActorRole)
	if commentID == "" || actorUserID == "" {
		return ports.ErrInvalidArgument
	}

	c, err := s.repo.GetCommentByID(ctx, commentID)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ErrNotFound
	}
	if err != nil {
		return err
	}
	if actorRole != "admin" && strings.TrimSpace(c.UserID) != actorUserID {
		return ports.ErrForbidden
	}
	return s.repo.DeleteComment(ctx, commentID)
}

func toDiscussionDTO(d Discussion) ports.DiscussionDTO {
	return ports.DiscussionDTO{ID: d.ID, Title: d.Title, Content: d.Content, UserID: d.UserID, Username: d.Username, Avatar: d.Avatar, CreatedAt: d.CreatedAt.UTC().Format(time.RFC3339)}
}

func toCommentDTO(c Comment) ports.CommentDTO {
	return ports.CommentDTO{ID: c.ID, DiscussionID: c.DiscussionID, Content: c.Content, UserID: c.UserID, Username: c.Username, Avatar: c.Avatar, Profanity: c.Profanity, CreatedAt: c.CreatedAt.UTC().Format(time.RFC3339)}
}
