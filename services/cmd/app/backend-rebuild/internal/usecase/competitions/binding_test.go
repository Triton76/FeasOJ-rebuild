package competitions

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"testing"
	"time"
)

type fakeBindingRepo struct {
	contest          Contest
	bindings         []ContestProblemBinding
	existingProblems map[int64]bool
}

func (r *fakeBindingRepo) ListVisible(ctx context.Context, offset, limit int, visibility, ruleType, status, actorUserID, actorRole string) ([]Contest, error) {
	return []Contest{r.contest}, nil
}

func (r *fakeBindingRepo) GetVisibleByID(ctx context.Context, contestID int64, actorUserID, actorRole string) (Contest, error) {
	if r.contest.ID != contestID {
		return Contest{}, ports.ErrNotFound
	}
	return r.contest, nil
}

func (r *fakeBindingRepo) GetByID(ctx context.Context, contestID int64) (Contest, error) {
	if r.contest.ID != contestID {
		return Contest{}, ports.ErrNotFound
	}
	return r.contest, nil
}

func (r *fakeBindingRepo) Create(ctx context.Context, c Contest) (Contest, error) { return c, nil }
func (r *fakeBindingRepo) Update(ctx context.Context, c Contest) (Contest, error) { return c, nil }
func (r *fakeBindingRepo) Delete(ctx context.Context, contestID int64) error      { return nil }

func (r *fakeBindingRepo) ListProblemBindings(ctx context.Context, contestID int64) ([]ContestProblemBinding, error) {
	return r.bindings, nil
}

func (r *fakeBindingRepo) ReplaceProblemBindings(ctx context.Context, contestID int64, items []ContestProblemBinding) error {
	r.bindings = append([]ContestProblemBinding(nil), items...)
	return nil
}

func (r *fakeBindingRepo) CountProblemBindings(ctx context.Context, contestID int64) (int64, error) {
	return int64(len(r.bindings)), nil
}

func (r *fakeBindingRepo) CountExistingProblems(ctx context.Context, problemIDs []int64) (int64, error) {
	var count int64
	for _, id := range problemIDs {
		if r.existingProblems[id] {
			count++
		}
	}
	return count, nil
}

func (r *fakeBindingRepo) CreateParticipant(ctx context.Context, p Participant) (Participant, error) {
	return p, nil
}

func (r *fakeBindingRepo) ListScoreboardSubmissions(ctx context.Context, contestID int64, before time.Time) ([]ScoreboardSubmission, error) {
	return nil, nil
}

func TestReplaceContestProblemsRejectsDuplicateAlias(t *testing.T) {
	repo := &fakeBindingRepo{contest: Contest{ID: 10, OwnerUserID: "teacher-1"}, existingProblems: map[int64]bool{1: true, 2: true}}
	svc := NewService(repo)

	_, err := svc.ReplaceContestProblems(context.Background(), ports.ReplaceContestProblemsRequest{
		ContestID:   10,
		ActorUserID: "teacher-1",
		ActorRole:   "teacher",
		Items: []ports.ContestProblemBindingUpsert{
			{ProblemID: 1, DisplayOrder: 1, Alias: "A"},
			{ProblemID: 2, DisplayOrder: 2, Alias: "A"},
		},
	})
	if err != ports.ErrConflict {
		t.Fatalf("expected ErrConflict for duplicate alias, got %v", err)
	}
}

func TestReplaceContestProblemsRejectsMissingProblem(t *testing.T) {
	repo := &fakeBindingRepo{contest: Contest{ID: 10, OwnerUserID: "teacher-1"}, existingProblems: map[int64]bool{1: true}}
	svc := NewService(repo)

	_, err := svc.ReplaceContestProblems(context.Background(), ports.ReplaceContestProblemsRequest{
		ContestID:   10,
		ActorUserID: "teacher-1",
		ActorRole:   "teacher",
		Items: []ports.ContestProblemBindingUpsert{
			{ProblemID: 1, DisplayOrder: 1, Alias: "A"},
			{ProblemID: 99, DisplayOrder: 2, Alias: "B"},
		},
	})
	if err != ports.ErrNotFound {
		t.Fatalf("expected ErrNotFound for missing problem, got %v", err)
	}
}

func TestUpdateContestPublishGateRequiresBindings(t *testing.T) {
	start := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)
	repo := &fakeBindingRepo{
		contest: Contest{ID: 20, OwnerUserID: "teacher-1", Visibility: "public", Status: "draft", RuleType: "acm", Title: "C1"},
	}
	svc := NewService(repo)

	_, err := svc.UpdateContest(context.Background(), ports.UpdateContestRequest{
		ContestID:    20,
		ActorUserID:  "teacher-1",
		ActorRole:    "teacher",
		Title:        "C1",
		Subtitle:     "",
		Description:  "",
		Announcement: "",
		Visibility:   "public",
		RuleType:     "acm",
		Status:       "scheduled",
		IsEncrypted:  false,
		StartAt:      start.Format(time.RFC3339),
		EndAt:        end.Format(time.RFC3339),
	})
	if err != ports.ErrConflict {
		t.Fatalf("expected ErrConflict when publishing without bindings, got %v", err)
	}
}
