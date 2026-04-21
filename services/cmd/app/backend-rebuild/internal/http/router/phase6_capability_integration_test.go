package router

import (
	"FeasOJ/app/backend-rebuild/internal/config"
	"FeasOJ/app/backend-rebuild/internal/http/handler"
	"FeasOJ/app/backend-rebuild/internal/ports"
	"FeasOJ/app/backend-rebuild/internal/security"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakePhase6AuthService struct{}

func (s fakePhase6AuthService) Register(ctx context.Context, req ports.RegisterRequest) (ports.UserDTO, error) {
	role := "student"
	if req.Username == "teacher" || req.Username == "admin" {
		role = req.Username
	}
	id := req.Username + "-id"
	return ports.UserDTO{ID: id, Username: req.Username, Email: req.Email, Role: role, Status: "active"}, nil
}

func (s fakePhase6AuthService) Login(ctx context.Context, req ports.LoginRequest) (ports.LoginResponse, error) {
	role := "student"
	if req.Username == "teacher" || req.Username == "admin" {
		role = req.Username
	}
	id := req.Username + "-id"
	token, err := security.GenerateToken("test-secret", "test-issuer", 2*time.Hour, id, role, "active")
	if err != nil {
		return ports.LoginResponse{}, err
	}
	return ports.LoginResponse{Token: token, User: ports.UserDTO{ID: id, Role: role, Status: "active"}}, nil
}

func (s fakePhase6AuthService) Verify(ctx context.Context) (ports.VerifyResponse, error) {
	claims, _ := security.ClaimsFromContext(ctx)
	return ports.VerifyResponse{User: ports.UserDTO{ID: claims.UserID, Role: claims.Role, Status: "active"}, Capabilities: ports.AuthCapabilities{PasswordResetEnabled: false}}, nil
}

func (s fakePhase6AuthService) SendPasswordResetCode(ctx context.Context, req ports.PasswordResetCodeRequest) (ports.PasswordResetCodeResponse, error) {
	return ports.PasswordResetCodeResponse{ExpiresInSeconds: 300}, nil
}

func (s fakePhase6AuthService) ResetPassword(ctx context.Context, req ports.PasswordResetRequest) error {
	return nil
}

type phase6Participant struct {
	id        string
	contestID int64
	userID    string
	status    string
	joinedAt  string
}

type fakePhase6CompetitionsService struct {
	participants map[string]phase6Participant
}

func newFakePhase6CompetitionsService() *fakePhase6CompetitionsService {
	return &fakePhase6CompetitionsService{participants: map[string]phase6Participant{}}
}

func (s *fakePhase6CompetitionsService) ListContests(context.Context, ports.ContestsQuery) ([]ports.ContestDTO, error) {
	return []ports.ContestDTO{}, nil
}

func (s *fakePhase6CompetitionsService) GetContest(context.Context, int64) (ports.ContestDTO, error) {
	return ports.ContestDTO{ID: 1, Title: "C1", OwnerUserID: "teacher-id"}, nil
}

func (s *fakePhase6CompetitionsService) CreateContest(context.Context, ports.CreateContestRequest) (ports.ContestDTO, error) {
	return ports.ContestDTO{}, ports.ErrNotImplemented
}

func (s *fakePhase6CompetitionsService) UpdateContest(context.Context, ports.UpdateContestRequest) (ports.ContestDTO, error) {
	return ports.ContestDTO{}, ports.ErrNotImplemented
}

func (s *fakePhase6CompetitionsService) DeleteContest(context.Context, ports.DeleteContestRequest) error {
	return ports.ErrNotImplemented
}

func (s *fakePhase6CompetitionsService) ListContestProblems(context.Context, ports.ContestProblemsQuery) ([]ports.ContestProblemBindingDTO, error) {
	return []ports.ContestProblemBindingDTO{{ContestID: 1, ProblemID: 1001, DisplayOrder: 1, Alias: "A"}}, nil
}

func (s *fakePhase6CompetitionsService) ReplaceContestProblems(ctx context.Context, req ports.ReplaceContestProblemsRequest) ([]ports.ContestProblemBindingDTO, error) {
	resp := make([]ports.ContestProblemBindingDTO, 0, len(req.Items))
	for _, item := range req.Items {
		resp = append(resp, ports.ContestProblemBindingDTO{ContestID: req.ContestID, ProblemID: item.ProblemID, DisplayOrder: item.DisplayOrder, Alias: item.Alias})
	}
	return resp, nil
}

func (s *fakePhase6CompetitionsService) JoinContest(ctx context.Context, req ports.JoinContestRequest) (ports.ContestParticipantDTO, error) {
	key := req.UserID
	if existing, ok := s.participants[key]; ok && existing.status != "quit" {
		return ports.ContestParticipantDTO{}, ports.ErrConflict
	}
	now := time.Now().UTC().Format(time.RFC3339)
	p := phase6Participant{id: "p-" + key, contestID: req.ContestID, userID: req.UserID, status: "registered", joinedAt: now}
	s.participants[key] = p
	return ports.ContestParticipantDTO{ID: p.id, ContestID: p.contestID, UserID: p.userID, Status: p.status}, nil
}

func (s *fakePhase6CompetitionsService) GetContestMembership(ctx context.Context, req ports.ContestMembershipQuery) (ports.ContestMembershipDTO, error) {
	if p, ok := s.participants[req.ActorUserID]; ok {
		return ports.ContestMembershipDTO{ContestID: req.ContestID, UserID: req.ActorUserID, Joined: true, Status: p.status}, nil
	}
	return ports.ContestMembershipDTO{ContestID: req.ContestID, UserID: req.ActorUserID, Joined: false, Status: ""}, nil
}

func (s *fakePhase6CompetitionsService) ListContestParticipants(ctx context.Context, req ports.ContestParticipantsQuery) ([]ports.ContestParticipantDetailDTO, error) {
	if _, ok := s.participants[req.ActorUserID]; req.ActorRole != "admin" && req.ActorUserID != "teacher-id" && !ok {
		return nil, ports.ErrForbidden
	}
	resp := make([]ports.ContestParticipantDetailDTO, 0, len(s.participants))
	for _, p := range s.participants {
		resp = append(resp, ports.ContestParticipantDetailDTO{ID: p.id, ContestID: p.contestID, UserID: p.userID, Username: p.userID, Avatar: "", Status: p.status, JoinedAt: p.joinedAt})
	}
	return resp, nil
}

func (s *fakePhase6CompetitionsService) QuitContest(ctx context.Context, req ports.QuitContestRequest) (ports.ContestParticipantDTO, error) {
	p, ok := s.participants[req.ActorUserID]
	if !ok {
		return ports.ContestParticipantDTO{}, ports.ErrNotFound
	}
	if p.status == "quit" || p.status == "finished" {
		return ports.ContestParticipantDTO{}, ports.ErrConflict
	}
	p.status = "quit"
	s.participants[req.ActorUserID] = p
	return ports.ContestParticipantDTO{ID: p.id, ContestID: p.contestID, UserID: p.userID, Status: p.status}, nil
}

func (s *fakePhase6CompetitionsService) GetScoreboard(context.Context, ports.ContestScoreboardQuery) (ports.ContestScoreboardResponse, error) {
	return ports.ContestScoreboardResponse{}, nil
}

type fakePhase6ProblemsService struct{}

func (s fakePhase6ProblemsService) ListProblems(context.Context, ports.ProblemsQuery) ([]ports.ProblemDTO, error) {
	return nil, ports.ErrNotImplemented
}

func (s fakePhase6ProblemsService) GetProblem(context.Context, int64) (ports.ProblemDTO, error) {
	return ports.ProblemDTO{}, ports.ErrForbidden
}

func (s fakePhase6ProblemsService) GetProblemForJudge(_ context.Context, problemID int64) (ports.ProblemDTO, error) {
	return ports.ProblemDTO{ID: problemID, Title: "Contest Problem", TimeLimitMS: 1000, MemoryLimitMB: 128}, nil
}

func (s fakePhase6ProblemsService) CreateProblem(context.Context, ports.CreateProblemRequest) (ports.ProblemDTO, error) {
	return ports.ProblemDTO{}, ports.ErrNotImplemented
}

func (s fakePhase6ProblemsService) UpdateProblem(context.Context, ports.UpdateProblemRequest) (ports.ProblemDTO, error) {
	return ports.ProblemDTO{}, ports.ErrNotImplemented
}

func (s fakePhase6ProblemsService) DeleteProblem(context.Context, ports.DeleteProblemRequest) error {
	return ports.ErrNotImplemented
}

type fakePhase6DiscussionsService struct {
	discussions map[string]ports.DiscussionDTO
	comments    map[string]ports.CommentDTO
}

func newFakePhase6DiscussionsService() *fakePhase6DiscussionsService {
	return &fakePhase6DiscussionsService{
		discussions: map[string]ports.DiscussionDTO{
			"d-1": {ID: "d-1", UserID: "owner-id", Username: "owner"},
		},
		comments: map[string]ports.CommentDTO{
			"c-1": {ID: "c-1", DiscussionID: "d-1", UserID: "owner-id", Username: "owner"},
		},
	}
}

func (s *fakePhase6DiscussionsService) ListDiscussions(context.Context, ports.DiscussionsQuery) ([]ports.DiscussionDTO, error) {
	return []ports.DiscussionDTO{}, nil
}

func (s *fakePhase6DiscussionsService) GetDiscussion(context.Context, string) (ports.DiscussionDTO, error) {
	return ports.DiscussionDTO{}, nil
}

func (s *fakePhase6DiscussionsService) CreateDiscussion(context.Context, ports.CreateDiscussionRequest) (ports.DiscussionDTO, error) {
	return ports.DiscussionDTO{}, nil
}

func (s *fakePhase6DiscussionsService) CreateComment(context.Context, ports.CreateCommentRequest) (ports.CommentDTO, error) {
	return ports.CommentDTO{}, nil
}

func (s *fakePhase6DiscussionsService) ListComments(context.Context, ports.CommentsQuery) ([]ports.CommentDTO, error) {
	return []ports.CommentDTO{}, nil
}

func (s *fakePhase6DiscussionsService) DeleteDiscussion(ctx context.Context, req ports.DeleteDiscussionRequest) error {
	d, ok := s.discussions[req.DiscussionID]
	if !ok {
		return ports.ErrNotFound
	}
	if req.ActorRole != "admin" && req.ActorUserID != d.UserID {
		return ports.ErrForbidden
	}
	delete(s.discussions, req.DiscussionID)
	return nil
}

func (s *fakePhase6DiscussionsService) DeleteComment(ctx context.Context, req ports.DeleteCommentRequest) error {
	c, ok := s.comments[req.CommentID]
	if !ok {
		return ports.ErrNotFound
	}
	if req.ActorRole != "admin" && req.ActorUserID != c.UserID {
		return ports.ErrForbidden
	}
	delete(s.comments, req.CommentID)
	return nil
}

func TestContestMembershipParticipantsAndQuitSemantics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := handler.New(handler.Handlers{
		Auth:         fakePhase6AuthService{},
		Problems:     fakePhase6ProblemsService{},
		Competitions: newFakePhase6CompetitionsService(),
	})
	r := gin.New()
	r.Use(gin.Recovery())
	Register(r, h, config.Config{JWTSecret: "test-secret", EnableScoreboard: true})

	studentToken := loginToken(t, r, "student")
	outsiderToken := loginToken(t, r, "outsider")

	joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/contests/join", strings.NewReader(`{"contest_id":1}`))
	joinReq.Header.Set("Content-Type", "application/json")
	joinReq.Header.Set("Authorization", "Bearer "+studentToken)
	joinRec := httptest.NewRecorder()
	r.ServeHTTP(joinRec, joinReq)
	if joinRec.Code != http.StatusOK {
		t.Fatalf("join contest failed status=%d body=%s", joinRec.Code, joinRec.Body.String())
	}

	membershipReq := httptest.NewRequest(http.MethodGet, "/api/v1/contests/1/participant/self", nil)
	membershipReq.Header.Set("Authorization", "Bearer "+studentToken)
	membershipRec := httptest.NewRecorder()
	r.ServeHTTP(membershipRec, membershipReq)
	if membershipRec.Code != http.StatusOK || !strings.Contains(membershipRec.Body.String(), `"joined":true`) {
		t.Fatalf("membership query mismatch status=%d body=%s", membershipRec.Code, membershipRec.Body.String())
	}

	participantsForbiddenReq := httptest.NewRequest(http.MethodGet, "/api/v1/contests/1/participants", nil)
	participantsForbiddenReq.Header.Set("Authorization", "Bearer "+outsiderToken)
	participantsForbiddenRec := httptest.NewRecorder()
	r.ServeHTTP(participantsForbiddenRec, participantsForbiddenReq)
	if participantsForbiddenRec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden participants query for outsider, got %d body=%s", participantsForbiddenRec.Code, participantsForbiddenRec.Body.String())
	}

	participantsReq := httptest.NewRequest(http.MethodGet, "/api/v1/contests/1/participants", nil)
	participantsReq.Header.Set("Authorization", "Bearer "+studentToken)
	participantsRec := httptest.NewRecorder()
	r.ServeHTTP(participantsRec, participantsReq)
	if participantsRec.Code != http.StatusOK {
		t.Fatalf("participants query failed status=%d body=%s", participantsRec.Code, participantsRec.Body.String())
	}

	problemsReq := httptest.NewRequest(http.MethodGet, "/api/v1/contests/1/problems", nil)
	problemsReq.Header.Set("Authorization", "Bearer "+studentToken)
	problemsRec := httptest.NewRecorder()
	r.ServeHTTP(problemsRec, problemsReq)
	if problemsRec.Code != http.StatusOK {
		t.Fatalf("contest problems query failed status=%d body=%s", problemsRec.Code, problemsRec.Body.String())
	}

	contestProblemReq := httptest.NewRequest(http.MethodGet, "/api/v1/contests/1/problems/1001", nil)
	contestProblemReq.Header.Set("Authorization", "Bearer "+studentToken)
	contestProblemRec := httptest.NewRecorder()
	r.ServeHTTP(contestProblemRec, contestProblemReq)
	if contestProblemRec.Code != http.StatusOK || !strings.Contains(contestProblemRec.Body.String(), `"id":1001`) {
		t.Fatalf("contest problem detail failed status=%d body=%s", contestProblemRec.Code, contestProblemRec.Body.String())
	}

	quitReq := httptest.NewRequest(http.MethodDelete, "/api/v1/contests/1/participant/self", nil)
	quitReq.Header.Set("Authorization", "Bearer "+studentToken)
	quitRec := httptest.NewRecorder()
	r.ServeHTTP(quitRec, quitReq)
	if quitRec.Code != http.StatusOK {
		t.Fatalf("quit contest failed status=%d body=%s", quitRec.Code, quitRec.Body.String())
	}

	quitAgainReq := httptest.NewRequest(http.MethodDelete, "/api/v1/contests/1/participant/self", nil)
	quitAgainReq.Header.Set("Authorization", "Bearer "+studentToken)
	quitAgainRec := httptest.NewRecorder()
	r.ServeHTTP(quitAgainRec, quitAgainReq)
	if quitAgainRec.Code != http.StatusConflict {
		t.Fatalf("expected quit conflict after terminal state, got %d body=%s", quitAgainRec.Code, quitAgainRec.Body.String())
	}
}

