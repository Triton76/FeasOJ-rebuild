package competitions

import (
	"FeasOJ/app/backend-rebuild/internal/observability"
	"FeasOJ/app/backend-rebuild/internal/ports"
	"FeasOJ/app/backend-rebuild/internal/security"
	passwordutil "FeasOJ/pkg/auth"
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct {
	repo  Repository
	nowFn func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, nowFn: time.Now}
}

func (s *Service) ListContests(ctx context.Context, req ports.ContestsQuery) ([]ports.ContestDTO, error) {
	if s.repo == nil {
		return nil, ports.ErrNotImplemented
	}
	page := req.Page
	if page <= 0 {
		page = defaultPage
	}
	limit := req.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset := (page - 1) * limit
	items, err := s.repo.ListVisible(
		ctx,
		offset,
		limit,
		strings.TrimSpace(req.Visibility),
		strings.TrimSpace(req.RuleType),
		strings.TrimSpace(req.Status),
		strings.TrimSpace(req.ActorUserID),
		strings.TrimSpace(req.ActorRole),
	)
	if err != nil {
		return nil, err
	}

	resp := make([]ports.ContestDTO, 0, len(items))
	for _, item := range items {
		resp = append(resp, toContestDTO(item))
	}
	return resp, nil
}

func (s *Service) GetContest(ctx context.Context, contestID int64) (ports.ContestDTO, error) {
	if s.repo == nil {
		return ports.ContestDTO{}, ports.ErrNotImplemented
	}
	if contestID <= 0 {
		return ports.ContestDTO{}, ports.ErrInvalidArgument
	}
	claims, _ := security.ClaimsFromContext(ctx)
	actorUserID := ""
	actorRole := ""
	if claims.UserID != "" {
		actorUserID = claims.UserID
		actorRole = claims.Role
	}

	c, err := s.repo.GetVisibleByID(ctx, contestID, actorUserID, actorRole)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ContestDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.ContestDTO{}, err
	}
	return toContestDTO(c), nil
}

