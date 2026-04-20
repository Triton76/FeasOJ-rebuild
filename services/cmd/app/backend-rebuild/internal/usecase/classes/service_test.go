package classes

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeClassRepo struct {
	mu          sync.Mutex
	classes     map[string]Class
	byCode      map[string]string
	memberships map[string]Membership
}

func newFakeClassRepo() *fakeClassRepo {
	return &fakeClassRepo{
		classes:     make(map[string]Class),
		byCode:      make(map[string]string),
		memberships: make(map[string]Membership),
	}
}

func (r *fakeClassRepo) CreateClass(ctx context.Context, c Class) (Class, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byCode[c.Code]; ok {
		return Class{}, ports.ErrConflict
	}
	r.classes[c.ID] = c
	r.byCode[c.Code] = c.ID
	return c, nil
}

func (r *fakeClassRepo) FindClassByID(ctx context.Context, classID string) (Class, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.classes[classID]
	if !ok {
		return Class{}, ports.ErrNotFound
	}
	return item, nil
}

func (r *fakeClassRepo) CreateMembership(ctx context.Context, m Membership) (Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := m.ClassID + ":" + m.UserID
	if _, ok := r.memberships[key]; ok {
		return Membership{}, ports.ErrConflict
	}
	r.memberships[key] = m
	return m, nil
}

func (r *fakeClassRepo) FindClassByCode(ctx context.Context, code string) (Class, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	classID, ok := r.byCode[code]
	if !ok {
		return Class{}, ports.ErrNotFound
	}
	return r.classes[classID], nil
}

func (r *fakeClassRepo) FindMembershipByID(ctx context.Context, membershipID string) (Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, item := range r.memberships {
		if item.ID == membershipID {
			return item, nil
		}
	}
	return Membership{}, ports.ErrNotFound
}

func (r *fakeClassRepo) FindMembershipByClassAndUser(ctx context.Context, classID, userID string) (Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.memberships[classID+":"+userID]
	if !ok {
		return Membership{}, ports.ErrNotFound
	}
	return item, nil
}

func (r *fakeClassRepo) ListMembershipsByClassID(ctx context.Context, classID string) ([]Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]Membership, 0)
	for _, item := range r.memberships {
		if item.ClassID == classID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *fakeClassRepo) ListMembershipsByUserID(ctx context.Context, userID string) ([]Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]Membership, 0)
	for _, item := range r.memberships {
		if item.UserID == userID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *fakeClassRepo) UpdateClass(ctx context.Context, classID, ownerUserID, name, description string, updatedAt time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.classes[classID]
	if !ok || item.OwnerUserID != ownerUserID || item.Status == "archived" {
		return false, nil
	}
	item.Name = name
	item.Description = description
	item.UpdatedAt = updatedAt
	r.classes[classID] = item
	return true, nil
}

func (r *fakeClassRepo) ArchiveClass(ctx context.Context, classID, ownerUserID string, updatedAt time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.classes[classID]
	if !ok || item.OwnerUserID != ownerUserID || item.Status == "archived" {
		return false, nil
	}
	item.Status = "archived"
	item.UpdatedAt = updatedAt
	r.classes[classID] = item
	return true, nil
}

func (r *fakeClassRepo) UpdateMembershipStatus(ctx context.Context, membershipID, status string, joinedAt *time.Time, updatedAt time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key, item := range r.memberships {
		if item.ID != membershipID {
			continue
		}
		item.Status = status
		item.JoinedAt = joinedAt
		item.UpdatedAt = updatedAt
		r.memberships[key] = item
		return true, nil
	}
	return false, nil
}

func TestClassWorkflowCreatesTeacherMembership(t *testing.T) {
	repo := newFakeClassRepo()
	svc := NewService(repo)

	created, err := svc.CreateClass(context.Background(), ports.CreateClassRequest{
		Name:        "Math",
		Code:        "MATH101",
		Description: "basic",
		OwnerUserID: "teacher-1",
	})
	if err != nil {
		t.Fatalf("CreateClass failed: %v", err)
	}
	if created.OwnerUserID != "teacher-1" {
		t.Fatalf("unexpected owner: %s", created.OwnerUserID)
	}

	items, err := svc.ListMyMemberships(context.Background(), "teacher-1")
	if err != nil {
		t.Fatalf("ListMyMemberships failed: %v", err)
	}
	if len(items) != 1 || items[0].RoleInClass != "teacher" || items[0].Status != "active" {
		t.Fatalf("expected active teacher membership, got %+v", items)
	}
}

func TestClassWorkflowApplyReviewAndArchive(t *testing.T) {
	repo := newFakeClassRepo()
	svc := NewService(repo)

	created, err := svc.CreateClass(context.Background(), ports.CreateClassRequest{
		Name:        "Physics",
		Code:        "PHY101",
		Description: "basic",
		OwnerUserID: "teacher-1",
	})
	if err != nil {
		t.Fatalf("CreateClass failed: %v", err)
	}

	joined, err := svc.ApplyJoinClass(context.Background(), ports.ApplyJoinClassRequest{ClassCode: created.Code, UserID: "student-1"})
	if err != nil {
		t.Fatalf("ApplyJoinClass failed: %v", err)
	}
	if joined.Status != "pending" {
		t.Fatalf("expected pending membership, got %s", joined.Status)
	}
	if _, err := svc.ApplyJoinClass(context.Background(), ports.ApplyJoinClassRequest{ClassCode: created.Code, UserID: "student-1"}); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected duplicate apply conflict, got %v", err)
	}

	list, err := svc.ListClassMemberships(context.Background(), created.ID, "teacher-1")
	if err != nil {
		t.Fatalf("ListClassMemberships failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected teacher + student memberships, got %d", len(list))
	}

	updated, err := svc.ReviewMembership(context.Background(), ports.ReviewMembershipRequest{MembershipID: joined.ID, Approve: true, ActorUserID: "teacher-1"})
	if err != nil {
		t.Fatalf("ReviewMembership approve failed: %v", err)
	}
	if updated.Status != "active" {
		t.Fatalf("expected active membership, got %s", updated.Status)
	}
	if _, err := svc.ReviewMembership(context.Background(), ports.ReviewMembershipRequest{MembershipID: joined.ID, Approve: false, ActorUserID: "teacher-1"}); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected duplicate review conflict, got %v", err)
	}

	if err := svc.ArchiveClass(context.Background(), created.ID, "teacher-1"); err != nil {
		t.Fatalf("ArchiveClass failed: %v", err)
	}
	if _, err := svc.ApplyJoinClass(context.Background(), ports.ApplyJoinClassRequest{ClassCode: created.Code, UserID: "student-2"}); !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected archived class apply conflict, got %v", err)
	}
}
