// Layer: Ports (接口契约层)
// Responsibility: 定义系统中所有服务的接口契约、DTO(数据传输对象)、错误码
// Dependency: 不依赖任何其他层，是系统的最内层契约
// Note: Handler和Usecase都依赖此层，实现解耦
package ports

import "context"

type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (UserDTO, error)
	Login(ctx context.Context, req LoginRequest) (LoginResponse, error)
	Verify(ctx context.Context) (VerifyResponse, error)
	SendPasswordResetCode(ctx context.Context, req PasswordResetCodeRequest) (PasswordResetCodeResponse, error)
	ResetPassword(ctx context.Context, req PasswordResetRequest) error
}

type UsersService interface {
	GetProfile(ctx context.Context, userID string) (UserDTO, error)
	UploadAvatar(ctx context.Context, userID string, req UploadAvatarRequest) (AvatarUploadResponse, error)
	UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (UserDTO, error)
	ListRanking(ctx context.Context, req RankingQuery) ([]RankingItem, error)
}

type ClassesService interface {
	CreateClass(ctx context.Context, req CreateClassRequest) (ClassDTO, error)
	UpdateClass(ctx context.Context, req UpdateClassRequest) (ClassDTO, error)
	ArchiveClass(ctx context.Context, classID string, actorUserID string) error
	ApplyJoinClass(ctx context.Context, req ApplyJoinClassRequest) (ClassMembershipDTO, error)
	ReviewMembership(ctx context.Context, req ReviewMembershipRequest) (ClassMembershipDTO, error)
	ListClassMemberships(ctx context.Context, classID string, actorUserID string) ([]ClassMembershipDTO, error)
	ListMyMemberships(ctx context.Context, userID string) ([]ClassMembershipDTO, error)
}

type ProblemsService interface {
	ListProblems(ctx context.Context, req ProblemsQuery) ([]ProblemDTO, error)
	GetProblem(ctx context.Context, problemID int64) (ProblemDTO, error)
	GetProblemForJudge(ctx context.Context, problemID int64) (ProblemDTO, error)
	CreateProblem(ctx context.Context, req CreateProblemRequest) (ProblemDTO, error)
	UpdateProblem(ctx context.Context, req UpdateProblemRequest) (ProblemDTO, error)
	DeleteProblem(ctx context.Context, req DeleteProblemRequest) error
}

type TestcasesService interface {
	CreateTestcase(ctx context.Context, req CreateTestcaseRequest) (TestcaseDTO, error)
	ListTestcases(ctx context.Context, req ListTestcasesRequest) ([]TestcaseDTO, error)
	UpdateTestcase(ctx context.Context, req UpdateTestcaseRequest) (TestcaseDTO, error)
	DeleteTestcase(ctx context.Context, req DeleteTestcaseRequest) error
	ReorderTestcases(ctx context.Context, req ReorderTestcasesRequest) ([]TestcaseDTO, error)
	ListTestcasesForJudge(ctx context.Context, req JudgeListTestcasesRequest) ([]TestcaseDTO, error)
}

type CompetitionsService interface {
	ListContests(ctx context.Context, req ContestsQuery) ([]ContestDTO, error)
	GetContest(ctx context.Context, contestID int64) (ContestDTO, error)
	CreateContest(ctx context.Context, req CreateContestRequest) (ContestDTO, error)
	UpdateContest(ctx context.Context, req UpdateContestRequest) (ContestDTO, error)
	DeleteContest(ctx context.Context, req DeleteContestRequest) error
	ListContestProblems(ctx context.Context, req ContestProblemsQuery) ([]ContestProblemBindingDTO, error)
	ReplaceContestProblems(ctx context.Context, req ReplaceContestProblemsRequest) ([]ContestProblemBindingDTO, error)
	GetContestMembership(ctx context.Context, req ContestMembershipQuery) (ContestMembershipDTO, error)
	ListContestParticipants(ctx context.Context, req ContestParticipantsQuery) ([]ContestParticipantDetailDTO, error)
	QuitContest(ctx context.Context, req QuitContestRequest) (ContestParticipantDTO, error)
	JoinContest(ctx context.Context, req JoinContestRequest) (ContestParticipantDTO, error)
	GetScoreboard(ctx context.Context, req ContestScoreboardQuery) (ContestScoreboardResponse, error)
}

type DiscussionsService interface {
	ListDiscussions(ctx context.Context, req DiscussionsQuery) ([]DiscussionDTO, error)
	GetDiscussion(ctx context.Context, discussionID string) (DiscussionDTO, error)
	CreateDiscussion(ctx context.Context, req CreateDiscussionRequest) (DiscussionDTO, error)
	CreateComment(ctx context.Context, req CreateCommentRequest) (CommentDTO, error)
	ListComments(ctx context.Context, req CommentsQuery) ([]CommentDTO, error)
	DeleteDiscussion(ctx context.Context, req DeleteDiscussionRequest) error
	DeleteComment(ctx context.Context, req DeleteCommentRequest) error
}

