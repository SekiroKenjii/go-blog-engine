package mailers

import (
	"context"
	"sync"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"
	"go.uber.org/zap"
)

// EmailJob represents an email job to be processed using strategy pattern
type EmailJob struct {
	ToEmail      string
	StrategyName string
	Params       map[string]any
	Attempts     int
	CreatedAt    time.Time
}

// StrategicEmailWorker handles asynchronous email processing using strategy pattern
type StrategicEmailWorker struct {
	jobQueue    chan EmailJob
	workerPool  chan chan EmailJob
	quit        chan bool
	wg          sync.WaitGroup
	mailer      IEmailSender
	workerCount int
	maxRetries  int
}

// StrategicWorker represents an individual worker using strategy pattern
type StrategicWorker struct {
	id         int
	jobChannel chan EmailJob
	workerPool chan chan EmailJob
	quit       chan bool
	mailer     IEmailSender
	maxRetries int
}

// NewStrategicEmailWorker creates a new strategic email worker
func NewStrategicEmailWorker(mailer IEmailSender, workerCount, queueSize, maxRetries int) *StrategicEmailWorker {
	return &StrategicEmailWorker{
		jobQueue:    make(chan EmailJob, queueSize),
		workerPool:  make(chan chan EmailJob, workerCount),
		quit:        make(chan bool),
		mailer:      mailer,
		workerCount: workerCount,
		maxRetries:  maxRetries,
	}
}

// Start starts the email worker
func (ew *StrategicEmailWorker) Start() {
	logger.Info("Starting strategic email worker", zap.Int("workers", ew.workerCount))

	for i := 1; i <= ew.workerCount; i++ {
		worker := &StrategicWorker{
			id:         i,
			jobChannel: make(chan EmailJob),
			workerPool: ew.workerPool,
			quit:       make(chan bool),
			mailer:     ew.mailer,
			maxRetries: ew.maxRetries,
		}
		worker.start(&ew.wg)
	}

	ew.wg.Add(1)
	go ew.dispatch()
}

// Stop stops the email worker gracefully
func (ew *StrategicEmailWorker) Stop() {
	logger.Info("Stopping strategic email worker...")
	close(ew.quit)
	ew.wg.Wait()
	logger.Info("Strategic email worker stopped")
}

// AddJob adds a job to the queue
func (ew *StrategicEmailWorker) AddJob(job EmailJob) bool {
	select {
	case ew.jobQueue <- job:
		return true
	default:
		logger.Warn("Email job queue is full, job dropped",
			zap.String("to", job.ToEmail),
			zap.String("strategy", job.StrategyName))
		return false
	}
}

// dispatch dispatches jobs to available workers
func (ew *StrategicEmailWorker) dispatch() {
	defer ew.wg.Done()

	for {
		select {
		case job := <-ew.jobQueue:
			// Set created time if not set
			if job.CreatedAt.IsZero() {
				job.CreatedAt = time.Now()
			}

			select {
			case workerJobChannel := <-ew.workerPool:
				workerJobChannel <- job
				logger.Debug("Job dispatched to worker",
					zap.String("to", job.ToEmail),
					zap.String("strategy", job.StrategyName))
			case <-ew.quit:
				logger.Info("Dispatcher shutting down")
				return
			}

		case <-ew.quit:
			logger.Info("Dispatcher shutting down")
			return
		}
	}
}

// start starts the individual worker
func (w *StrategicWorker) start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Debug("Worker started", zap.Int("id", w.id))

		for {
			// Add worker to the pool
			w.workerPool <- w.jobChannel

			select {
			case job := <-w.jobChannel:
				w.processJob(job)

			case <-w.quit:
				logger.Debug("Worker shutting down", zap.Int("id", w.id))
				return
			}
		}
	}()
}

// processJob processes an individual email job
func (w *StrategicWorker) processJob(job EmailJob) {
	logger.Debug("Processing email job",
		zap.Int("worker", w.id),
		zap.String("to", job.ToEmail),
		zap.String("strategy", job.StrategyName),
		zap.Int("attempts", job.Attempts))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Send email using strategy pattern
	err := w.mailer.SendEmail(ctx, job.StrategyName, job.ToEmail, job.Params)

	if err != nil {
		logger.Error("Failed to send email",
			zap.Error(err),
			zap.String("to", job.ToEmail),
			zap.String("strategy", job.StrategyName),
			zap.Int("attempts", job.Attempts))

		// Retry logic
		if job.Attempts < w.maxRetries {
			job.Attempts++

			// Exponential backoff
			backoff := time.Duration(job.Attempts) * time.Second
			time.Sleep(backoff)

			// Retry by re-adding to queue
			select {
			case w.workerPool <- w.jobChannel:
				w.jobChannel <- job
				logger.Info("Retrying email job",
					zap.String("to", job.ToEmail),
					zap.String("strategy", job.StrategyName),
					zap.Int("attempts", job.Attempts))
			default:
				logger.Error("Failed to requeue email job - queue full",
					zap.String("to", job.ToEmail),
					zap.String("strategy", job.StrategyName))
			}
		} else {
			logger.Error("Email job failed after maximum retries",
				zap.String("to", job.ToEmail),
				zap.String("strategy", job.StrategyName),
				zap.Int("max_retries", w.maxRetries))
		}
	} else {
		logger.Info("Email sent successfully",
			zap.String("to", job.ToEmail),
			zap.String("strategy", job.StrategyName))
	}
}

// GetJobQueue returns the job queue (for external access)
func (ew *StrategicEmailWorker) GetJobQueue() chan<- EmailJob {
	return ew.jobQueue
}
