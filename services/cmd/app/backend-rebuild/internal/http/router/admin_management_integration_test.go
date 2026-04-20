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

type fakeAdminAuthService struct{}

func (s fakeAdminAuthService) Register(ctx context.Context, req ports.RegisterRequest) (ports.UserDTO, error) {
	return ports.UserDTO{ID: "admin-1", Username: req.Username, Email: req.Email, Role: "admin", Status: "active"}, nil
}

func (s fakeAdminAuthService) Login(ctx context.Context, req ports.LoginRequest) (ports.LoginResponse, error) {
	token, err := security.GenerateToken("test-secret", "test-issuer", 2*time.Hour, "admin-1", "admin", "active")
	if err != nil {
		return ports.LoginResponse{}, err
	}
	return ports.LoginResponse{Token: token, User: ports.UserDTO{ID: "admin-1", Role: "admin", Status: "active"}}, nil
}

func (s fakeAdminAuthService) Verify(ctx context.Context) (ports.VerifyResponse, error) {
	return ports.VerifyResponse{User: ports.UserDTO{ID: "admin-1", Role: "admin", Status: "active"}}, nil
}

func (s fakeAdminAuthService) SendPasswordResetCode(ctx context.Context, req ports.PasswordResetCodeRequest) (ports.PasswordResetCodeResponse, error) {
	return ports.PasswordResetCodeResponse{ExpiresInSeconds: 300}, nil
}

func (s fakeAdminAuthService) ResetPassword(ctx context.Context, req ports.PasswordResetRequest) error {
	return nil
}

type fakeProblemsService struct {
	nextID int64
	items  map[int64]ports.ProblemDTO
}

func newFakeProblemsService() *fakeProblemsService {
	return &fakeProblemsService{nextID: 1, items: map[int64]ports.ProblemDTO{}}
}

func (s *fakeProblemsService) ListProblems(ctx context.Context, req ports.ProblemsQuery) ([]ports.ProblemDTO, error) {
	resp := make([]ports.ProblemDTO, 0, len(s.items))
	for _, item := range s.items {
		resp = append(resp, item)
	}
	return resp, nil
}

func (s *fakeProblemsService) GetProblem(ctx context.Context, problemID int64) (ports.ProblemDTO, error) {
	item, ok := s.items[problemID]
	if !ok {
		return ports.ProblemDTO{}, ports.ErrNotFound
	}
	return item, nil
}

func (s *fakeProblemsService) CreateProblem(ctx context.Context, req ports.CreateProblemRequest) (ports.ProblemDTO, error) {
	id := s.nextID
	s.nextID++
	item := ports.ProblemDTO{
		ID:            id,
		Title:         req.Title,
		Content:       req.Content,
		Input:         req.Input,
		Output:        req.Output,
		Difficulty:    req.Difficulty,
		TimeLimitMS:   req.TimeLimitMS,
		MemoryLimitMB: req.MemoryLimitMB,
		OwnerUserID:   req.ActorUserID,
		ClassID:       req.ClassID,
		Visibility:    req.Visibility,
		Status:        req.Status,
	}
	s.items[id] = item
	return item, nil
}

func (s *fakeProblemsService) UpdateProblem(ctx context.Context, req ports.UpdateProblemRequest) (ports.ProblemDTO, error) {
	item, ok := s.items[req.ProblemID]
	if !ok {
		return ports.ProblemDTO{}, ports.ErrNotFound
	}
	item.Title = req.Title
	item.Content = req.Content
	item.Input = req.Input
	item.Output = req.Output
	item.Difficulty = req.Difficulty
	item.TimeLimitMS = req.TimeLimitMS
	item.MemoryLimitMB = req.MemoryLimitMB
	item.Visibility = req.Visibility
	item.Status = req.Status
	s.items[req.ProblemID] = item
	return item, nil
}

func (s *fakeProblemsService) DeleteProblem(ctx context.Context, req ports.DeleteProblemRequest) error {
	if _, ok := s.items[req.ProblemID]; !ok {
		return ports.ErrNotFound
	}
	delete(s.items, req.ProblemID)
	return nil
}

type fakeCompetitionsService struct {
	nextID int64
	items  map[int64]ports.ContestDTO
}

