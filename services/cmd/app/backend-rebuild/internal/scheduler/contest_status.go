package scheduler

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"
)

func StartContestStatusReconciler(ctx context.Context, db *gorm.DB, interval time.Duration) {
	if db == nil || interval <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			if err := ReconcileContestStatus(ctx, db); err != nil {
				log.Printf("[backend-rebuild] reconcile contest status failed: %v", err)
			}

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func ReconcileContestStatus(ctx context.Context, db *gorm.DB) error {
	now := time.Now().UTC()

	if err := db.WithContext(ctx).Exec(`
		UPDATE contests
		SET status = 'scheduled', updated_at = ?
		WHERE status <> 'draft'
		  AND start_at IS NOT NULL
		  AND end_at IS NOT NULL
		  AND ? < start_at
		  AND status <> 'scheduled'
	`, now, now).Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Exec(`
		UPDATE contests
		SET status = 'running', updated_at = ?
		WHERE status <> 'draft'
		  AND start_at IS NOT NULL
		  AND end_at IS NOT NULL
		  AND start_at <= ?
		  AND ? < end_at
		  AND status <> 'running'
	`, now, now, now).Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Exec(`
		UPDATE contests
		SET status = 'ended', updated_at = ?
		WHERE status <> 'draft'
		  AND end_at IS NOT NULL
		  AND end_at <= ?
		  AND status <> 'ended'
	`, now, now).Error; err != nil {
		return err
	}

	return nil
}
