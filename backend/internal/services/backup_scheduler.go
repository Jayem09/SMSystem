package services

import (
	"log"

	"github.com/robfig/cron/v3"

	"smsystem-backend/internal/config"
)

// BackupScheduler manages automatic backup scheduling
type BackupScheduler struct {
	cron          *cron.Cron
	backupService *BackupService
	cfg           *config.Config
}

// NewBackupScheduler creates a new backup scheduler
func NewBackupScheduler(cfg *config.Config, backupService *BackupService) *BackupScheduler {
	c := cron.New()

	scheduler := &BackupScheduler{
		cron:          c,
		backupService: backupService,
		cfg:           cfg,
	}

	// Schedule the backup job
	// Cron format: "0 2 * * *" = At 02:00 every day
	_, err := c.AddFunc(cfg.AutoBackupCron, func() {
		log.Println("[BackupScheduler] Starting scheduled backup...")
		backup, err := scheduler.backupService.RunAutoBackupWithConfig(cfg)
		if err != nil {
			log.Printf("[BackupScheduler] Auto backup failed: %v\n", err)
		} else {
			log.Printf("[BackupScheduler] Auto backup completed: %s (%.2f KB)",
				backup.Filename, float64(backup.Size)/1024)
		}
	})

	if err != nil {
		log.Printf("[BackupScheduler] Failed to schedule backup: %v\n", err)
	} else {
		log.Printf("[BackupScheduler] Scheduled backup with cron: %s", cfg.AutoBackupCron)
	}

	return scheduler
}

// Start starts the scheduler
func (s *BackupScheduler) Start() {
	s.cron.Start()
	log.Println("[BackupScheduler] Backup scheduler started")
}

// Stop stops the scheduler
func (s *BackupScheduler) Stop() error {
	ctx := s.cron.Stop()
	<-ctx.Done()
	return nil
}

// GetNextRun returns the next scheduled run time
func (s *BackupScheduler) GetNextRun() string {
	return s.cfg.AutoBackupCron
}

// IsEnabled returns whether auto backup is enabled
func (s *BackupScheduler) IsEnabled() bool {
	return s.cfg.AutoBackupEnabled
}
