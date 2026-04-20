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

type fakeClassAuthService struct{}

func (s fakeClassAuthService) Register(ctx context.Context, req ports.RegisterRequest) (ports.UserDTO, error) {
	role := "student"
	id := "student-1"
	if req.Username == "teacher" {
		role = "teacher"
		id = "teacher-1"
	}
	return ports.UserDTO{ID: id, Username: req.Username, Email: req.Email, Role: role, Status: "active"}, nil
}

func (s fakeClassAuthService) Login(ctx context.Context, req ports.LoginRequest) (ports.LoginResponse, error) {
	role := "student"
	id := "student-1"
	if req.Username == "teacher" {
		role = "teacher"
		id = "teacher-1"
	}
	token, err := security.GenerateToken("test-secret", "test-issuer", 2*time.Hour, id, role, "active")
	if err != nil {
		return ports.LoginResponse{}, err
	}
	return ports.LoginResponse{Token: token, User: ports.UserDTO{ID: id, Role: role, Status: "active"}}, nil
}

func (s fakeClassAuthService) Verify(ctx context.Context) (ports.VerifyResponse, error) {
	claims, _ := security.ClaimsFromContext(ctx)
	return ports.VerifyResponse{User: ports.UserDTO{ID: claims.UserID, Role: claims.Role, Status: "active"}}, nil
}

func (s fakeClassAuthService) SendPasswordResetCode(ctx context.Context, req ports.PasswordResetCodeRequest) (ports.PasswordResetCodeResponse, error) {
	return ports.PasswordResetCodeResponse{ExpiresInSeconds: 300}, nil
}

func (s fakeClassAuthService) ResetPassword(ctx context.Context, req ports.PasswordResetRequest) error {
	return nil
}

type fakeClassService struct {
	memberships map[string]ports.ClassMembershipDTO
}

func newFakeClassService() *fakeClassService {
	return &fakeClassService{memberships: map[string]ports.ClassMembershipDTO{}}
}

func (s *fakeClassService) CreateClass(context.Context, ports.CreateClassRequest) (ports.ClassDTO, error) {
	return ports.ClassDTO{}, ports.ErrNotImplemented
}
func (s *fakeClassService) UpdateClass(context.Context, ports.UpdateClassRequest) (ports.ClassDTO, error) {
	return ports.ClassDTO{}, ports.ErrNotImplemented
}
func (s *fakeClassService) ArchiveClass(context.Context, string, string) error {
	return ports.ErrNotImplemented
}

func (s *fakeClassService) ApplyJoinClass(ctx context.Context, req ports.ApplyJoinClassRequest) (ports.ClassMembershipDTO, error) {
	if req.ClassCode != "CLS-1" {
		return ports.ClassMembershipDTO{}, ports.ErrNotFound
	}
	if _, ok := s.memberships[req.UserID]; ok {
		return ports.ClassMembershipDTO{}, ports.ErrConflict
	}
	m := ports.ClassMembershipDTO{ID: "m-" + req.UserID, ClassID: "class-1", UserID: req.UserID, RoleInClass: "student", Status: "pending"}
	s.memberships[req.UserID] = m
	return m, nil
}

func (s *fakeClassService) ReviewMembership(ctx context.Context, req ports.ReviewMembershipRequest) (ports.ClassMembershipDTO, error) {
	if req.ActorUserID != "teacher-1" {
		return ports.ClassMembershipDTO{}, ports.ErrForbidden
	}
	for userID, m := range s.memberships {
		if m.ID == req.MembershipID {
			if m.Status != "pending" {
				return ports.ClassMembershipDTO{}, ports.ErrConflict
			}
			if req.Approve {
				m.Status = "active"
			} else {
				m.Status = "rejected"
			}
			s.memberships[userID] = m
			return m, nil
		}
	}
	return ports.ClassMembershipDTO{}, ports.ErrNotFound
}

func (s *fakeClassService) ListClassMemberships(ctx context.Context, classID string, actorUserID string) ([]ports.ClassMembershipDTO, error) {
	resp := make([]ports.ClassMembershipDTO, 0, len(s.memberships))
	for _, m := range s.memberships {
		resp = append(resp, m)
	}
	return resp, nil
}