func TestDiscussionDeletionAuthorizationRegression(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := handler.New(handler.Handlers{
		Auth:        fakePhase6AuthService{},
		Discussions: newFakePhase6DiscussionsService(),
	})
	r := gin.New()
	r.Use(gin.Recovery())
	Register(r, h, config.Config{JWTSecret: "test-secret"})

	ownerToken := loginToken(t, r, "owner")
	otherToken := loginToken(t, r, "other")

	forbiddenReq := httptest.NewRequest(http.MethodDelete, "/api/v1/comments/c-1", nil)
	forbiddenReq.Header.Set("Authorization", "Bearer "+otherToken)
	forbiddenRec := httptest.NewRecorder()
	r.ServeHTTP(forbiddenRec, forbiddenReq)
	if forbiddenRec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden for non-owner delete comment, got %d body=%s", forbiddenRec.Code, forbiddenRec.Body.String())
	}

	successReq := httptest.NewRequest(http.MethodDelete, "/api/v1/comments/c-1", nil)
	successReq.Header.Set("Authorization", "Bearer "+ownerToken)
	successRec := httptest.NewRecorder()
	r.ServeHTTP(successRec, successReq)
	if successRec.Code != http.StatusNoContent {
		t.Fatalf("expected owner delete success, got %d body=%s", successRec.Code, successRec.Body.String())
	}
}

