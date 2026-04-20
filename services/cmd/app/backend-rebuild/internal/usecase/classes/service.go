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
	ownerUserID := strings.TrimSpace(req.OwnerUserID)
	if name == "" || code == "" || ownerUserID == "" {
		return ports.ClassDTO{}, ports.ErrInvalidArgument
	}

	now := time.Now().UTC()
	item, err := s.repo.CreateClass(ctx, Class{
		ID:          uuid.NewString(),
		Name:        name,
		Code:        code,
		Description: strings.TrimSpace(req.Description),
		OwnerUserID: ownerUserID,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return ports.ClassDTO{}, err
	}
	if _, err := s.repo.CreateMembership(ctx, Membership{
		ID:          uuid.NewString(),
		ClassID:     item.ID,
		UserID:      ownerUserID,
		RoleInClass: "teacher",
		Status:      "active",
		JoinedAt:    &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
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

func (s *Service) UpdateClass(ctx context.Context, req ports.UpdateClassRequest) (ports.ClassDTO, error) {
	if s.repo == nil {
		return ports.ClassDTO{}, ports.ErrNotImplemented
	}

	classID := strings.TrimSpace(req.ClassID)
	actorUserID := strings.TrimSpace(req.ActorUserID)
	name := strings.TrimSpace(req.Name)
	if classID == "" || actorUserID == "" || name == "" {
		return ports.ClassDTO{}, ports.ErrInvalidArgument
	}

	affected, err := s.repo.UpdateClass(ctx, classID, actorUserID, name, strings.TrimSpace(req.Description), time.Now().UTC())
	if err != nil {
		return ports.ClassDTO{}, err
	}
	if !affected {
		return ports.ClassDTO{}, ports.ErrNotFound
	}

	item, err := s.repo.FindClassByID(ctx, classID)
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

func (s *Service) ArchiveClass(ctx context.Context, classID string, actorUserID string) error {
	if s.repo == nil {
		return ports.ErrNotImplemented
	}
	classID = strings.TrimSpace(classID)
	actorUserID = strings.TrimSpace(actorUserID)
	if classID == "" || actorUserID == "" {
		return ports.ErrInvalidArgument
	}

	affected, err := s.repo.ArchiveClass(ctx, classID, actorUserID, time.Now().UTC())
	if err != nil {
		return err
	}
	if !affected {
		return ports.ErrNotFound
	}
	return nil
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
	if classItem.Status == "archived" {
		return ports.ClassMembershipDTO{}, ports.ErrConflict
	}
	existingMembership, existingErr := s.repo.FindMembershipByClassAndUser(ctx, classItem.ID, userID)
	if existingErr == nil && existingMembership.ID != "" {
		return ports.ClassMembershipDTO{}, ports.ErrConflict
	}
	if existingErr != nil && !errors.Is(existingErr, ports.ErrNotFound) {
		return ports.ClassMembershipDTO{}, existingErr
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
	actorUserID := strings.TrimSpace(req.ActorUserID)
	if membershipID == "" {
		return ports.ClassMembershipDTO{}, ports.ErrInvalidArgument
	}
	if actorUserID == "" {
		return ports.ClassMembershipDTO{}, ports.ErrUnauthorized
	}

	target, err := s.repo.FindMembershipByID(ctx, membershipID)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ClassMembershipDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.ClassMembershipDTO{}, err
	}
	if target.Status != "pending" {
		return ports.ClassMembershipDTO{}, ports.ErrConflict
	}

	if ok, err := s.canManageClass(ctx, target.ClassID, actorUserID); err != nil {
		return ports.ClassMembershipDTO{}, err
	} else if !ok {
		return ports.ClassMembershipDTO{}, ports.ErrForbidden
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
		return ports.ClassMembershipDTO{}, ports.ErrConflict
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

func (s *Service) ListClassMemberships(ctx context.Context, classID string, actorUserID string) ([]ports.ClassMembershipDTO, error) {
	if s.repo == nil {
		return nil, ports.ErrNotImplemented
	}
	classID = strings.TrimSpace(classID)
	actorUserID = strings.TrimSpace(actorUserID)
	if classID == "" || actorUserID == "" {
		return nil, ports.ErrInvalidArgument
	}

	ok, err := s.canManageClass(ctx, classID, actorUserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ports.ErrForbidden
	}

	items, err := s.repo.ListMembershipsByClassID(ctx, classID)
	if err != nil {
		return nil, err
	}
	resp := make([]ports.ClassMembershipDTO, 0, len(items))
	for _, item := range items {
		resp = append(resp, toMembershipDTO(item))
	}
	return resp, nil
}

func (s *Service) ListMyMemberships(ctx context.Context, userID string) ([]ports.ClassMembershipDTO, error) {
	if s.repo == nil {
		return nil, ports.ErrNotImplemented
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ports.ErrInvalidArgument
	}

	items, err := s.repo.ListMembershipsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := make([]ports.ClassMembershipDTO, 0, len(items))
	for _, item := range items {
		resp = append(resp, toMembershipDTO(item))
	}
	return resp, nil
}

func (s *Service) canManageClass(ctx context.Context, classID, actorUserID string) (bool, error) {
	classItem, err := s.repo.FindClassByID(ctx, classID)
	if errors.Is(err, ports.ErrNotFound) {
		return false, ports.ErrNotFound
	}
	if err != nil {
		return false, err
	}
	if classItem.OwnerUserID == actorUserID {
		return true, nil
	}
	membership, err := s.repo.FindMembershipByClassAndUser(ctx, classID, actorUserID)
	if errors.Is(err, ports.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if membership.Status != "active" {
		return false, nil
	}
	return membership.RoleInClass == "teacher" || membership.RoleInClass == "assistant", nil
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
