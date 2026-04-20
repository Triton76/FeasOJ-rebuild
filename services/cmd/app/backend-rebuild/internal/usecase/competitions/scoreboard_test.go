package competitions

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	passwordutil "FeasOJ/pkg/auth"
	"context"
	"testing"
	"time"
)

type fakeScoreboardRepo struct {
	contest     Contest
	submissions []ScoreboardSubmission
}

func (r *fakeScoreboardRepo) ListVisible(ctx context.Context, offset, limit int, visibility, ruleType, status, actorUserID, actorRole string) ([]Contest, error) {
	return []Contest{r.contest}, nil
}

func (r *fakeScoreboardRepo) GetVisibleByID(ctx context.Context, contestID int64, actorUserID, actorRole string) (Contest, error) {
	if r.contest.ID != contestID {
		return Contest{}, ports.ErrNotFound
	}
	return r.contest, nil
}

func (r *fakeScoreboardRepo) GetByID(ctx context.Context, contestID int64) (Contest, error) {
	if r.contest.ID != contestID {
		return Contest{}, ports.ErrNotFound
	}
	return r.contest, nil
}

func (r *fakeScoreboardRepo) Create(ctx context.Context, c Contest) (Contest, error) { return c, nil }
func (r *fakeScoreboardRepo) Update(ctx context.Context, c Contest) (Contest, error) { return c, nil }
func (r *fakeScoreboardRepo) Delete(ctx context.Context, contestID int64) error      { return nil }

func (r *fakeScoreboardRepo) CreateParticipant(ctx context.Context, p Participant) (Participant, error) {
	return p, nil
}

func (r *fakeScoreboardRepo) ListScoreboardSubmissions(ctx context.Context, contestID int64, before time.Time) ([]ScoreboardSubmission, error) {
	resp := make([]ScoreboardSubmission, 0, len(r.submissions))
	for _, item := range r.submissions {
		if item.SubmittedAt.Before(before) {
			resp = append(resp, item)
		}
	}
	return resp, nil
}

func TestScoreboardSortBySolvedPenaltyAndReachedAt(t *testing.T) {
	start := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)
	repo := &fakeScoreboardRepo{
		contest: Contest{ID: 1, StartAt: &start, EndAt: &end, RuleType: "acm"},
		submissions: []ScoreboardSubmission{
			{UserID: "u1", Username: "alice", ProblemID: 1, Result: "accepted", Score: 100, SubmittedAt: start.Add(20 * time.Minute)},
			{UserID: "u1", Username: "alice", ProblemID: 2, Result: "accepted", Score: 100, SubmittedAt: start.Add(40 * time.Minute)},
			{UserID: "u2", Username: "bob", ProblemID: 1, Result: "accepted", Score: 100, SubmittedAt: start.Add(20 * time.Minute)},
			{UserID: "u2", Username: "bob", ProblemID: 2, Result: "accepted", Score: 100, SubmittedAt: start.Add(50 * time.Minute)},
		},
	}

	svc := NewService(repo)
	svc.nowFn = func() time.Time { return start.Add(70 * time.Minute) }

	resp, err := svc.GetScoreboard(context.Background(), ports.ContestScoreboardQuery{ContestID: 1})
	if err != nil {
		t.Fatalf("GetScoreboard failed: %v", err)
	}
	if len(resp.VisibleItems) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(resp.VisibleItems))
	}
	if resp.VisibleItems[0].UserID != "u1" {
		t.Fatalf("expected u1 ranked first by earlier reached_at tie breaker, got %s", resp.VisibleItems[0].UserID)
	}
}

