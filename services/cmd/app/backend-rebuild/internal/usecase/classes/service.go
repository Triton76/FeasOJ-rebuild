package classes

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateClass(ctx context.Context, req ports.CreateClassRequest) (ports.ClassDTO, error) {
	if s.repo == nil {
		return ports.ClassDTO{}, ports.ErrNotImplemented
	}
	name := strings.TrimSpace(req.Name)
	code := strings.TrimSpace(req.Code)
	if name == "" || code == "" {
		return ports.ClassDTO{}, ports.ErrInvalidArgument
	}

	now := time.Now().UTC()
	item, err := s.repo.CreateClass(ctx, Class{
		ID:          uuid.NewString(),
		Name:        name,
		Code:        code,
		Description: strings.TrimSpace(req.Description),
		OwnerUserID: "",
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return ports.ClassDTO{}, err
	}

	return ports.ClassDTO{
		ID:          item.ID,
		Name:        item.Name,
		Code:        item.Code,
		Description: item.Description,
		OwnerUserID: item.OwnerUserID,
		Status:      item.Status,
	}, nil
}

func (s *Service) ApplyJoinClass(ctx context.Context, req ports.ApplyJoinClassRequest) (ports.ClassMembershipDTO, error) {
	if s.repo == nil {
		return ports.ClassMembershipDTO{}, ports.ErrNotImplemented
	}
	classCode := strings.TrimSpace(req.ClassCode)
	userID := strings.TrimSpace(req.UserID)
	if classCode == "" || userID == "" {
		return ports.ClassMembershipDTO{}, ports.ErrInvalidArgument
	}

	classItem, err := s.repo.FindClassByCode(ctx, classCode)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ClassMembershipDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.ClassMembershipDTO{}, err
	}

	now := time.Now().UTC()
	m, err := s.repo.CreateMembership(ctx, Membership{
		ID:          uuid.NewString(),
		ClassID:     classItem.ID,
		UserID:      userID,
		RoleInClass: "student",
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return ports.ClassMembershipDTO{}, err
	}

	return toMembershipDTO(m), nil
}

func (s *Service) ReviewMembership(ctx context.Context, req ports.ReviewMembershipRequest) (ports.ClassMembershipDTO, error) {
	if s.repo == nil {
		return ports.ClassMembershipDTO{}, ports.ErrNotImplemented
	}
	membershipID := strings.TrimSpace(req.MembershipID)
	if membershipID == "" {
		return ports.ClassMembershipDTO{}, ports.ErrInvalidArgument
	}

	status := "rejected"
	var joinedAt *time.Time
	now := time.Now().UTC()
	if req.Approve {
		status = "active"
		joinedAt = &now
	}

	affected, err := s.repo.UpdateMembershipStatus(ctx, membershipID, status, joinedAt, now)
	if err != nil {
		return ports.ClassMembershipDTO{}, err
	}
	if !affected {
		return ports.ClassMembershipDTO{}, ports.ErrNotFound
	}

	m, err := s.repo.FindMembershipByID(ctx, membershipID)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ClassMembershipDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.ClassMembershipDTO{}, err
	}

	return toMembershipDTO(m), nil
}

func toMembershipDTO(m Membership) ports.ClassMembershipDTO {
	return ports.ClassMembershipDTO{
		ID:          m.ID,
		ClassID:     m.ClassID,
		UserID:      m.UserID,
		RoleInClass: m.RoleInClass,
		Status:      m.Status,
	}
}
