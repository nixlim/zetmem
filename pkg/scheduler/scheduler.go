package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zetmem/mcp-server/pkg/config"
	"github.com/zetmem/mcp-server/pkg/memory"
	"github.com/zetmem/mcp-server/pkg/models"
	"go.uber.org/zap"
)

// Scheduler manages scheduled tasks for memory evolution
type Scheduler struct {
	config       config.EvolutionConfig
	logger       *zap.Logger
	evolutionMgr *memory.EvolutionManager
	jobs         map[string]*Job
	running      bool
	mu           sync.RWMutex
	stopChan     chan struct{}
	eventChan    chan Event
}

// Job represents a scheduled job
type Job struct {
	ID         string
	Name       string
	Schedule   string // Cron expression
	JobType    JobType
	Config     JobConfig
	LastRun    time.Time
	NextRun    time.Time
	Enabled    bool
	RunCount   int64
	ErrorCount int64
	LastError  error
}

// JobType represents the type of scheduled job
type JobType string

const (
	JobTypeEvolution   JobType = "evolution"
	JobTypeCleanup     JobType = "cleanup"
	JobTypeMaintenance JobType = "maintenance"
)

// JobConfig holds job-specific configuration
type JobConfig struct {
	EvolutionConfig *EvolutionJobConfig `json:"evolution_config,omitempty"`
	CleanupConfig   *CleanupJobConfig   `json:"cleanup_config,omitempty"`
}

// EvolutionJobConfig holds evolution job configuration
type EvolutionJobConfig struct {
	Scope       string `json:"scope"`
	MaxMemories int    `json:"max_memories"`
	ProjectPath string `json:"project_path,omitempty"`
}

// CleanupJobConfig holds cleanup job configuration
type CleanupJobConfig struct {
	MaxAge      time.Duration `json:"max_age"`
	MaxMemories int           `json:"max_memories"`
}

// Event represents a scheduler event
type Event struct {
	Type      EventType
	JobID     string
	Timestamp time.Time
	Data      interface{}
}

// EventType represents the type of scheduler event
type EventType string

const (
	EventJobStarted   EventType = "job_started"
	EventJobCompleted EventType = "job_completed"
	EventJobFailed    EventType = "job_failed"
	EventJobScheduled EventType = "job_scheduled"
)

// NewScheduler creates a new scheduler
func NewScheduler(cfg config.EvolutionConfig, evolutionMgr *memory.EvolutionManager, logger *zap.Logger) *Scheduler {
	return &Scheduler{
		config:       cfg,
		logger:       logger,
		evolutionMgr: evolutionMgr,
		jobs:         make(map[string]*Job),
		stopChan:     make(chan struct{}),
		eventChan:    make(chan Event, 100),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("Starting scheduler")

	// Add default evolution job if enabled
	if s.config.Enabled {
		err := s.AddJob(&Job{
			ID:       "default_evolution",
			Name:     "Default Memory Evolution",
			Schedule: s.config.Schedule,
			JobType:  JobTypeEvolution,
			Config: JobConfig{
				EvolutionConfig: &EvolutionJobConfig{
					Scope:       "recent",
					MaxMemories: s.config.BatchSize,
				},
			},
			Enabled: true,
		})
		if err != nil {
			s.logger.Error("Failed to add default evolution job", zap.Error(err))
		}
	}

	// Start the main scheduler loop
	go s.run(ctx)

	// Start event processor
	go s.processEvents(ctx)

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.logger.Info("Stopping scheduler")
	s.running = false
	close(s.stopChan)
}

// AddJob adds a new job to the scheduler
func (s *Scheduler) AddJob(job *Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[job.ID]; exists {
		return fmt.Errorf("job with ID %s already exists", job.ID)
	}

	// Parse and validate schedule
	nextRun, err := s.parseSchedule(job.Schedule)
	if err != nil {
		return fmt.Errorf("invalid schedule %s: %w", job.Schedule, err)
	}

	job.NextRun = nextRun
	s.jobs[job.ID] = job

	s.logger.Info("Job added",
		zap.String("id", job.ID),
		zap.String("name", job.Name),
		zap.Time("next_run", job.NextRun))

	// Emit event
	s.eventChan <- Event{
		Type:      EventJobScheduled,
		JobID:     job.ID,
		Timestamp: time.Now(),
		Data:      job,
	}

	return nil
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[jobID]; !exists {
		return fmt.Errorf("job with ID %s not found", jobID)
	}

	delete(s.jobs, jobID)
	s.logger.Info("Job removed", zap.String("id", jobID))

	return nil
}

// GetJob returns a job by ID
func (s *Scheduler) GetJob(jobID string) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job with ID %s not found", jobID)
	}

	return job, nil
}

// ListJobs returns all jobs
func (s *Scheduler) ListJobs() []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]*Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}

	return jobs
}

// TriggerJob manually triggers a job
func (s *Scheduler) TriggerJob(jobID string) error {
	s.mu.RLock()
	job, exists := s.jobs[jobID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("job with ID %s not found", jobID)
	}

	if !job.Enabled {
		return fmt.Errorf("job %s is disabled", jobID)
	}

	go s.executeJob(context.Background(), job)
	return nil
}