func (s *Service) CreateContest(ctx context.Context, req ports.CreateContestRequest) (ports.ContestDTO, error) {
	if s.repo == nil {
		return ports.ContestDTO{}, ports.ErrNotImplemented
	}
	if !canManageContest(req.ActorRole) {
		return ports.ContestDTO{}, ports.ErrForbidden
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return ports.ContestDTO{}, ports.ErrInvalidArgument
	}
	visibility := normalizeContestVisibility(req.Visibility)
	ruleType := normalizeContestRuleType(req.RuleType)
	status := normalizeContestStatus(req.Status)
	if visibility == "" || ruleType == "" || status == "" {
		return ports.ContestDTO{}, ports.ErrInvalidArgument
	}
	// Note: 创建时不检查题目绑定，因为题目可以稍后绑定
	// 只在更新状态时（从draft发布）才检查题目绑定
	startAt, endAt, err := parseContestTimeRange(req.StartAt, req.EndAt)
	if err != nil {
		return ports.ContestDTO{}, ports.ErrInvalidArgument
	}
	passwordHash, err := normalizeContestPassword(req.IsEncrypted, req.Password)
	if err != nil {
		return ports.ContestDTO{}, ports.ErrInvalidArgument
	}

	now := time.Now().UTC()
	created, err := s.repo.Create(ctx, Contest{
		Title:        title,
		Subtitle:     strings.TrimSpace(req.Subtitle),
		Description:  strings.TrimSpace(req.Description),
		Announcement: strings.TrimSpace(req.Announcement),
		OwnerUserID:  strings.TrimSpace(req.ActorUserID),
		ClassID:      strings.TrimSpace(req.ClassID),
		Visibility:   visibility,
		RuleType:     ruleType,
		Status:       status,
		IsEncrypted:  req.IsEncrypted,
		PasswordHash: passwordHash,
		StartAt:      startAt,
		EndAt:        endAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return ports.ContestDTO{}, err
	}
	return toContestDTO(created), nil
}

func (s *Service) UpdateContest(ctx context.Context, req ports.UpdateContestRequest) (ports.ContestDTO, error) {
	if s.repo == nil {
		return ports.ContestDTO{}, ports.ErrNotImplemented
	}
	if req.ContestID <= 0 {
		return ports.ContestDTO{}, ports.ErrInvalidArgument
	}
	if !canManageContest(req.ActorRole) {
		return ports.ContestDTO{}, ports.ErrForbidden
	}

	existing, err := s.repo.GetByID(ctx, req.ContestID)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ContestDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.ContestDTO{}, err
	}
	if req.ActorRole != "admin" && existing.OwnerUserID != strings.TrimSpace(req.ActorUserID) {
		return ports.ContestDTO{}, ports.ErrForbidden
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return ports.ContestDTO{}, ports.ErrInvalidArgument
	}
	visibility := normalizeContestVisibility(req.Visibility)
	ruleType := normalizeContestRuleType(req.RuleType)
	status := normalizeContestStatus(req.Status)
	if visibility == "" || ruleType == "" || status == "" {
		return ports.ContestDTO{}, ports.ErrInvalidArgument
	}
	if requiresBoundProblems(status, visibility) {
		boundCount, countErr := s.repo.CountProblemBindings(ctx, existing.ID)
		if countErr != nil {
			return ports.ContestDTO{}, countErr
		}
		if boundCount == 0 {
			return ports.ContestDTO{}, ports.ErrConflict
		}
	}
	startAt, endAt, err := parseContestTimeRange(req.StartAt, req.EndAt)
	if err != nil {
		return ports.ContestDTO{}, ports.ErrInvalidArgument
	}
	passwordHash, err := normalizeContestPassword(req.IsEncrypted, req.Password)
	if err != nil {
		return ports.ContestDTO{}, ports.ErrInvalidArgument
	}
	if req.IsEncrypted && strings.TrimSpace(req.Password) == "" {
		passwordHash = existing.PasswordHash
		if strings.TrimSpace(passwordHash) == "" {
			return ports.ContestDTO{}, ports.ErrInvalidArgument
		}
	}

	existing.Title = title
	existing.Subtitle = strings.TrimSpace(req.Subtitle)
	existing.Description = strings.TrimSpace(req.Description)
	existing.Announcement = strings.TrimSpace(req.Announcement)
	existing.ClassID = strings.TrimSpace(req.ClassID)
	existing.Visibility = visibility
	existing.RuleType = ruleType
	existing.Status = status
	existing.IsEncrypted = req.IsEncrypted
	existing.PasswordHash = passwordHash
	existing.StartAt = startAt
	existing.EndAt = endAt
	existing.UpdatedAt = time.Now().UTC()

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return ports.ContestDTO{}, err
	}
	return toContestDTO(updated), nil
}

func (s *Service) DeleteContest(ctx context.Context, req ports.DeleteContestRequest) error {
	if s.repo == nil {
		return ports.ErrNotImplemented
	}
	if req.ContestID <= 0 {
		return ports.ErrInvalidArgument
	}
	if !canManageContest(req.ActorRole) {
		return ports.ErrForbidden
	}

	existing, err := s.repo.GetByID(ctx, req.ContestID)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ErrNotFound
	}
	if err != nil {
		return err
	}
	if req.ActorRole != "admin" && existing.OwnerUserID != strings.TrimSpace(req.ActorUserID) {
		return ports.ErrForbidden
	}

	return s.repo.Delete(ctx, req.ContestID)
}