type SubmitRecordsService interface {
	CreateSubmission(ctx context.Context, req CreateSubmissionRequest) (SubmissionDTO, error)
	ListSubmissions(ctx context.Context, req SubmissionsQuery) ([]SubmissionDTO, error)
	MarkSubmissionJudging(ctx context.Context, submissionID int64, source string) (SubmissionDTO, error)
	WritebackSubmission(ctx context.Context, req JudgeWritebackRequest) (SubmissionDTO, error)
}

type AdminService interface {
	ListUsers(ctx context.Context, req AdminUsersQuery) ([]UserDTO, error)
	UpdateUserStatus(ctx context.Context, req UpdateUserStatusRequest) (UserDTO, error)
	UpdateUserRole(ctx context.Context, req UpdateUserRoleRequest) (UserDTO, error)
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type VerifyResponse struct {
	User         UserDTO          `json:"user"`
	Capabilities AuthCapabilities `json:"capabilities"`
}

type AuthCapabilities struct {
	PasswordResetEnabled bool `json:"password_reset_enabled"`
}

type PasswordResetCodeRequest struct {
	Email string `json:"email"`
}

type PasswordResetCodeResponse struct {
	ExpiresInSeconds int `json:"expires_in_seconds"`
}

type PasswordResetRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

type UserDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Synopsis string `json:"synopsis"`
	Score    int    `json:"score"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

type UpdateProfileRequest struct {
	Avatar   *string `json:"avatar"`
	Synopsis *string `json:"synopsis"`
}

type UploadAvatarRequest struct {
	Filename    string
	ContentType string
	Content     []byte
}

type AvatarUploadResponse struct {
	Avatar      string `json:"avatar"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type RankingQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type RankingItem struct {
	Rank int `json:"rank"`
	UserDTO
}

type CreateClassRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	OwnerUserID string `json:"owner_user_id"`
}

type UpdateClassRequest struct {
	ClassID     string `json:"class_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ActorUserID string `json:"-"`
}

type ApplyJoinClassRequest struct {
	ClassCode string `json:"class_code"`
	UserID    string `json:"user_id"`
}

type ReviewMembershipRequest struct {
	MembershipID string `json:"membership_id"`
	Approve      bool   `json:"approve"`
	ActorUserID  string `json:"-"`
}

type ClassDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	OwnerUserID string `json:"owner_user_id"`
	Status      string `json:"status"`
}

type ClassMembershipDTO struct {
	ID          string `json:"id"`
	ClassID     string `json:"class_id"`
	UserID      string `json:"user_id"`
	RoleInClass string `json:"role_in_class"`
	Status      string `json:"status"`
}

type CreateTestcaseRequest struct {
	ProblemID   int64  `json:"problem_id"`
	InputData   string `json:"input_data"`
	OutputData  string `json:"output_data"`
	IsSample    bool   `json:"is_sample"`
	ActorUserID string `json:"-"`
	ActorRole   string `json:"-"`
}

type ListTestcasesRequest struct {
	ProblemID   int64  `json:"problem_id"`
	ActorUserID string `json:"-"`
	ActorRole   string `json:"-"`
}

type UpdateTestcaseRequest struct {
	ProblemID   int64  `json:"problem_id"`
	TestcaseID  string `json:"testcase_id"`
	InputData   string `json:"input_data"`
	OutputData  string `json:"output_data"`
	IsSample    bool   `json:"is_sample"`
	ActorUserID string `json:"-"`
	ActorRole   string `json:"-"`
}

type DeleteTestcaseRequest struct {
	ProblemID   int64  `json:"problem_id"`
	TestcaseID  string `json:"testcase_id"`
	ActorUserID string `json:"-"`
	ActorRole   string `json:"-"`
}

type ReorderTestcasesRequest struct {
	ProblemID   int64    `json:"problem_id"`
	TestcaseIDs []string `json:"testcase_ids"`
	ActorUserID string   `json:"-"`
	ActorRole   string   `json:"-"`
}

type JudgeListTestcasesRequest struct {
	ProblemID int64 `json:"problem_id"`
}

type TestcaseDTO struct {
	ID         string `json:"id"`
	ProblemID  int64  `json:"problem_id"`
	InputData  string `json:"input_data"`
	OutputData string `json:"output_data"`
	IsSample   bool   `json:"is_sample"`
	SortOrder  int    `json:"sort_order"`
}

type ProblemsQuery struct {
	Page        int    `form:"page"`
	Limit       int    `form:"limit"`
	Visibility  string `form:"visibility"`
	Status      string `form:"status"`
	ActorUserID string `json:"-"`
	ActorRole   string `json:"-"`
}

type CreateProblemRequest struct {
	Title         string `json:"title"`
	Content       string `json:"content"`
	Input         string `json:"input"`
	Output        string `json:"output"`
	Difficulty    int    `json:"difficulty"`
	TimeLimitMS   int    `json:"time_limit_ms"`
	MemoryLimitMB int    `json:"memory_limit_mb"`
	ClassID       string `json:"class_id"`
	Visibility    string `json:"visibility"`
	Status        string `json:"status"`
	ActorUserID   string `json:"-"`
	ActorRole     string `json:"-"`
}

type UpdateProblemRequest struct {
	ProblemID     int64  `json:"problem_id"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	Input         string `json:"input"`
	Output        string `json:"output"`
	Difficulty    int    `json:"difficulty"`
	TimeLimitMS   int    `json:"time_limit_ms"`
	MemoryLimitMB int    `json:"memory_limit_mb"`
	ClassID       string `json:"class_id"`
	Visibility    string `json:"visibility"`
	Status        string `json:"status"`
	ActorUserID   string `json:"-"`
	ActorRole     string `json:"-"`
}