func TestScoreboardCompileErrorPenaltyOnlyWhenSolved(t *testing.T) {
	start := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	repo := &fakeScoreboardRepo{
		contest: Contest{ID: 2, StartAt: &start, RuleType: "acm"},
		submissions: []ScoreboardSubmission{
			{UserID: "u1", Username: "alice", ProblemID: 1, Result: "compile_error", Score: 0, SubmittedAt: start.Add(5 * time.Minute)},
			{UserID: "u1", Username: "alice", ProblemID: 1, Result: "accepted", Score: 100, SubmittedAt: start.Add(20 * time.Minute)},
			{UserID: "u2", Username: "bob", ProblemID: 1, Result: "compile_error", Score: 0, SubmittedAt: start.Add(5 * time.Minute)},
			{UserID: "u2", Username: "bob", ProblemID: 1, Result: "wrong_answer", Score: 0, SubmittedAt: start.Add(10 * time.Minute)},
		},
	}

	svc := NewService(repo)
	svc.nowFn = func() time.Time { return start.Add(30 * time.Minute) }

	resp, err := svc.GetScoreboard(context.Background(), ports.ContestScoreboardQuery{ContestID: 2})
	if err != nil {
		t.Fatalf("GetScoreboard failed: %v", err)
	}
	if len(resp.VisibleItems) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(resp.VisibleItems))
	}
	if resp.VisibleItems[0].UserID != "u1" || resp.VisibleItems[0].PenaltyMinutes != 40 {
		t.Fatalf("expected u1 solved with CE penalty=40, got user=%s penalty=%d", resp.VisibleItems[0].UserID, resp.VisibleItems[0].PenaltyMinutes)
	}
	if resp.VisibleItems[1].UserID != "u2" || resp.VisibleItems[1].PenaltyMinutes != 0 {
		t.Fatalf("expected unsolved CE to contribute no penalty, got user=%s penalty=%d", resp.VisibleItems[1].UserID, resp.VisibleItems[1].PenaltyMinutes)
	}
}

func TestScoreboardFreezeHidesLastSixtyMinutesBoundary(t *testing.T) {
	start := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)
	repo := &fakeScoreboardRepo{
		contest: Contest{ID: 3, StartAt: &start, EndAt: &end, RuleType: "acm"},
		submissions: []ScoreboardSubmission{
			{UserID: "u1", Username: "alice", ProblemID: 1, Result: "accepted", Score: 100, SubmittedAt: start.Add(50 * time.Minute)},
			{UserID: "u1", Username: "alice", ProblemID: 2, Result: "accepted", Score: 100, SubmittedAt: start.Add(60 * time.Minute)},
			{UserID: "u2", Username: "bob", ProblemID: 1, Result: "accepted", Score: 100, SubmittedAt: start.Add(70 * time.Minute)},
		},
	}

	svc := NewService(repo)
	svc.nowFn = func() time.Time { return start.Add(90 * time.Minute) }

	resp, err := svc.GetScoreboard(context.Background(), ports.ContestScoreboardQuery{ContestID: 3})
	if err != nil {
		t.Fatalf("GetScoreboard failed: %v", err)
	}
	if !resp.FreezeActive {
		t.Fatal("expected freeze_active=true within final 60 minutes")
	}
	if resp.FreezeStartAt != start.Add(60*time.Minute).UTC().Format(time.RFC3339) {
		t.Fatalf("unexpected freeze_start_at: %s", resp.FreezeStartAt)
	}
	if len(resp.VisibleItems) != 1 {
		t.Fatalf("expected only pre-freeze solved data visible, got %d rows", len(resp.VisibleItems))
	}
	if resp.VisibleItems[0].UserID != "u1" || resp.VisibleItems[0].Solved != 1 {
		t.Fatalf("expected only first pre-freeze acceptance visible, got user=%s solved=%d", resp.VisibleItems[0].UserID, resp.VisibleItems[0].Solved)
	}
}

func TestScoreboardOIUsesPerProblemBestScoreAggregation(t *testing.T) {
	start := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	repo := &fakeScoreboardRepo{
		contest: Contest{ID: 4, StartAt: &start, RuleType: "oi"},
		submissions: []ScoreboardSubmission{
			{UserID: "u1", Username: "alice", ProblemID: 1, Result: "wrong_answer", Score: 20, SubmittedAt: start.Add(5 * time.Minute)},
			{UserID: "u1", Username: "alice", ProblemID: 1, Result: "accepted", Score: 60, SubmittedAt: start.Add(15 * time.Minute)},
			{UserID: "u1", Username: "alice", ProblemID: 2, Result: "accepted", Score: 90, SubmittedAt: start.Add(25 * time.Minute)},
			{UserID: "u2", Username: "bob", ProblemID: 1, Result: "accepted", Score: 70, SubmittedAt: start.Add(10 * time.Minute)},
			{UserID: "u2", Username: "bob", ProblemID: 2, Result: "accepted", Score: 60, SubmittedAt: start.Add(20 * time.Minute)},
		},
	}

	svc := NewService(repo)
	svc.nowFn = func() time.Time { return start.Add(40 * time.Minute) }

	resp, err := svc.GetScoreboard(context.Background(), ports.ContestScoreboardQuery{ContestID: 4})
	if err != nil {
		t.Fatalf("GetScoreboard failed: %v", err)
	}
	if len(resp.VisibleItems) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(resp.VisibleItems))
	}
	if resp.VisibleItems[0].UserID != "u1" || resp.VisibleItems[0].TotalScore != 150 {
		t.Fatalf("expected u1 total_score=150, got user=%s score=%d", resp.VisibleItems[0].UserID, resp.VisibleItems[0].TotalScore)
	}
	if resp.VisibleItems[1].UserID != "u2" || resp.VisibleItems[1].TotalScore != 130 {
		t.Fatalf("expected u2 total_score=130, got user=%s score=%d", resp.VisibleItems[1].UserID, resp.VisibleItems[1].TotalScore)
	}
}