func (s *Service) ListContestProblems(ctx context.Context, req ports.ContestProblemsQuery) ([]ports.ContestProblemBindingDTO, error) {
	if s.repo == nil {
		return nil, ports.ErrNotImplemented
	}
	if req.ContestID <= 0 {
		return nil, ports.ErrInvalidArgument
	}

	contest, err := s.repo.GetByID(ctx, req.ContestID)
	if errors.Is(err, ports.ErrNotFound) {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	actorRole := strings.TrimSpace(req.ActorRole)
	actorUserID := strings.TrimSpace(req.ActorUserID)
	if actorRole != "admin" && contest.OwnerUserID != actorUserID {
		membership, membershipErr := s.repo.GetParticipantByContestAndUser(ctx, req.ContestID, actorUserID)
		if membershipErr != nil {
			return nil, ports.ErrForbidden
		}
		if membership.Status == "quit" || membership.Status == "finished" {
			return nil, ports.ErrForbidden
		}
	}

	items, err := s.repo.ListProblemBindings(ctx, req.ContestID)
	if err != nil {
		return nil, err
	}
	resp := make([]ports.ContestProblemBindingDTO, 0, len(items))
	for _, item := range items {
		resp = append(resp, ports.ContestProblemBindingDTO{
			ContestID:    item.ContestID,
			ProblemID:    item.ProblemID,
			DisplayOrder: item.DisplayOrder,
			Alias:        item.Alias,
		})
	}
	return resp, nil
}

func (s *Service) ReplaceContestProblems(ctx context.Context, req ports.ReplaceContestProblemsRequest) ([]ports.ContestProblemBindingDTO, error) {
	if s.repo == nil {
		return nil, ports.ErrNotImplemented
	}
	if req.ContestID <= 0 {
		return nil, ports.ErrInvalidArgument
	}

	contest, err := s.repo.GetByID(ctx, req.ContestID)
	if errors.Is(err, ports.ErrNotFound) {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.ActorRole) != "admin" && contest.OwnerUserID != strings.TrimSpace(req.ActorUserID) {
		return nil, ports.ErrForbidden
	}

	bindings, err := normalizeContestProblemBindings(req.ContestID, req.Items)
	if err != nil {
		return nil, err
	}

	problemIDs := make([]int64, 0, len(bindings))
	for _, item := range bindings {
		problemIDs = append(problemIDs, item.ProblemID)
	}
	if len(problemIDs) > 0 {
		count, countErr := s.repo.CountExistingProblems(ctx, problemIDs)
		if countErr != nil {
			return nil, countErr
		}
		if count != int64(len(problemIDs)) {
			return nil, ports.ErrNotFound
		}
	}

	if err := s.repo.ReplaceProblemBindings(ctx, req.ContestID, bindings); err != nil {
		return nil, err
	}
	stored, err := s.repo.ListProblemBindings(ctx, req.ContestID)
	if err != nil {
		return nil, err
	}

	resp := make([]ports.ContestProblemBindingDTO, 0, len(stored))
	for _, item := range stored {
		resp = append(resp, ports.ContestProblemBindingDTO{
			ContestID:    item.ContestID,
			ProblemID:    item.ProblemID,
			DisplayOrder: item.DisplayOrder,
			Alias:        item.Alias,
		})
	}
	return resp, nil
}

func (s *Service) GetContestMembership(ctx context.Context, req ports.ContestMembershipQuery) (ports.ContestMembershipDTO, error) {
	if s.repo == nil {
		return ports.ContestMembershipDTO{}, ports.ErrNotImplemented
	}
	if req.ContestID <= 0 || strings.TrimSpace(req.ActorUserID) == "" {
		return ports.ContestMembershipDTO{}, ports.ErrInvalidArgument
	}

	if _, err := s.repo.GetVisibleByID(ctx, req.ContestID, strings.TrimSpace(req.ActorUserID), strings.TrimSpace(req.ActorRole)); err != nil {
		return ports.ContestMembershipDTO{}, err
	}

	p, err := s.repo.GetParticipantByContestAndUser(ctx, req.ContestID, strings.TrimSpace(req.ActorUserID))
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ContestMembershipDTO{ContestID: req.ContestID, UserID: strings.TrimSpace(req.ActorUserID), Joined: false, Status: ""}, nil
	}
	if err != nil {
		return ports.ContestMembershipDTO{}, err
	}
	return ports.ContestMembershipDTO{ContestID: req.ContestID, UserID: p.UserID, Joined: true, Status: p.Status}, nil
}

