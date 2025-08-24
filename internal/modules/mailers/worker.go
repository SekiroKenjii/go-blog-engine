package mailers

import (
	"context"
	"sync"
	"time"

	"github.com/SekiroKenjii/go-blog-engine/pkg/logger"
	"go.uber.org/zap"
)

type MailJob struct {
	ToEmail      string
	StrategyName string
	Params       map[string]any
	Attempts     int
	CreatedAt    time.Time
}

type MailWorker struct {
	jobQueue    chan MailJob
	workerPool  chan chan MailJob
	quit        chan bool
	wg          sync.WaitGroup
	mailer      IMailSender
	workerCount int
	maxRetries  int
}

// Worker represents an individual worker using strategy pattern
type Worker struct {
	id         int
	jobChannel chan MailJob
	workerPool chan chan MailJob
	quit       chan bool
	mailer     IMailSender
	maxRetries int
}

func NewMailWorker(mailer IMailSender, workerCount, queueSize, maxRetries int) *MailWorker {
	return &MailWorker{
		jobQueue:    make(chan MailJob, queueSize),
		workerPool:  make(chan chan MailJob, workerCount),
		quit:        make(chan bool),
		mailer:      mailer,
		workerCount: workerCount,
		maxRetries:  maxRetries,
	}
}

// Start starts the mail worker
func (ew *MailWorker) Start() {
	logger.Info("Starting mail worker", zap.Int("workers", ew.workerCount))

	for i := 1; i <= ew.workerCount; i++ {
		worker := &Worker{
			id:         i,
			jobChannel: make(chan MailJob),
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

// Stop stops the mail worker gracefully
func (ew *MailWorker) Stop() {
	logger.Info("Stopping mail worker...")

	close(ew.quit)
	ew.wg.Wait()

	logger.Info("Mail worker stopped")
}

// AddJob adds a job to the queue
func (ew *MailWorker) AddJob(job MailJob) bool {
	select {
	case ew.jobQueue <- job:
		return true
	default:
		logger.Warn("Mail job queue is full, job dropped",
			zap.String("to", job.ToEmail),
			zap.String("strategy", job.StrategyName))
		return false
	}
}

// dispatch dispatches jobs to available workers
func (ew *MailWorker) dispatch() {
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
func (w *Worker) start(wg *sync.WaitGroup) {
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

// processJob processes an individual mail job
func (w *Worker) processJob(job MailJob) {
	logger.Debug("Processing mail job",
		zap.Int("worker", w.id),
		zap.String("to", job.ToEmail),
		zap.String("strategy", job.StrategyName),
		zap.Int("attempts", job.Attempts))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := w.mailer.Send(ctx, job.StrategyName, job.ToEmail, job.Params)

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
				logger.Info("Retrying mail job",
					zap.String("to", job.ToEmail),
					zap.String("strategy", job.StrategyName),
					zap.Int("attempts", job.Attempts))
			default:
				logger.Error("Failed to requeue mail job - queue full",
					zap.String("to", job.ToEmail),
					zap.String("strategy", job.StrategyName))
			}
		} else {
			logger.Error("Mail job failed after maximum retries",
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
func (ew *MailWorker) GetJobQueue() chan<- MailJob {
	return ew.jobQueue
}
