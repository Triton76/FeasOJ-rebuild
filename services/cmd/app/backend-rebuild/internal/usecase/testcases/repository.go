package testcases

import "context"

type Testcase struct {
	ID         string
	ProblemID  int64
	InputData  string
	OutputData string
	IsSample   bool
	SortOrder  int
}

type Repository interface {
	GetProblemOwner(ctx context.Context, problemID int64) (string, error)
	ListByProblem(ctx context.Context, problemID int64) ([]Testcase, error)
	Create(ctx context.Context, testcase Testcase) (Testcase, error)
	Update(ctx context.Context, testcase Testcase) (Testcase, error)
	Delete(ctx context.Context, problemID int64, testcaseID string) error
	GetByID(ctx context.Context, problemID int64, testcaseID string) (Testcase, error)
	Reorder(ctx context.Context, problemID int64, orderedIDs []string) ([]Testcase, error)
}