type DeleteProblemRequest struct {
	ProblemID   int64  `json:"problem_id"`
	ActorUserID string `json:"-"`
	ActorRole   string `json:"-"`
}

type ProblemDTO struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	Input         string `json:"input"`
	Output        string `json:"output"`
	Difficulty    int    `json:"difficulty"`
	TimeLimitMS   int    `json:"time_limit_ms"`
	MemoryLimitMB int    `json:"memory_limit_mb"`
	OwnerUserID   string `json:"owner_user_id"`
	ClassID       string `json:"class_id"`
	Visibility    string `json:"visibility"`
	Status        string `json:"status"`
}

type ContestsQuery struct {
	Page        int    `form:"page"`
	Limit       int    `form:"limit"`
	Visibility  string `form:"visibility"`
	RuleType    string `form:"rule_type"`
	Status      string `form:"status"`
	ActorUserID string `json:"-"`
	ActorRole   string `json:"-"`
}

type CreateContestRequest struct {
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Description  string `json:"description"`
	Announcement string `json:"announcement"`
	ClassID      string `json:"class_id"`
	Visibility   string `json:"visibility"`
	RuleType     string `json:"rule_type"`
	Status       string `json:"status"`
	IsEncrypted  bool   `json:"is_encrypted"`
	Password     string `json:"password"`
	StartAt      string `json:"start_at"`
	EndAt        string `json:"end_at"`
	ActorUserID  string `json:"-"`
	ActorRole    string `json:"-"`
}

type UpdateContestRequest struct {
	ContestID    int64  `json:"contest_id"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Description  string `json:"description"`
	Announcement string `json:"announcement"`
	ClassID      string `json:"class_id"`
	Visibility   string `json:"visibility"`
	RuleType     string `json:"rule_type"`
	Status       string `json:"status"`
	IsEncrypted  bool   `json:"is_encrypted"`
	Password     string `json:"password"`
	StartAt      string `json:"start_at"`
	EndAt        string `json:"end_at"`
	ActorUserID  string `json:"-"`
	ActorRole    string `json:"-"`
}

type DeleteContestRequest struct {
	ContestID   int64  `json:"contest_id"`
	ActorUserID string `json:"-"`
	ActorRole   string `json:"-"`
}

type ContestDTO struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Subtitle     string `json:"subtitle"`
	Description  string `json:"description"`
	Announcement string `json:"announcement"`
	OwnerUserID  string `json:"owner_user_id"`
	ClassID      string `json:"class_id"`
	Visibility   string `json:"visibility"`
	RuleType     string `json:"rule_type"`
	Status       string `json:"status"`
	IsEncrypted  bool   `json:"is_encrypted"`
	StartAt      string `json:"start_at"`
	EndAt        string `json:"end_at"`
}

type JoinContestRequest struct {
	ContestID int64  `json:"contest_id"`
	UserID    string `json:"user_id"`
	Password  string `json:"password"`
	ActorRole string `json:"-"`
}

type ContestScoreboardQuery struct {
	ContestID int64 `json:"contest_id"`
}

type ContestProblemsQuery struct {
	ContestID   int64  `json:"contest_id"`
	ActorUserID string `json:"-"`
	ActorRole   string `json:"-"`
}

type ReplaceContestProblemsRequest struct {
	ContestID   int64                         `json:"contest_id"`
	Items       []ContestProblemBindingUpsert `json:"items"`
	ActorUserID string                        `json:"-"`
	ActorRole   string                        `json:"-"`
}

type ContestProblemBindingUpsert struct {
	ProblemID    int64  `json:"problem_id"`
	DisplayOrder int    `json:"display_order"`
	Alias        string `json:"alias"`
}

type ContestProblemBindingDTO struct {
	ContestID    int64  `json:"contest_id"`
	ProblemID    int64  `json:"problem_id"`
	DisplayOrder int    `json:"display_order"`
	Alias        string `json:"alias"`
}

type ContestScoreboardResponse struct {
	ContestID     int64                   `json:"contest_id"`
	FreezeActive  bool                    `json:"freeze_active"`
	FreezeStartAt string                  `json:"freeze_start_at"`
	GeneratedAt   string                  `json:"generated_at"`
	VisibleItems  []ContestScoreboardItem `json:"visible_items"`
}

type ContestScoreboardItem struct {
	Rank           int    `json:"rank"`
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	Solved         int    `json:"solved"`
	TotalScore     int    `json:"total_score"`
	PenaltyMinutes int    `json:"penalty_minutes"`
	ReachedAt      string `json:"reached_at"`
}

type ContestParticipantDTO struct {
	ID        string `json:"id"`
	ContestID int64  `json:"contest_id"`
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
}

type ContestParticipantDetailDTO struct {
	ID        string `json:"id"`
	ContestID int64  `json:"contest_id"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	Status    string `json:"status"`
	JoinedAt  string `json:"joined_at"`
}