func (s *fakeClassService) ListMyMemberships(ctx context.Context, userID string) ([]ports.ClassMembershipDTO, error) {
	if m, ok := s.memberships[userID]; ok {
		return []ports.ClassMembershipDTO{m}, nil
	}
	return []ports.ClassMembershipDTO{}, nil
}

func loginToken(t *testing.T, r http.Handler, username string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"`+username+`","password":"pass123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode login response failed: %v", err)
	}
	return resp.Data.Token
}

func TestClassJoinReviewFeedbackIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	classes := newFakeClassService()
	h := handler.New(handler.Handlers{Auth: fakeClassAuthService{}, Classes: classes})
	r := gin.New()
	r.Use(gin.Recovery())
	Register(r, h, config.Config{JWTSecret: "test-secret"})

	studentToken := loginToken(t, r, "student")
	teacherToken := loginToken(t, r, "teacher")

	joinReq := httptest.NewRequest(http.MethodPost, "/api/v1/classes/join", strings.NewReader(`{"class_code":"CLS-1"}`))
	joinReq.Header.Set("Content-Type", "application/json")
	joinReq.Header.Set("Authorization", "Bearer "+studentToken)
	joinRec := httptest.NewRecorder()
	r.ServeHTTP(joinRec, joinReq)
	if joinRec.Code != http.StatusOK {
		t.Fatalf("join failed status=%d body=%s", joinRec.Code, joinRec.Body.String())
	}

	dupReq := httptest.NewRequest(http.MethodPost, "/api/v1/classes/join", strings.NewReader(`{"class_code":"CLS-1"}`))
	dupReq.Header.Set("Content-Type", "application/json")
	dupReq.Header.Set("Authorization", "Bearer "+studentToken)
	dupRec := httptest.NewRecorder()
	r.ServeHTTP(dupRec, dupReq)
	if dupRec.Code != http.StatusConflict {
		t.Fatalf("expected duplicate join conflict, got %d body=%s", dupRec.Code, dupRec.Body.String())
	}

	reviewForbiddenReq := httptest.NewRequest(http.MethodPost, "/api/v1/classes/memberships/review", strings.NewReader(`{"membership_id":"m-student-1","approve":true}`))
	reviewForbiddenReq.Header.Set("Content-Type", "application/json")
	reviewForbiddenReq.Header.Set("Authorization", "Bearer "+studentToken)
	reviewForbiddenRec := httptest.NewRecorder()
	r.ServeHTTP(reviewForbiddenRec, reviewForbiddenReq)
	if reviewForbiddenRec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden review by student, got %d body=%s", reviewForbiddenRec.Code, reviewForbiddenRec.Body.String())
	}

	reviewReq := httptest.NewRequest(http.MethodPost, "/api/v1/classes/memberships/review", strings.NewReader(`{"membership_id":"m-student-1","approve":true}`))
	reviewReq.Header.Set("Content-Type", "application/json")
	reviewReq.Header.Set("Authorization", "Bearer "+teacherToken)
	reviewRec := httptest.NewRecorder()
	r.ServeHTTP(reviewRec, reviewReq)
	if reviewRec.Code != http.StatusOK {
		t.Fatalf("approve review failed status=%d body=%s", reviewRec.Code, reviewRec.Body.String())
	}

	myReq := httptest.NewRequest(http.MethodGet, "/api/v1/classes/memberships/self", nil)
	myReq.Header.Set("Authorization", "Bearer "+studentToken)
	myRec := httptest.NewRecorder()
	r.ServeHTTP(myRec, myReq)
	if myRec.Code != http.StatusOK {
		t.Fatalf("list self memberships failed status=%d body=%s", myRec.Code, myRec.Body.String())
	}
	if !strings.Contains(myRec.Body.String(), `"status":"active"`) {
		t.Fatalf("expected active status in membership list, body=%s", myRec.Body.String())
	}
}