func (s *Service) ListContestParticipants(ctx context.Context, req ports.ContestParticipantsQuery) ([]ports.ContestParticipantDetailDTO, error) {
	if s.repo == nil {
		return nil, ports.ErrNotImplemented
	}
	if req.ContestID <= 0 || strings.TrimSpace(req.ActorUserID) == "" {
		return nil, ports.ErrInvalidArgument
	}

	contest, err := s.repo.GetByID(ctx, req.ContestID)
	if errors.Is(err, ports.ErrNotFound) {
		return nil, ports.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	actorRole := strings.TrimSpace(req.ActorRole)
	actorUserID := strings.TrimSpace(req.ActorUserID)
	if actorRole != "admin" && contest.OwnerUserID != actorUserID {
		membership, mErr := s.repo.GetParticipantByContestAndUser(ctx, req.ContestID, actorUserID)
		if mErr != nil {
			return nil, ports.ErrForbidden
		}
		if membership.Status == "quit" || membership.Status == "finished" {
			return nil, ports.ErrForbidden
		}
	}

	items, err := s.repo.ListParticipantsByContest(ctx, req.ContestID)
	if err != nil {
		return nil, err
	}

	resp := make([]ports.ContestParticipantDetailDTO, 0, len(items))
	for _, item := range items {
		joinedAt := ""
		if item.JoinedAt != nil {
			joinedAt = item.JoinedAt.UTC().Format(time.RFC3339)
		}
		resp = append(resp, ports.ContestParticipantDetailDTO{
			ID:        item.ID,
			ContestID: item.ContestID,
			UserID:    item.UserID,
			Username:  item.Username,
			Avatar:    item.Avatar,
			Status:    item.Status,
			JoinedAt:  joinedAt,
		})
	}
	return resp, nil
}

func (s *Service) QuitContest(ctx context.Context, req ports.QuitContestRequest) (ports.ContestParticipantDTO, error) {
	if s.repo == nil {
		return ports.ContestParticipantDTO{}, ports.ErrNotImplemented
	}
	if req.ContestID <= 0 || strings.TrimSpace(req.ActorUserID) == "" {
		return ports.ContestParticipantDTO{}, ports.ErrInvalidArgument
	}

	if _, err := s.repo.GetVisibleByID(ctx, req.ContestID, strings.TrimSpace(req.ActorUserID), strings.TrimSpace(req.ActorRole)); err != nil {
		return ports.ContestParticipantDTO{}, err
	}

	current, err := s.repo.GetParticipantByContestAndUser(ctx, req.ContestID, strings.TrimSpace(req.ActorUserID))
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ContestParticipantDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.ContestParticipantDTO{}, err
	}
	if current.Status == "quit" || current.Status == "finished" {
		return ports.ContestParticipantDTO{}, ports.ErrConflict
	}

	updated, err := s.repo.UpdateParticipantStatus(ctx, req.ContestID, strings.TrimSpace(req.ActorUserID), current.Status, "quit", time.Now().UTC())
	if err != nil {
		return ports.ContestParticipantDTO{}, err
	}
	return ports.ContestParticipantDTO{ID: updated.ID, ContestID: updated.ContestID, UserID: updated.UserID, Status: updated.Status}, nil
}

func (s *Service) JoinContest(ctx context.Context, req ports.JoinContestRequest) (ports.ContestParticipantDTO, error) {
	if s.repo == nil {
		return ports.ContestParticipantDTO{}, ports.ErrNotImplemented
	}
	if req.ContestID <= 0 || strings.TrimSpace(req.UserID) == "" {
		return ports.ContestParticipantDTO{}, ports.ErrInvalidArgument
	}

	contest, err := s.repo.GetVisibleByID(ctx, req.ContestID, strings.TrimSpace(req.UserID), strings.TrimSpace(req.ActorRole))
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ContestParticipantDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.ContestParticipantDTO{}, err
	}

	if contest.IsEncrypted {
		if !passwordutil.VerifyPassword(strings.TrimSpace(req.Password), contest.PasswordHash) {
			return ports.ContestParticipantDTO{}, ports.ErrForbidden
		}
	}

	now := time.Now().UTC()
	p, err := s.repo.CreateParticipant(ctx, Participant{
		ID:        uuid.NewString(),
		ContestID: req.ContestID,
		UserID:    strings.TrimSpace(req.UserID),
		Status:    "registered",
		JoinedAt:  &now,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return ports.ContestParticipantDTO{}, err
	}
	return ports.ContestParticipantDTO{ID: p.ID, ContestID: p.ContestID, UserID: p.UserID, Status: p.Status}, nil
}

