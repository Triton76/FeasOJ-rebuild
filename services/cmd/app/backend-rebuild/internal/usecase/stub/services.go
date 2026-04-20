// Layer: Usecase Stub (业务逻辑占位层)
// Responsibility: 为未实现的服务提供默认占位实现，支持渐进式重构
// Dependency: 依赖 Ports 层，实现 Ports 定义的服务接口
package stub

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
)

type Services struct {
	Auth          ports.AuthService
	Users         ports.UsersService
	Classes       ports.ClassesService
	Problems      ports.ProblemsService
	Competitions  ports.CompetitionsService
	Discussions   ports.DiscussionsService
	SubmitRecords ports.SubmitRecordsService
	Admin         ports.AdminService
}

func NewServices() Services {
	return Services{
		Auth:          AuthService{},
		Users:         UsersService{},
		Classes:       ClassesService{},
		Problems:      ProblemsService{},
		Competitions:  CompetitionsService{},
		Discussions:   DiscussionsService{},
		SubmitRecords: SubmitRecordsService{},
		Admin:         AdminService{},
	}
}

type AuthService struct{}

func (AuthService) Register(context.Context, ports.RegisterRequest) (ports.UserDTO, error) {
	return ports.UserDTO{}, ports.ErrNotImplemented
}

func (AuthService) Login(context.Context, ports.LoginRequest) (ports.LoginResponse, error) {
	return ports.LoginResponse{}, ports.ErrNotImplemented
}

func (AuthService) Verify(context.Context) (ports.VerifyResponse, error) {
	return ports.VerifyResponse{}, ports.ErrNotImplemented
}

type UsersService struct{}

func (UsersService) GetProfile(context.Context, string) (ports.UserDTO, error) {
	return ports.UserDTO{}, ports.ErrNotImplemented
}

func (UsersService) UpdateProfile(context.Context, string, ports.UpdateProfileRequest) (ports.UserDTO, error) {
	return ports.UserDTO{}, ports.ErrNotImplemented
}

func (UsersService) ListRanking(context.Context, ports.RankingQuery) ([]ports.RankingItem, error) {
	return nil, ports.ErrNotImplemented
}

type ClassesService struct{}

func (ClassesService) CreateClass(context.Context, ports.CreateClassRequest) (ports.ClassDTO, error) {
	return ports.ClassDTO{}, ports.ErrNotImplemented
}

func (ClassesService) UpdateClass(context.Context, ports.UpdateClassRequest) (ports.ClassDTO, error) {
	return ports.ClassDTO{}, ports.ErrNotImplemented
}

func (ClassesService) ArchiveClass(context.Context, string, string) error {
	return ports.ErrNotImplemented
}

func (ClassesService) ApplyJoinClass(context.Context, ports.ApplyJoinClassRequest) (ports.ClassMembershipDTO, error) {
	return ports.ClassMembershipDTO{}, ports.ErrNotImplemented
}

func (ClassesService) ReviewMembership(context.Context, ports.ReviewMembershipRequest) (ports.ClassMembershipDTO, error) {
	return ports.ClassMembershipDTO{}, ports.ErrNotImplemented
}

func (ClassesService) ListClassMemberships(context.Context, string, string) ([]ports.ClassMembershipDTO, error) {
	return nil, ports.ErrNotImplemented
}

func (ClassesService) ListMyMemberships(context.Context, string) ([]ports.ClassMembershipDTO, error) {
	return nil, ports.ErrNotImplemented
}

type ProblemsService struct{}

func (ProblemsService) ListProblems(context.Context, ports.ProblemsQuery) ([]ports.ProblemDTO, error) {
	return nil, ports.ErrNotImplemented
}

func (ProblemsService) GetProblem(context.Context, int64) (ports.ProblemDTO, error) {
	return ports.ProblemDTO{}, ports.ErrNotImplemented
}