type ContestMembershipQuery struct {
	ContestID   int64  `json:"contest_id"`
	ActorUserID string `json:"-"`
	ActorRole   string `json:"-"`
}

type ContestMembershipDTO struct {
	ContestID int64  `json:"contest_id"`
	UserID    string `json:"user_id"`
	Joined    bool   `json:"joined"`
	Status    string `json:"status"`
}

type ContestParticipantsQuery struct {
	ContestID   int64  `json:"contest_id"`
	ActorUserID string `json:"-"`
	ActorRole   string `json:"-"`
}

type QuitContestRequest struct {
	ContestID   int64  `json:"contest_id"`
	ActorUserID string `json:"-"`
	ActorRole   string `json:"-"`
}

type DiscussionsQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type DiscussionDTO struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	CreatedAt string `json:"created_at"`
}

type CreateDiscussionRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	UserID  string `json:"user_id"`
}

type CommentDTO struct {
	ID           string `json:"id"`
	DiscussionID string `json:"discussion_id"`
	Content      string `json:"content"`
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Avatar       string `json:"avatar"`
	Profanity    bool   `json:"profanity"`
	CreatedAt    string `json:"created_at"`
}

type CreateCommentRequest struct {
	DiscussionID string `json:"discussion_id"`
	Content      string `json:"content"`
	UserID       string `json:"user_id"`
}

type CommentsQuery struct {
	DiscussionID string `json:"discussion_id"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
}

type DeleteDiscussionRequest struct {
	DiscussionID string `json:"discussion_id"`
	ActorUserID  string `json:"-"`
	ActorRole    string `json:"-"`
}

type DeleteCommentRequest struct {
	CommentID   string `json:"comment_id"`
	ActorUserID string `json:"-"`
	ActorRole   string `json:"-"`
}

type CreateSubmissionRequest struct {
	ProblemID  int64  `json:"problem_id"`
	ContestID  int64  `json:"contest_id"`
	Language   string `json:"language"`
	SourceCode string `json:"source_code"`
	UserID     string `json:"user_id"`
}

const JudgeWritebackContractV1 = "v1"

type JudgeWritebackRequest struct {
	ContractVersion string `json:"contract_version"`
	SubmissionID    int64  `json:"submission_id"`
	Result          string `json:"result"`
	Score           *int   `json:"score"`
	Source          string `json:"source"`
}

type SubmissionsQuery struct {
	UserID    string `form:"user_id"`
	ProblemID int64  `form:"problem_id"`
	ContestID int64  `form:"contest_id"`
	Page      int    `form:"page"`
	Limit     int    `form:"limit"`
}

type SubmissionDTO struct {
	ID          int64  `json:"id"`
	UserID      string `json:"user_id"`
	ProblemID   int64  `json:"problem_id"`
	ContestID   int64  `json:"contest_id"`
	Language    string `json:"language"`
	Result      string `json:"result"`
	Score       int    `json:"score"`
	SubmittedAt string `json:"submitted_at"`
}

type AdminUsersQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type UpdateUserStatusRequest struct {
	UserID      string `json:"user_id"`
	Status      string `json:"status"`
	ActorUserID string `json:"-"`
}

type UpdateUserRoleRequest struct {
	UserID      string `json:"user_id"`
	Role        string `json:"role"`
	ActorUserID string `json:"-"`
}