func newFakeCompetitionsService() *fakeCompetitionsService {
	return &fakeCompetitionsService{nextID: 1, items: map[int64]ports.ContestDTO{}}
}

func (s *fakeCompetitionsService) ListContests(ctx context.Context, req ports.ContestsQuery) ([]ports.ContestDTO, error) {
	resp := make([]ports.ContestDTO, 0, len(s.items))
	for _, item := range s.items {
		resp = append(resp, item)
	}
	return resp, nil
}

func (s *fakeCompetitionsService) GetContest(ctx context.Context, contestID int64) (ports.ContestDTO, error) {
	item, ok := s.items[contestID]
	if !ok {
		return ports.ContestDTO{}, ports.ErrNotFound
	}
	return item, nil
}

func (s *fakeCompetitionsService) CreateContest(ctx context.Context, req ports.CreateContestRequest) (ports.ContestDTO, error) {
	id := s.nextID
	s.nextID++
	item := ports.ContestDTO{
		ID:          id,
		Title:       req.Title,
		Subtitle:    req.Subtitle,
		Visibility:  req.Visibility,
		RuleType:    req.RuleType,
		Status:      req.Status,
		IsEncrypted: req.IsEncrypted,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
	}
	s.items[id] = item
	return item, nil
}

func (s *fakeCompetitionsService) UpdateContest(ctx context.Context, req ports.UpdateContestRequest) (ports.ContestDTO, error) {
	item, ok := s.items[req.ContestID]
	if !ok {
		return ports.ContestDTO{}, ports.ErrNotFound
	}
	item.Title = req.Title
	item.Subtitle = req.Subtitle
	item.Visibility = req.Visibility
	item.RuleType = req.RuleType
	item.Status = req.Status
	item.IsEncrypted = req.IsEncrypted
	item.StartAt = req.StartAt
	item.EndAt = req.EndAt
	s.items[req.ContestID] = item
	return item, nil
}

func (s *fakeCompetitionsService) DeleteContest(ctx context.Context, req ports.DeleteContestRequest) error {
	if _, ok := s.items[req.ContestID]; !ok {
		return ports.ErrNotFound
	}
	delete(s.items, req.ContestID)
	return nil
}

func (s *fakeCompetitionsService) JoinContest(ctx context.Context, req ports.JoinContestRequest) (ports.ContestParticipantDTO, error) {
	return ports.ContestParticipantDTO{}, ports.ErrNotImplemented
}

func (s *fakeCompetitionsService) ListContestProblems(ctx context.Context, req ports.ContestProblemsQuery) ([]ports.ContestProblemBindingDTO, error) {
	return []ports.ContestProblemBindingDTO{}, nil
}

func (s *fakeCompetitionsService) ReplaceContestProblems(ctx context.Context, req ports.ReplaceContestProblemsRequest) ([]ports.ContestProblemBindingDTO, error) {
	resp := make([]ports.ContestProblemBindingDTO, 0, len(req.Items))
	for _, item := range req.Items {
		resp = append(resp, ports.ContestProblemBindingDTO{ContestID: req.ContestID, ProblemID: item.ProblemID, DisplayOrder: item.DisplayOrder, Alias: item.Alias})
	}
	return resp, nil
}

func (s *fakeCompetitionsService) GetContestMembership(ctx context.Context, req ports.ContestMembershipQuery) (ports.ContestMembershipDTO, error) {
	return ports.ContestMembershipDTO{ContestID: req.ContestID, UserID: req.ActorUserID, Joined: true, Status: "registered"}, nil
}

func (s *fakeCompetitionsService) ListContestParticipants(ctx context.Context, req ports.ContestParticipantsQuery) ([]ports.ContestParticipantDetailDTO, error) {
	return []ports.ContestParticipantDetailDTO{}, nil
}

func (s *fakeCompetitionsService) QuitContest(ctx context.Context, req ports.QuitContestRequest) (ports.ContestParticipantDTO, error) {
	return ports.ContestParticipantDTO{ContestID: req.ContestID, UserID: req.ActorUserID, Status: "quit"}, nil
}

func (s *fakeCompetitionsService) GetScoreboard(ctx context.Context, req ports.ContestScoreboardQuery) (ports.ContestScoreboardResponse, error) {
	return ports.ContestScoreboardResponse{ContestID: req.ContestID, VisibleItems: []ports.ContestScoreboardItem{}}, nil
}