// run is the main scheduler loop
func (s *Scheduler) run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndRunJobs(ctx)
		}
	}
}

// checkAndRunJobs checks for jobs that need to run
func (s *Scheduler) checkAndRunJobs(ctx context.Context) {
	s.mu.RLock()
	now := time.Now()
	jobsToRun := make([]*Job, 0)

	for _, job := range s.jobs {
		if job.Enabled && now.After(job.NextRun) {
			jobsToRun = append(jobsToRun, job)
		}
	}
	s.mu.RUnlock()

	// Execute jobs
	for _, job := range jobsToRun {
		go s.executeJob(ctx, job)
	}
}

// executeJob executes a single job
func (s *Scheduler) executeJob(ctx context.Context, job *Job) {
	s.logger.Info("Executing job",
		zap.String("id", job.ID),
		zap.String("name", job.Name))

	start := time.Now()

	// Emit start event
	s.eventChan <- Event{
		Type:      EventJobStarted,
		JobID:     job.ID,
		Timestamp: start,
	}

	// Execute based on job type
	var err error
	switch job.JobType {
	case JobTypeEvolution:
		err = s.executeEvolutionJob(ctx, job)
	case JobTypeCleanup:
		err = s.executeCleanupJob(ctx, job)
	case JobTypeMaintenance:
		err = s.executeMaintenanceJob(ctx, job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.JobType)
	}

	duration := time.Since(start)

	// Update job status
	s.mu.Lock()
	job.LastRun = start
	job.RunCount++
	if err != nil {
		job.ErrorCount++
		job.LastError = err
	} else {
		job.LastError = nil
	}

	// Schedule next run
	nextRun, scheduleErr := s.parseSchedule(job.Schedule)
	if scheduleErr != nil {
		s.logger.Error("Failed to schedule next run",
			zap.String("job_id", job.ID),
			zap.Error(scheduleErr))
	} else {
		job.NextRun = nextRun
	}
	s.mu.Unlock()

	// Emit completion event
	eventType := EventJobCompleted
	if err != nil {
		eventType = EventJobFailed
	}

	s.eventChan <- Event{
		Type:      eventType,
		JobID:     job.ID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"duration": duration,
			"error":    err,
		},
	}

	if err != nil {
		s.logger.Error("Job execution failed",
			zap.String("id", job.ID),
			zap.Duration("duration", duration),
			zap.Error(err))
	} else {
		s.logger.Info("Job execution completed",
			zap.String("id", job.ID),
			zap.Duration("duration", duration))
	}
}

// executeEvolutionJob executes an evolution job
func (s *Scheduler) executeEvolutionJob(ctx context.Context, job *Job) error {
	if job.Config.EvolutionConfig == nil {
		return fmt.Errorf("evolution config is required for evolution job")
	}

	config := job.Config.EvolutionConfig
	request := models.EvolveNetworkRequest{
		TriggerType: "scheduled",
		Scope:       config.Scope,
		MaxMemories: config.MaxMemories,
		ProjectPath: config.ProjectPath,
	}

	_, err := s.evolutionMgr.EvolveNetwork(ctx, request)
	return err
}

// executeCleanupJob executes a cleanup job
func (s *Scheduler) executeCleanupJob(ctx context.Context, job *Job) error {
	// Placeholder for cleanup logic
	s.logger.Info("Executing cleanup job", zap.String("job_id", job.ID))
	return nil
}

// executeMaintenanceJob executes a maintenance job
func (s *Scheduler) executeMaintenanceJob(ctx context.Context, job *Job) error {
	// Placeholder for maintenance logic
	s.logger.Info("Executing maintenance job", zap.String("job_id", job.ID))
	return nil
}

// processEvents processes scheduler events
func (s *Scheduler) processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-s.eventChan:
			s.handleEvent(event)
		}
	}
}

// handleEvent handles a scheduler event
func (s *Scheduler) handleEvent(event Event) {
	s.logger.Debug("Scheduler event",
		zap.String("type", string(event.Type)),
		zap.String("job_id", event.JobID),
		zap.Time("timestamp", event.Timestamp))

	// Here you could add webhook notifications, metrics updates, etc.
}

// parseSchedule parses a cron schedule and returns the next run time
func (s *Scheduler) parseSchedule(schedule string) (time.Time, error) {
	// Simple implementation - in production, use a proper cron parser
	// For now, just support basic intervals
	switch schedule {
	case "0 2 * * *": // Daily at 2 AM
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 2, 0, 0, 0, now.Location())
		if now.Hour() < 2 {
			next = time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
		}
		return next, nil
	case "0 */6 * * *": // Every 6 hours
		return time.Now().Add(6 * time.Hour), nil
	case "0 * * * *": // Every hour
		return time.Now().Add(1 * time.Hour), nil
	default:
		// Default to 1 hour from now
		return time.Now().Add(1 * time.Hour), nil
	}
}
