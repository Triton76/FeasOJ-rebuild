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
