package judge

import (
	"FeasOJ/app/judgecore/internal/config"
	"FeasOJ/app/judgecore/internal/global"
	"FeasOJ/app/judgecore/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ProcessJudgeTasks consumes rebuild submission jobs from RabbitMQ and completes
// the full judge flow through backend-rebuild HTTP callbacks.
func ProcessJudgeTasks(rmqConfig config.RabbitMQ, backendCfg config.BackendRebuild, pool *JudgePool, codeDir string) {
	client := utils.NewBackendRebuildClient(backendCfg)

	for {
		conn, ch, err := utils.ConnectRabbitMQ(rmqConfig)
		if err != nil {
			log.Println("[FeasOJ] RabbitMQ connect error, retrying in 3s:", err)
			time.Sleep(3 * time.Second)
			continue
		}
		log.Println("[FeasOJ] RabbitMQ connected")

		if err := consumeLoop(conn, ch, rmqConfig, client, pool, codeDir); err != nil {
			log.Println("[FeasOJ] Judge consumer stopped, retrying in 3s:", err)
		}
		time.Sleep(3 * time.Second)
	}
}

func consumeLoop(conn *amqp.Connection, ch *amqp.Channel, rmqConfig config.RabbitMQ, client *utils.BackendRebuildClient, pool *JudgePool, codeDir string) error {
	defer conn.Close()
	defer ch.Close()

	mainQueue := strings.TrimSpace(rmqConfig.MainQueue)
	if mainQueue == "" {
		mainQueue = "judge.submission.main"
	}

	msgs, err := ch.Consume(
		mainQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	taskChan := make(chan amqp.Delivery)
	var wg sync.WaitGroup

	workers := pool.sandboxConfig.MaxConcurrent
	if workers <= 0 {
		workers = 1
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(taskChan, client, pool, codeDir)
		}()
	}

	for msg := range msgs {
		taskChan <- msg
	}

	close(taskChan)
	wg.Wait()
	return fmt.Errorf("rabbitmq consumer channel closed")
}

func worker(taskChan <-chan amqp.Delivery, client *utils.BackendRebuildClient, pool *JudgePool, codeDir string) {
	for msg := range taskChan {
		if err := handleDelivery(msg, client, pool, codeDir); err != nil {
			log.Printf("[FeasOJ] Judge task failed, requeueing: %v", err)
			if nackErr := msg.Nack(false, true); nackErr != nil {
				log.Printf("[FeasOJ] Failed to nack message: %v", nackErr)
			}
		}
	}
}

func handleDelivery(msg amqp.Delivery, client *utils.BackendRebuildClient, pool *JudgePool, codeDir string) error {
	var job utils.SubmissionJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		_ = msg.Ack(false)
		return fmt.Errorf("decode submission job: %w", err)
	}
	if strings.TrimSpace(job.ContractVersion) != "" && strings.TrimSpace(job.ContractVersion) != "v1" {
		_ = msg.Ack(false)
		return fmt.Errorf("unsupported contract version: %s", job.ContractVersion)
	}

	markCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	err := client.MarkSubmissionJudging(markCtx, job.SubmissionID)
	cancel()
	if err != nil {
		return fmt.Errorf("mark submission judging: %w", err)
	}

	bundleCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	bundle, err := client.FetchProblemBundle(bundleCtx, job.ProblemID)
	cancel()
	if err != nil {
		writebackCtx, writebackCancel := context.WithTimeout(context.Background(), 20*time.Second)
		writeErr := client.WritebackSubmission(writebackCtx, job.SubmissionID, global.SystemError, intPtr(0))
		writebackCancel()
		if writeErr != nil {
			return fmt.Errorf("fetch problem bundle: %w; writeback system error also failed: %v", err, writeErr)
		}
		if ackErr := msg.Ack(false); ackErr != nil {
			return fmt.Errorf("ack after bundle failure: %w", ackErr)
		}
		return nil
	}

	filename, err := persistSubmissionSource(codeDir, job)
	if err != nil {
		writebackCtx, writebackCancel := context.WithTimeout(context.Background(), 20*time.Second)
		writeErr := client.WritebackSubmission(writebackCtx, job.SubmissionID, global.SystemError, intPtr(0))
		writebackCancel()
		if writeErr != nil {
			return fmt.Errorf("persist source: %w; writeback system error also failed: %v", err, writeErr)
		}
		if ackErr := msg.Ack(false); ackErr != nil {
			return fmt.Errorf("ack after source failure: %w", ackErr)
		}
		return nil
	}

	containerID := pool.AcquireContainer()
	pool.containerIDs.Store(filename, containerID)

	result := CompileAndRun(filename, containerID, bundle.Problem, toJudgeTestcases(bundle.Testcases))

	pool.ReleaseContainer(containerID)
	pool.containerIDs.Delete(filename)

	score := 0
	if result == global.Accepted {
		score = 100
	}
	writebackCtx, writebackCancel := context.WithTimeout(context.Background(), 20*time.Second)
	err = client.WritebackSubmission(writebackCtx, job.SubmissionID, result, intPtr(score))
	writebackCancel()
	if err != nil {
		return fmt.Errorf("writeback result: %w", err)
	}
	if err := msg.Ack(false); err != nil {
		return fmt.Errorf("ack message: %w", err)
	}
	return nil
}

func persistSubmissionSource(codeDir string, job utils.SubmissionJob) (string, error) {
	ext, err := languageToExtension(job.Language)
	if err != nil {
		return "", err
	}
	filename := fmt.Sprintf("submission_%d%s", job.SubmissionID, ext)
	fullPath := filepath.Join(codeDir, filename)
	if err := os.WriteFile(fullPath, []byte(job.SourceCode), 0644); err != nil {
		return "", err
	}
	return filename, nil
}

func toJudgeTestcases(items []utils.TestcaseDTO) []*JudgeTestcase {
	resp := make([]*JudgeTestcase, 0, len(items))
	for _, item := range items {
		resp = append(resp, &JudgeTestcase{
			InputData:  item.InputData,
			OutputData: item.OutputData,
		})
	}
	return resp
}

func languageToExtension(language string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "cpp", "c++":
		return ".cpp", nil
	case "java":
		return ".java", nil
	case "python", "py":
		return ".py", nil
	case "rust", "rs":
		return ".rs", nil
	case "php":
		return ".php", nil
	case "pascal", "pas":
		return ".pas", nil
	case "golang", "go":
		return ".go", nil
	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}
}

func intPtr(v int) *int {
	return &v
}