type fakeResetCapabilityAuthService struct {
	enabled bool
}

func (s fakeResetCapabilityAuthService) Register(ctx context.Context, req ports.RegisterRequest) (ports.UserDTO, error) {
	return ports.UserDTO{ID: "u-1", Username: req.Username, Email: req.Email, Role: "student", Status: "active"}, nil
}

func (s fakeResetCapabilityAuthService) Login(ctx context.Context, req ports.LoginRequest) (ports.LoginResponse, error) {
	token, err := security.GenerateToken("test-secret", "test-issuer", 2*time.Hour, "u-1", "student", "active")
	if err != nil {
		return ports.LoginResponse{}, err
	}
	return ports.LoginResponse{Token: token, User: ports.UserDTO{ID: "u-1", Role: "student", Status: "active"}}, nil
}

func (s fakeResetCapabilityAuthService) Verify(ctx context.Context) (ports.VerifyResponse, error) {
	return ports.VerifyResponse{
		User: ports.UserDTO{ID: "u-1", Username: "u1", Role: "student", Status: "active"},
		Capabilities: ports.AuthCapabilities{
			PasswordResetEnabled: s.enabled,
		},
	}, nil
}

func (s fakeResetCapabilityAuthService) SendPasswordResetCode(ctx context.Context, req ports.PasswordResetCodeRequest) (ports.PasswordResetCodeResponse, error) {
	if !s.enabled {
		return ports.PasswordResetCodeResponse{}, ports.ErrCapabilityDisabled
	}
	return ports.PasswordResetCodeResponse{ExpiresInSeconds: 300}, nil
}