func (s *Service) GetScoreboard(ctx context.Context, req ports.ContestScoreboardQuery) (ports.ContestScoreboardResponse, error) {
	startAt := time.Now()
	success := false
	defer func() {
		observability.ObserveScoreboardQueryLatency(time.Since(startAt), success)
	}()

	if s.repo == nil {
		return ports.ContestScoreboardResponse{}, ports.ErrNotImplemented
	}
	if req.ContestID <= 0 {
		return ports.ContestScoreboardResponse{}, ports.ErrInvalidArgument
	}

	claims, _ := security.ClaimsFromContext(ctx)
	contest, err := s.repo.GetVisibleByID(ctx, req.ContestID, strings.TrimSpace(claims.UserID), strings.TrimSpace(claims.Role))
	if err != nil {
		return ports.ContestScoreboardResponse{}, err
	}

	now := s.nowFn().UTC()
	cutoff := now
	freezeActive := false
	freezeStartAt := ""

	if contest.EndAt != nil {
		contestEnd := contest.EndAt.UTC()
		freezeStart := contestEnd.Add(-60 * time.Minute)
		freezeStartAt = freezeStart.Format(time.RFC3339)
		switch {
		case now.Before(freezeStart):
			cutoff = now
		case now.Before(contestEnd):
			cutoff = freezeStart
			freezeActive = true
		default:
			cutoff = contestEnd
		}
	}

	submissions, err := s.repo.ListScoreboardSubmissions(ctx, req.ContestID, cutoff)
	if err != nil {
		return ports.ContestScoreboardResponse{}, err
	}

	items := buildScoreboardItemsByRuleType(submissions, contest.StartAt, contest.RuleType)
	observability.LogJSON("scoreboard.query", map[string]any{
		"contest_id":    req.ContestID,
		"rule_type":     strings.TrimSpace(contest.RuleType),
		"freeze_active": freezeActive,
		"items":         len(items),
	})
	success = true

	return ports.ContestScoreboardResponse{
		ContestID:     req.ContestID,
		FreezeActive:  freezeActive,
		FreezeStartAt: freezeStartAt,
		GeneratedAt:   now.Format(time.RFC3339),
		VisibleItems:  items,
	}, nil
}

type problemBucket struct {
	wrongAttempts int
	ceAttempts    int
	solved        bool
	acceptedAt    time.Time
}

type userAgg struct {
	userID      string
	username    string
	solved      int
	penalty     int
	reachedAt   time.Time
	problemStat map[int64]*problemBucket
}

func buildScoreboardItemsByRuleType(submissions []ScoreboardSubmission, contestStart *time.Time, ruleType string) []ports.ContestScoreboardItem {
	if strings.TrimSpace(ruleType) == "oi" {
		return buildOIScoreboardItems(submissions)
	}
	return buildACMScoreboardItems(submissions, contestStart)
}

