package testcases

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"errors"
	"sort"
	"testing"
)

type fakeRepo struct {
	owner     string
	cases     map[string]Testcase
	byProblem map[int64][]string
}

func newFakeRepo(owner string) *fakeRepo {
	return &fakeRepo{
		owner:     owner,
		cases:     make(map[string]Testcase),
		byProblem: make(map[int64][]string),
	}
}

func (r *fakeRepo) GetProblemOwner(ctx context.Context, problemID int64) (string, error) {
	if problemID <= 0 {
		return "", ports.ErrNotFound
	}
	return r.owner, nil
}

func (r *fakeRepo) ListByProblem(ctx context.Context, problemID int64) ([]Testcase, error) {
	ids := r.byProblem[problemID]
	items := make([]Testcase, 0, len(ids))
	for _, id := range ids {
		items = append(items, r.cases[id])
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].SortOrder != items[j].SortOrder {
			return items[i].SortOrder < items[j].SortOrder
		}
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func (r *fakeRepo) Create(ctx context.Context, testcase Testcase) (Testcase, error) {
	r.cases[testcase.ID] = testcase
	r.byProblem[testcase.ProblemID] = append(r.byProblem[testcase.ProblemID], testcase.ID)
	return testcase, nil
}

func (r *fakeRepo) Update(ctx context.Context, testcase Testcase) (Testcase, error) {
	if _, ok := r.cases[testcase.ID]; !ok {
		return Testcase{}, ports.ErrNotFound
	}
	r.cases[testcase.ID] = testcase
	return testcase, nil
}

func (r *fakeRepo) Delete(ctx context.Context, problemID int64, testcaseID string) error {
	if _, ok := r.cases[testcaseID]; !ok {
		return ports.ErrNotFound
	}
	delete(r.cases, testcaseID)
	ids := r.byProblem[problemID]
	filtered := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != testcaseID {
			filtered = append(filtered, id)
		}
	}
	r.byProblem[problemID] = filtered
	return nil
}

func (r *fakeRepo) GetByID(ctx context.Context, problemID int64, testcaseID string) (Testcase, error) {
	item, ok := r.cases[testcaseID]
	if !ok || item.ProblemID != problemID {
		return Testcase{}, ports.ErrNotFound
	}
	return item, nil
}

func (r *fakeRepo) Reorder(ctx context.Context, problemID int64, orderedIDs []string) ([]Testcase, error) {
	for i, id := range orderedIDs {
		item, ok := r.cases[id]
		if !ok || item.ProblemID != problemID {
			return nil, ports.ErrNotFound
		}
		item.SortOrder = i + 1
		r.cases[id] = item
	}
	r.byProblem[problemID] = append([]string{}, orderedIDs...)
	return r.ListByProblem(ctx, problemID)
}

func TestCreateTestcasePermissionAndValidation(t *testing.T) {
	repo := newFakeRepo("teacher-1")
	svc := NewService(repo)

	_, err := svc.CreateTestcase(context.Background(), ports.CreateTestcaseRequest{
		ProblemID:   1,
		InputData:   "in",
		OutputData:  "out",
		ActorUserID: "student-1",
		ActorRole:   "student",
	})
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected forbidden for student, got %v", err)
	}

	_, err = svc.CreateTestcase(context.Background(), ports.CreateTestcaseRequest{
		ProblemID:   1,
		InputData:   "",
		OutputData:  "out",
		ActorUserID: "teacher-1",
		ActorRole:   "teacher",
	})
	if !errors.Is(err, ports.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument for missing input, got %v", err)
	}
}

func TestCreateAndReorderAndJudgeRead(t *testing.T) {
	repo := newFakeRepo("teacher-1")
	svc := NewService(repo)

	first, err := svc.CreateTestcase(context.Background(), ports.CreateTestcaseRequest{
		ProblemID:   1,
		InputData:   "in1",
		OutputData:  "out1",
		ActorUserID: "teacher-1",
		ActorRole:   "teacher",
	})
	if err != nil {
		t.Fatalf("create first failed: %v", err)
	}
	second, err := svc.CreateTestcase(context.Background(), ports.CreateTestcaseRequest{
		ProblemID:   1,
		InputData:   "in2",
		OutputData:  "out2",
		ActorUserID: "teacher-1",
		ActorRole:   "teacher",
	})
	if err != nil {
		t.Fatalf("create second failed: %v", err)
	}

	items, err := svc.ReorderTestcases(context.Background(), ports.ReorderTestcasesRequest{
		ProblemID:   1,
		TestcaseIDs: []string{second.ID, first.ID},
		ActorUserID: "teacher-1",
		ActorRole:   "teacher",
	})
	if err != nil {
		t.Fatalf("reorder failed: %v", err)
	}
	if len(items) != 2 || items[0].ID != second.ID || items[0].SortOrder != 1 {
		t.Fatalf("unexpected reorder output: %+v", items)
	}

	judgeItems, err := svc.ListTestcasesForJudge(context.Background(), ports.JudgeListTestcasesRequest{ProblemID: 1})
	if err != nil {
		t.Fatalf("judge read failed: %v", err)
	}
	if len(judgeItems) != 2 {
		t.Fatalf("expected 2 judge testcases, got %d", len(judgeItems))
	}
}

func TestListUpdateDeletePermission(t *testing.T) {
	repo := newFakeRepo("teacher-1")
	svc := NewService(repo)

	created, err := svc.CreateTestcase(context.Background(), ports.CreateTestcaseRequest{
		ProblemID:   1,
		InputData:   "in",
		OutputData:  "out",
		ActorUserID: "teacher-1",
		ActorRole:   "teacher",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	_, err = svc.ListTestcases(context.Background(), ports.ListTestcasesRequest{
		ProblemID:   1,
		ActorUserID: "teacher-2",
		ActorRole:   "teacher",
	})
	if !errors.Is(err, ports.ErrForbidden) {
		t.Fatalf("expected forbidden for non-owner teacher, got %v", err)
	}

	updated, err := svc.UpdateTestcase(context.Background(), ports.UpdateTestcaseRequest{
		ProblemID:   1,
		TestcaseID:  created.ID,
		InputData:   "in-updated",
		OutputData:  "out-updated",
		IsSample:    true,
		ActorUserID: "teacher-1",
		ActorRole:   "teacher",
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.InputData != "in-updated" || !updated.IsSample {
		t.Fatalf("unexpected update result: %+v", updated)
	}

	err = svc.DeleteTestcase(context.Background(), ports.DeleteTestcaseRequest{
		ProblemID:   1,
		TestcaseID:  created.ID,
		ActorUserID: "teacher-1",
		ActorRole:   "teacher",
	})
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
}