type fakeSubmitForAdminIT struct{}

func (s fakeSubmitForAdminIT) CreateSubmission(context.Context, ports.CreateSubmissionRequest) (ports.SubmissionDTO, error) {
	return ports.SubmissionDTO{}, ports.ErrNotImplemented
}
func (s fakeSubmitForAdminIT) ListSubmissions(context.Context, ports.SubmissionsQuery) ([]ports.SubmissionDTO, error) {
	return nil, ports.ErrNotImplemented
}
func (s fakeSubmitForAdminIT) MarkSubmissionJudging(context.Context, int64, string) (ports.SubmissionDTO, error) {
	return ports.SubmissionDTO{}, ports.ErrNotImplemented
}
func (s fakeSubmitForAdminIT) WritebackSubmission(context.Context, ports.JudgeWritebackRequest) (ports.SubmissionDTO, error) {
	return ports.SubmissionDTO{}, ports.ErrNotImplemented
}

func TestAdminProblemContestManagementIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := handler.New(handler.Handlers{
		Auth:          fakeAdminAuthService{},
		Problems:      newFakeProblemsService(),
		Competitions:  newFakeCompetitionsService(),
		SubmitRecords: fakeSubmitForAdminIT{},
	})

	r := gin.New()
	r.Use(gin.Recovery())
	Register(r, h, config.Config{JWTSecret: "test-secret", EnableScoreboard: true, EnableJudgeWriteback: true})

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"pass123"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	r.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login failed status=%d body=%s", loginRec.Code, loginRec.Body.String())
	}

	var loginResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("decode login failed: %v", err)
	}

	createProblemReq := httptest.NewRequest(http.MethodPost, "/api/v1/problems", strings.NewReader(`{"title":"P1","content":"desc","input":"in","output":"out","difficulty":1,"time_limit_ms":1000,"memory_limit_mb":128,"class_id":"","visibility":"public","status":"published"}`))
	createProblemReq.Header.Set("Content-Type", "application/json")
	createProblemReq.Header.Set("Authorization", "Bearer "+loginResp.Data.Token)
	createProblemRec := httptest.NewRecorder()
	r.ServeHTTP(createProblemRec, createProblemReq)
	if createProblemRec.Code != http.StatusOK {
		t.Fatalf("create problem failed status=%d body=%s", createProblemRec.Code, createProblemRec.Body.String())
	}

	listProblemsReq := httptest.NewRequest(http.MethodGet, "/api/v1/problems", nil)
	listProblemsReq.Header.Set("Authorization", "Bearer "+loginResp.Data.Token)
	listProblemsRec := httptest.NewRecorder()
	r.ServeHTTP(listProblemsRec, listProblemsReq)
	if listProblemsRec.Code != http.StatusOK {
		t.Fatalf("list problems failed status=%d body=%s", listProblemsRec.Code, listProblemsRec.Body.String())
	}

	createContestReq := httptest.NewRequest(http.MethodPost, "/api/v1/contests", strings.NewReader(`{"title":"C1","subtitle":"phase2","description":"","announcement":"","class_id":"","visibility":"public","rule_type":"acm","status":"draft","is_encrypted":false,"password":"","start_at":"2026-04-20T10:00:00Z","end_at":"2026-04-20T12:00:00Z"}`))
	createContestReq.Header.Set("Content-Type", "application/json")
	createContestReq.Header.Set("Authorization", "Bearer "+loginResp.Data.Token)
	createContestRec := httptest.NewRecorder()
	r.ServeHTTP(createContestRec, createContestReq)
	if createContestRec.Code != http.StatusOK {
		t.Fatalf("create contest failed status=%d body=%s", createContestRec.Code, createContestRec.Body.String())
	}

	listContestsReq := httptest.NewRequest(http.MethodGet, "/api/v1/contests", nil)
	listContestsReq.Header.Set("Authorization", "Bearer "+loginResp.Data.Token)
	listContestsRec := httptest.NewRecorder()
	r.ServeHTTP(listContestsRec, listContestsReq)
	if listContestsRec.Code != http.StatusOK {
		t.Fatalf("list contests failed status=%d body=%s", listContestsRec.Code, listContestsRec.Body.String())
	}
}