func TestScoreboardOITieBreakUsesEarlierReachedAtThenUserID(t *testing.T) {
	start := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	repo := &fakeScoreboardRepo{
		contest: Contest{ID: 5, StartAt: &start, RuleType: "oi"},
		submissions: []ScoreboardSubmission{
			{UserID: "u2", Username: "bob", ProblemID: 1, Result: "accepted", Score: 100, SubmittedAt: start.Add(20 * time.Minute)},
			{UserID: "u1", Username: "alice", ProblemID: 1, Result: "accepted", Score: 100, SubmittedAt: start.Add(15 * time.Minute)},
		},
	}

	svc := NewService(repo)
	svc.nowFn = func() time.Time { return start.Add(30 * time.Minute) }

	resp, err := svc.GetScoreboard(context.Background(), ports.ContestScoreboardQuery{ContestID: 5})
	if err != nil {
		t.Fatalf("GetScoreboard failed: %v", err)
	}
	if len(resp.VisibleItems) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(resp.VisibleItems))
	}
	if resp.VisibleItems[0].UserID != "u1" {
		t.Fatalf("expected u1 first by earlier reached_at, got %s", resp.VisibleItems[0].UserID)
	}
}

func TestJoinContestWrongPasswordForbiddenAcrossRules(t *testing.T) {
	hash := passwordutil.EncryptPassword("correct-password")
	rules := []string{"acm", "oi", "assignment"}

	for _, rule := range rules {
		repo := &fakeScoreboardRepo{
			contest: Contest{ID: 100, RuleType: rule, IsEncrypted: true, PasswordHash: hash},
		}
		svc := NewService(repo)

		_, err := svc.JoinContest(context.Background(), ports.JoinContestRequest{
			ContestID: 100,
			UserID:    "u-1",
			Password:  "wrong-password",
			ActorRole: "student",
		})
		if err != ports.ErrForbidden {
			t.Fatalf("rule=%s expected ErrForbidden, got %v", rule, err)
		}
	}
}

type fakeVisibilityRepo struct {
	contest Contest
}

func (r *fakeVisibilityRepo) ListVisible(ctx context.Context, offset, limit int, visibility, ruleType, status, actorUserID, actorRole string) ([]Contest, error) {
	return []Contest{r.contest}, nil
}

func (r *fakeVisibilityRepo) GetVisibleByID(ctx context.Context, contestID int64, actorUserID, actorRole string) (Contest, error) {
	if r.contest.ID != contestID || r.contest.Visibility == "private" {
		return Contest{}, ports.ErrNotFound
	}
	return r.contest, nil
}

func (r *fakeVisibilityRepo) GetByID(ctx context.Context, contestID int64) (Contest, error) {
	if r.contest.ID != contestID {
		return Contest{}, ports.ErrNotFound
	}
	return r.contest, nil
}

func (r *fakeVisibilityRepo) Create(ctx context.Context, c Contest) (Contest, error) { return c, nil }
func (r *fakeVisibilityRepo) Update(ctx context.Context, c Contest) (Contest, error) { return c, nil }
func (r *fakeVisibilityRepo) Delete(ctx context.Context, contestID int64) error      { return nil }
func (r *fakeVisibilityRepo) CreateParticipant(ctx context.Context, p Participant) (Participant, error) {
	return p, nil
}
func (r *fakeVisibilityRepo) ListScoreboardSubmissions(ctx context.Context, contestID int64, before time.Time) ([]ScoreboardSubmission, error) {
	return nil, nil
}

func TestJoinContestPrivateVisibilityNotFoundAcrossRules(t *testing.T) {
	rules := []string{"acm", "oi", "assignment"}
	for _, rule := range rules {
		repo := &fakeVisibilityRepo{contest: Contest{ID: 88, RuleType: rule, Visibility: "private"}}
		svc := NewService(repo)
		_, err := svc.JoinContest(context.Background(), ports.JoinContestRequest{ContestID: 88, UserID: "u-2", ActorRole: "student"})
		if err != ports.ErrNotFound {
			t.Fatalf("rule=%s expected ErrNotFound, got %v", rule, err)
		}
	}
}
