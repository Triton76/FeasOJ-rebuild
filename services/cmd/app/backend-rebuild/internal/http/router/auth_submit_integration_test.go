package router

import (
	"FeasOJ/app/backend-rebuild/internal/config"
	"FeasOJ/app/backend-rebuild/internal/http/handler"
	"FeasOJ/app/backend-rebuild/internal/ports"
	"FeasOJ/app/backend-rebuild/internal/security"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakeAuthService struct{}

func (s fakeAuthService) Register(ctx context.Context, req ports.RegisterRequest) (ports.UserDTO, error) {
	return ports.UserDTO{ID: "u-1", Username: req.Username, Email: req.Email, Role: "student", Status: "active"}, nil
}

func (s fakeAuthService) Login(ctx context.Context, req ports.LoginRequest) (ports.LoginResponse, error) {
	token, err := security.GenerateToken("test-secret", "test-issuer", 2*time.Hour, "u-1", "student", "active")
	if err != nil {
		return ports.LoginResponse{}, err
	}
	return ports.LoginResponse{
		Token: token,
		User:  ports.UserDTO{ID: "u-1", Username: req.Username, Role: "student", Status: "active"},
	}, nil
}

func (s fakeAuthService) Verify(ctx context.Context) (ports.VerifyResponse, error) {
	return ports.VerifyResponse{User: ports.UserDTO{ID: "u-1", Role: "student", Status: "active"}}, nil
}

func (s fakeAuthService) SendPasswordResetCode(ctx context.Context, req ports.PasswordResetCodeRequest) (ports.PasswordResetCodeResponse, error) {
	return ports.PasswordResetCodeResponse{ExpiresInSeconds: 300}, nil
}

func (s fakeAuthService) ResetPassword(ctx context.Context, req ports.PasswordResetRequest) error {
	return nil
}

type fakeSubmitService struct {
	lastReq ports.CreateSubmissionRequest
}

func (s *fakeSubmitService) CreateSubmission(ctx context.Context, req ports.CreateSubmissionRequest) (ports.SubmissionDTO, error) {
	s.lastReq = req
	return ports.SubmissionDTO{ID: 1, UserID: req.UserID, ProblemID: req.ProblemID, ContestID: req.ContestID, Language: req.Language, Result: "pending"}, nil
}

func (s *fakeSubmitService) ListSubmissions(ctx context.Context, req ports.SubmissionsQuery) ([]ports.SubmissionDTO, error) {
	return []ports.SubmissionDTO{}, nil
}

func (s *fakeSubmitService) MarkSubmissionJudging(ctx context.Context, submissionID int64, source string) (ports.SubmissionDTO, error) {
	return ports.SubmissionDTO{ID: submissionID, Result: "judging"}, nil
}

func (s *fakeSubmitService) WritebackSubmission(ctx context.Context, req ports.JudgeWritebackRequest) (ports.SubmissionDTO, error) {
	return ports.SubmissionDTO{ID: req.SubmissionID, Result: req.Result}, nil
}

type fakeJudgeProblemsService struct{}

func (s fakeJudgeProblemsService) ListProblems(context.Context, ports.ProblemsQuery) ([]ports.ProblemDTO, error) {
	return nil, ports.ErrNotImplemented
}

func (s fakeJudgeProblemsService) GetProblem(context.Context, int64) (ports.ProblemDTO, error) {
	return ports.ProblemDTO{}, ports.ErrNotImplemented
}

func (s fakeJudgeProblemsService) GetProblemForJudge(_ context.Context, problemID int64) (ports.ProblemDTO, error) {
	return ports.ProblemDTO{ID: problemID, TimeLimitMS: 1000, MemoryLimitMB: 128}, nil
}

func (s fakeJudgeProblemsService) CreateProblem(context.Context, ports.CreateProblemRequest) (ports.ProblemDTO, error) {
	return ports.ProblemDTO{}, ports.ErrNotImplemented
}

func (s fakeJudgeProblemsService) UpdateProblem(context.Context, ports.UpdateProblemRequest) (ports.ProblemDTO, error) {
	return ports.ProblemDTO{}, ports.ErrNotImplemented
}

func (s fakeJudgeProblemsService) DeleteProblem(context.Context, ports.DeleteProblemRequest) error {
	return ports.ErrNotImplemented
}

type fakeJudgeTestcasesService struct{}

func (s fakeJudgeTestcasesService) CreateTestcase(context.Context, ports.CreateTestcaseRequest) (ports.TestcaseDTO, error) {
	return ports.TestcaseDTO{}, ports.ErrNotImplemented
}

func (s fakeJudgeTestcasesService) ListTestcases(context.Context, ports.ListTestcasesRequest) ([]ports.TestcaseDTO, error) {
	return nil, ports.ErrNotImplemented
}

func (s fakeJudgeTestcasesService) UpdateTestcase(context.Context, ports.UpdateTestcaseRequest) (ports.TestcaseDTO, error) {
	return ports.TestcaseDTO{}, ports.ErrNotImplemented
}

func (s fakeJudgeTestcasesService) DeleteTestcase(context.Context, ports.DeleteTestcaseRequest) error {
	return ports.ErrNotImplemented
}

func (s fakeJudgeTestcasesService) ReorderTestcases(context.Context, ports.ReorderTestcasesRequest) ([]ports.TestcaseDTO, error) {
	return nil, ports.ErrNotImplemented
}

func (s fakeJudgeTestcasesService) ListTestcasesForJudge(context.Context, ports.JudgeListTestcasesRequest) ([]ports.TestcaseDTO, error) {
	return []ports.TestcaseDTO{{ID: "tc-1", ProblemID: 1, InputData: "1 2", OutputData: "3", SortOrder: 1}}, nil
}

func TestAuthThenSubmitIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	submitSvc := &fakeSubmitService{}
	h := handler.New(handler.Handlers{
		Auth:          fakeAuthService{},
		Users:         nil,
		Classes:       nil,
		Problems:      nil,
		Competitions:  nil,
		Discussions:   nil,
		SubmitRecords: submitSvc,
		Admin:         nil,
	})

	r := gin.New()
	r.Use(gin.Recovery())
	Register(r, h, config.Config{JWTSecret: "test-secret"})

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"alice","password":"pass123"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	r.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", loginRec.Code, loginRec.Body.String())
	}

	var loginResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("decode login response failed: %v", err)
	}
	if loginResp.Data.Token == "" {
		t.Fatal("empty login token")
	}

	submitReq := httptest.NewRequest(http.MethodPost, "/api/v1/submit-records", strings.NewReader(`{"problem_id":1001,"contest_id":0,"language":"cpp","source_code":"int main(){return 0;}"}`))
	submitReq.Header.Set("Content-Type", "application/json")
	submitReq.Header.Set("Authorization", "Bearer "+loginResp.Data.Token)
	submitRec := httptest.NewRecorder()
	r.ServeHTTP(submitRec, submitReq)
	if submitRec.Code != http.StatusOK {
		t.Fatalf("submit status=%d body=%s", submitRec.Code, submitRec.Body.String())
	}

	if submitSvc.lastReq.UserID != "u-1" {
		t.Fatalf("expected user_id from claims to be u-1, got %s", submitSvc.lastReq.UserID)
	}
}

func TestJudgeWritebackAuthFailureIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	submitSvc := &fakeSubmitService{}
	h := handler.New(handler.Handlers{
		Auth:                fakeAuthService{},
		SubmitRecords:       submitSvc,
		JudgeWritebackToken: "judge-secret",
	})

	r := gin.New()
	r.Use(gin.Recovery())
	Register(r, h, config.Config{JWTSecret: "test-secret", EnableJudgeWriteback: true})

	payload := `{"contract_version":"v1","submission_id":1,"result":"accepted","score":100,"source":"judgecore"}`

	badReq := httptest.NewRequest(http.MethodPost, "/api/v1/judge/writeback", strings.NewReader(payload))
	badReq.Header.Set("Content-Type", "application/json")
	badRec := httptest.NewRecorder()
	r.ServeHTTP(badRec, badReq)
	if badRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d body=%s", badRec.Code, badRec.Body.String())
	}

	goodReq := httptest.NewRequest(http.MethodPost, "/api/v1/judge/writeback", strings.NewReader(payload))
	goodReq.Header.Set("Content-Type", "application/json")
	goodReq.Header.Set("X-Judge-Token", "judge-secret")
	goodRec := httptest.NewRecorder()
	r.ServeHTTP(goodRec, goodReq)
	if goodRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid token, got %d body=%s", goodRec.Code, goodRec.Body.String())
	}
}

func TestJudgeProblemBundleAndMarkJudgingAuthIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	submitSvc := &fakeSubmitService{}
	h := handler.New(handler.Handlers{
		Auth:                fakeAuthService{},
		Problems:            fakeJudgeProblemsService{},
		Testcases:           fakeJudgeTestcasesService{},
		SubmitRecords:       submitSvc,
		JudgeWritebackToken: "judge-secret",
	})

	r := gin.New()
	r.Use(gin.Recovery())
	Register(r, h, config.Config{
		JWTSecret:            "test-secret",
		EnableJudgeWriteback: true,
		EnableTestcaseAPIs:   true,
	})

	badBundleReq := httptest.NewRequest(http.MethodGet, "/api/v1/judge/problems/1/bundle", nil)
	badBundleRec := httptest.NewRecorder()
	r.ServeHTTP(badBundleRec, badBundleReq)
	if badBundleRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing judge token, got %d body=%s", badBundleRec.Code, badBundleRec.Body.String())
	}

	goodBundleReq := httptest.NewRequest(http.MethodGet, "/api/v1/judge/problems/1/bundle", nil)
	goodBundleReq.Header.Set("X-Judge-Token", "judge-secret")
	goodBundleRec := httptest.NewRecorder()
	r.ServeHTTP(goodBundleRec, goodBundleReq)
	if goodBundleRec.Code != http.StatusOK || !strings.Contains(goodBundleRec.Body.String(), `"testcases"`) {
		t.Fatalf("expected 200 bundle response, got %d body=%s", goodBundleRec.Code, goodBundleRec.Body.String())
	}

	goodMarkReq := httptest.NewRequest(http.MethodPost, "/api/v1/judge/submissions/1/judging", nil)
	goodMarkReq.Header.Set("X-Judge-Token", "judge-secret")
	goodMarkRec := httptest.NewRecorder()
	r.ServeHTTP(goodMarkRec, goodMarkReq)
	if goodMarkRec.Code != http.StatusOK || !strings.Contains(goodMarkRec.Body.String(), `"judging"`) {
		t.Fatalf("expected 200 judging response, got %d body=%s", goodMarkRec.Code, goodMarkRec.Body.String())
	}
}
