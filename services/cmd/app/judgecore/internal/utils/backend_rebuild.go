package utils

import (
	"FeasOJ/app/judgecore/internal/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const judgeWritebackContractV1 = "v1"

type SubmissionJob struct {
	ContractVersion string `json:"contract_version"`
	SubmissionID    int64  `json:"submission_id"`
	UserID          string `json:"user_id"`
	ProblemID       int64  `json:"problem_id"`
	ContestID       int64  `json:"contest_id"`
	Language        string `json:"language"`
	SourceCode      string `json:"source_code"`
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

type TestcaseDTO struct {
	ID         string `json:"id"`
	ProblemID  int64  `json:"problem_id"`
	InputData  string `json:"input_data"`
	OutputData string `json:"output_data"`
	IsSample   bool   `json:"is_sample"`
	SortOrder  int    `json:"sort_order"`
}

type JudgeProblemBundle struct {
	Problem   ProblemDTO    `json:"problem"`
	Testcases []TestcaseDTO `json:"testcases"`
}

type JudgeWritebackRequest struct {
	ContractVersion string `json:"contract_version"`
	SubmissionID    int64  `json:"submission_id"`
	Result          string `json:"result"`
	Score           *int   `json:"score,omitempty"`
	Source          string `json:"source"`
}

type BackendRebuildClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewBackendRebuildClient(cfg config.BackendRebuild) *BackendRebuildClient {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8082/api/v1"
	}
	return &BackendRebuildClient{
		baseURL: baseURL,
		token:   strings.TrimSpace(cfg.JudgeToken),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *BackendRebuildClient) FetchProblemBundle(ctx context.Context, problemID int64) (JudgeProblemBundle, error) {
	var resp struct {
		Data JudgeProblemBundle `json:"data"`
	}
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("%s/judge/problems/%d/bundle", c.baseURL, problemID), nil, &resp); err != nil {
		return JudgeProblemBundle{}, err
	}
	return resp.Data, nil
}

func (c *BackendRebuildClient) MarkSubmissionJudging(ctx context.Context, submissionID int64) error {
	return c.doJSON(ctx, http.MethodPost, fmt.Sprintf("%s/judge/submissions/%d/judging", c.baseURL, submissionID), map[string]any{}, nil)
}

func (c *BackendRebuildClient) WritebackSubmission(ctx context.Context, submissionID int64, result string, score *int) error {
	return c.doJSON(ctx, http.MethodPost, fmt.Sprintf("%s/judge/writeback", c.baseURL), JudgeWritebackRequest{
		ContractVersion: judgeWritebackContractV1,
		SubmissionID:    submissionID,
		Result:          result,
		Score:           score,
		Source:          "judgecore",
	}, nil)
}

func (c *BackendRebuildClient) doJSON(ctx context.Context, method, url string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(payload); err != nil {
			return err
		}
		body = buf
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("X-Judge-Token", c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		msg := strings.TrimSpace(string(data))
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("backend-rebuild %s %s failed: status=%d body=%s", method, url, resp.StatusCode, msg)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