func buildACMScoreboardItems(submissions []ScoreboardSubmission, contestStart *time.Time) []ports.ContestScoreboardItem {
	byUser := make(map[string]*userAgg)

	for _, sub := range submissions {
		agg, ok := byUser[sub.UserID]
		if !ok {
			agg = &userAgg{
				userID:      sub.UserID,
				username:    sub.Username,
				problemStat: make(map[int64]*problemBucket),
			}
			byUser[sub.UserID] = agg
		}

		bucket, ok := agg.problemStat[sub.ProblemID]
		if !ok {
			bucket = &problemBucket{}
			agg.problemStat[sub.ProblemID] = bucket
		}

		if bucket.solved {
			continue
		}

		switch strings.TrimSpace(sub.Result) {
		case "accepted":
			bucket.solved = true
			bucket.acceptedAt = sub.SubmittedAt.UTC()
		case "compile_error":
			bucket.ceAttempts++
		case "pending", "judging":
			// ignore non-terminal states in scoreboard aggregation
		default:
			bucket.wrongAttempts++
		}
	}

	for _, agg := range byUser {
		for _, bucket := range agg.problemStat {
			if !bucket.solved {
				continue
			}
			agg.solved++
			attemptPenalty := bucket.wrongAttempts + bucket.ceAttempts
			agg.penalty += attemptPenalty * 20
			if contestStart != nil {
				minutes := int(bucket.acceptedAt.Sub(contestStart.UTC()).Minutes())
				if minutes > 0 {
					agg.penalty += minutes
				}
			}
			if agg.reachedAt.IsZero() || bucket.acceptedAt.After(agg.reachedAt) {
				agg.reachedAt = bucket.acceptedAt
			}
		}
	}

	sorted := make([]*userAgg, 0, len(byUser))
	for _, agg := range byUser {
		sorted = append(sorted, agg)
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		a := sorted[i]
		b := sorted[j]
		if a.solved != b.solved {
			return a.solved > b.solved
		}
		if a.penalty != b.penalty {
			return a.penalty < b.penalty
		}
		if !a.reachedAt.Equal(b.reachedAt) {
			if a.reachedAt.IsZero() {
				return false
			}
			if b.reachedAt.IsZero() {
				return true
			}
			return a.reachedAt.Before(b.reachedAt)
		}
		return a.userID < b.userID
	})

	items := make([]ports.ContestScoreboardItem, 0, len(sorted))
	for i, row := range sorted {
		reachedAt := ""
		if !row.reachedAt.IsZero() {
			reachedAt = row.reachedAt.UTC().Format(time.RFC3339)
		}
		items = append(items, ports.ContestScoreboardItem{
			Rank:           i + 1,
			UserID:         row.userID,
			Username:       row.username,
			Solved:         row.solved,
			TotalScore:     0,
			PenaltyMinutes: row.penalty,
			ReachedAt:      reachedAt,
		})
	}
	return items
}

type oiProblemBucket struct {
	bestScore int
	bestAt    time.Time
}

type oiUserAgg struct {
	userID      string
	username    string
	totalScore  int
	reachedAt   time.Time
	problemStat map[int64]*oiProblemBucket
}

func buildOIScoreboardItems(submissions []ScoreboardSubmission) []ports.ContestScoreboardItem {
	byUser := make(map[string]*oiUserAgg)

	for _, sub := range submissions {
		result := strings.TrimSpace(sub.Result)
		if result == "pending" || result == "judging" {
			continue
		}

		agg, ok := byUser[sub.UserID]
		if !ok {
			agg = &oiUserAgg{
				userID:      sub.UserID,
				username:    sub.Username,
				problemStat: make(map[int64]*oiProblemBucket),
			}
			byUser[sub.UserID] = agg
		}

		bucket, ok := agg.problemStat[sub.ProblemID]
		if !ok {
			bucket = &oiProblemBucket{}
			agg.problemStat[sub.ProblemID] = bucket
		}

		if sub.Score > bucket.bestScore || (sub.Score == bucket.bestScore && (bucket.bestAt.IsZero() || sub.SubmittedAt.UTC().Before(bucket.bestAt))) {
			bucket.bestScore = sub.Score
			bucket.bestAt = sub.SubmittedAt.UTC()
		}
	}

	for _, agg := range byUser {
		for _, bucket := range agg.problemStat {
			agg.totalScore += bucket.bestScore
			if !bucket.bestAt.IsZero() && (agg.reachedAt.IsZero() || bucket.bestAt.After(agg.reachedAt)) {
				agg.reachedAt = bucket.bestAt
			}
		}
	}

	sorted := make([]*oiUserAgg, 0, len(byUser))
	for _, agg := range byUser {
		sorted = append(sorted, agg)
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		a := sorted[i]
		b := sorted[j]
		if a.totalScore != b.totalScore {
			return a.totalScore > b.totalScore
		}
		if !a.reachedAt.Equal(b.reachedAt) {
			if a.reachedAt.IsZero() {
				return false
			}
			if b.reachedAt.IsZero() {
				return true
			}
			return a.reachedAt.Before(b.reachedAt)
		}
		return a.userID < b.userID
	})

	items := make([]ports.ContestScoreboardItem, 0, len(sorted))
	for i, row := range sorted {
		reachedAt := ""
		if !row.reachedAt.IsZero() {
			reachedAt = row.reachedAt.UTC().Format(time.RFC3339)
		}
		items = append(items, ports.ContestScoreboardItem{
			Rank:           i + 1,
			UserID:         row.userID,
			Username:       row.username,
			Solved:         0,
			TotalScore:     row.totalScore,
			PenaltyMinutes: 0,
			ReachedAt:      reachedAt,
		})
	}
	return items
}