func (s fakeResetCapabilityAuthService) ResetPassword(ctx context.Context, req ports.PasswordResetRequest) error {
	if !s.enabled {
		return ports.ErrCapabilityDisabled
	}
	return nil
}

func TestPasswordResetCapabilityEnabledDisabledIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, tc := range []struct {
		name           string
		enabled        bool
		expectCode     int
		expectVerifyKV string
	}{
		{name: "disabled", enabled: false, expectCode: http.StatusForbidden, expectVerifyKV: `"password_reset_enabled":false`},
		{name: "enabled", enabled: true, expectCode: http.StatusOK, expectVerifyKV: `"password_reset_enabled":true`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := handler.New(handler.Handlers{Auth: fakeResetCapabilityAuthService{enabled: tc.enabled}})
			r := gin.New()
			r.Use(gin.Recovery())
			Register(r, h, config.Config{JWTSecret: "test-secret"})

			token := loginToken(t, r, "student")

			verifyReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/verify", nil)
			verifyReq.Header.Set("Authorization", "Bearer "+token)
			verifyRec := httptest.NewRecorder()
			r.ServeHTTP(verifyRec, verifyReq)
			if verifyRec.Code != http.StatusOK || !strings.Contains(verifyRec.Body.String(), tc.expectVerifyKV) {
				t.Fatalf("verify capability mismatch status=%d body=%s", verifyRec.Code, verifyRec.Body.String())
			}

			codeReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/reset/code", strings.NewReader(`{"email":"u1@example.com"}`))
			codeReq.Header.Set("Content-Type", "application/json")
			codeRec := httptest.NewRecorder()
			r.ServeHTTP(codeRec, codeReq)
			if codeRec.Code != tc.expectCode {
				t.Fatalf("expected status=%d got=%d body=%s", tc.expectCode, codeRec.Code, codeRec.Body.String())
			}
		})
	}
}