func (ProblemsService) CreateProblem(context.Context, ports.CreateProblemRequest) (ports.ProblemDTO, error) {
	return ports.ProblemDTO{}, ports.ErrNotImplemented
}

func (ProblemsService) UpdateProblem(context.Context, ports.UpdateProblemRequest) (ports.ProblemDTO, error) {
	return ports.ProblemDTO{}, ports.ErrNotImplemented
}

func (ProblemsService) DeleteProblem(context.Context, ports.DeleteProblemRequest) error {
	return ports.ErrNotImplemented
}

type CompetitionsService struct{}

func (CompetitionsService) ListContests(context.Context, ports.ContestsQuery) ([]ports.ContestDTO, error) {
	return nil, ports.ErrNotImplemented
}

func (CompetitionsService) GetContest(context.Context, int64) (ports.ContestDTO, error) {
	return ports.ContestDTO{}, ports.ErrNotImplemented
}

func (CompetitionsService) CreateContest(context.Context, ports.CreateContestRequest) (ports.ContestDTO, error) {
	return ports.ContestDTO{}, ports.ErrNotImplemented
}

func (CompetitionsService) UpdateContest(context.Context, ports.UpdateContestRequest) (ports.ContestDTO, error) {
	return ports.ContestDTO{}, ports.ErrNotImplemented
}

func (CompetitionsService) DeleteContest(context.Context, ports.DeleteContestRequest) error {
	return ports.ErrNotImplemented
}

func (CompetitionsService) JoinContest(context.Context, ports.JoinContestRequest) (ports.ContestParticipantDTO, error) {
	return ports.ContestParticipantDTO{}, ports.ErrNotImplemented
}

func (CompetitionsService) GetScoreboard(context.Context, ports.ContestScoreboardQuery) (ports.ContestScoreboardResponse, error) {
	return ports.ContestScoreboardResponse{}, ports.ErrNotImplemented
}

type DiscussionsService struct{}

func (DiscussionsService) ListDiscussions(context.Context, ports.DiscussionsQuery) ([]ports.DiscussionDTO, error) {
	return nil, ports.ErrNotImplemented
}

func (DiscussionsService) GetDiscussion(context.Context, string) (ports.DiscussionDTO, error) {
	return ports.DiscussionDTO{}, ports.ErrNotImplemented
}

func (DiscussionsService) CreateDiscussion(context.Context, ports.CreateDiscussionRequest) (ports.DiscussionDTO, error) {
	return ports.DiscussionDTO{}, ports.ErrNotImplemented
}

func (DiscussionsService) CreateComment(context.Context, ports.CreateCommentRequest) (ports.CommentDTO, error) {
	return ports.CommentDTO{}, ports.ErrNotImplemented
}

type SubmitRecordsService struct{}

func (SubmitRecordsService) CreateSubmission(context.Context, ports.CreateSubmissionRequest) (ports.SubmissionDTO, error) {
	return ports.SubmissionDTO{}, ports.ErrNotImplemented
}

func (SubmitRecordsService) ListSubmissions(context.Context, ports.SubmissionsQuery) ([]ports.SubmissionDTO, error) {
	return nil, ports.ErrNotImplemented
}

func (SubmitRecordsService) MarkSubmissionJudging(context.Context, int64, string) (ports.SubmissionDTO, error) {
	return ports.SubmissionDTO{}, ports.ErrNotImplemented
}

func (SubmitRecordsService) WritebackSubmission(context.Context, ports.JudgeWritebackRequest) (ports.SubmissionDTO, error) {
	return ports.SubmissionDTO{}, ports.ErrNotImplemented
}

type AdminService struct{}

func (AdminService) ListUsers(context.Context, ports.AdminUsersQuery) ([]ports.UserDTO, error) {
	return nil, ports.ErrNotImplemented
}

func (AdminService) UpdateUserStatus(context.Context, ports.UpdateUserStatusRequest) (ports.UserDTO, error) {
	return ports.UserDTO{}, ports.ErrNotImplemented
}