func canManageContest(role string) bool {
	role = strings.TrimSpace(role)
	return role == "teacher" || role == "admin"
}

func normalizeContestVisibility(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case "public", "class", "private":
		return value
	default:
		return ""
	}
}

func normalizeContestRuleType(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case "acm", "oi", "assignment":
		return value
	default:
		return ""
	}
}

func normalizeContestStatus(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case "draft", "scheduled", "running", "ended":
		return value
	default:
		return ""
	}
}

func requiresBoundProblems(status, visibility string) bool {
	return status != "draft" && visibility != "private"
}

func normalizeContestProblemBindings(contestID int64, items []ports.ContestProblemBindingUpsert) ([]ContestProblemBinding, error) {
	now := time.Now().UTC()
	problemSet := make(map[int64]struct{}, len(items))
	orderSet := make(map[int]struct{}, len(items))
	aliasSet := make(map[string]struct{}, len(items))
	resp := make([]ContestProblemBinding, 0, len(items))

	for _, item := range items {
		if item.ProblemID <= 0 || item.DisplayOrder <= 0 {
			return nil, ports.ErrInvalidArgument
		}
		if _, ok := problemSet[item.ProblemID]; ok {
			return nil, ports.ErrConflict
		}
		if _, ok := orderSet[item.DisplayOrder]; ok {
			return nil, ports.ErrConflict
		}
		alias := strings.TrimSpace(item.Alias)
		if alias != "" {
			if _, ok := aliasSet[alias]; ok {
				return nil, ports.ErrConflict
			}
			aliasSet[alias] = struct{}{}
		}

		problemSet[item.ProblemID] = struct{}{}
		orderSet[item.DisplayOrder] = struct{}{}
		resp = append(resp, ContestProblemBinding{
			ContestID:    contestID,
			ProblemID:    item.ProblemID,
			DisplayOrder: item.DisplayOrder,
			Alias:        alias,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
	}

	return resp, nil
}

func normalizeContestPassword(isEncrypted bool, plain string) (string, error) {
	if !isEncrypted {
		return "", nil
	}
	plain = strings.TrimSpace(plain)
	if plain == "" {
		return "", nil
	}
	hash := passwordutil.EncryptPassword(plain)
	if hash == "" {
		return "", errors.New("hash password failed")
	}
	return hash, nil
}

func parseContestTimeRange(startAtText, endAtText string) (*time.Time, *time.Time, error) {
	startAtText = strings.TrimSpace(startAtText)
	endAtText = strings.TrimSpace(endAtText)
	if startAtText == "" && endAtText == "" {
		return nil, nil, nil
	}
	if startAtText == "" || endAtText == "" {
		return nil, nil, errors.New("invalid time range")
	}

	startAt, err := time.Parse(time.RFC3339, startAtText)
	if err != nil {
		return nil, nil, err
	}
	endAt, err := time.Parse(time.RFC3339, endAtText)
	if err != nil {
		return nil, nil, err
	}
	if !startAt.Before(endAt) {
		return nil, nil, errors.New("start_at must be before end_at")
	}
	startAt = startAt.UTC()
	endAt = endAt.UTC()
	return &startAt, &endAt, nil
}

func toContestDTO(c Contest) ports.ContestDTO {
	startAt := ""
	if c.StartAt != nil {
		startAt = c.StartAt.UTC().Format(time.RFC3339)
	}
	endAt := ""
	if c.EndAt != nil {
		endAt = c.EndAt.UTC().Format(time.RFC3339)
	}
	return ports.ContestDTO{
		ID:           c.ID,
		Title:        c.Title,
		Subtitle:     c.Subtitle,
		Description:  c.Description,
		Announcement: c.Announcement,
		OwnerUserID:  c.OwnerUserID,
		ClassID:      c.ClassID,
		Visibility:   c.Visibility,
		RuleType:     c.RuleType,
		Status:       c.Status,
		IsEncrypted:  c.IsEncrypted,
		StartAt:      startAt,
		EndAt:        endAt,
	}
}
