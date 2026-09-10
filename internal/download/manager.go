package download

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"audivo-media-downloader/internal/models"
)

type JobHandler func(
	ctx context.Context,
	jobID string,
	emit func(models.DownloadEvent),
	log func(models.TechnicalLogLine),
) (models.DownloadResult, error)

type EventSink func(models.DownloadEvent)
type LogSink func(models.TechnicalLogLine)

type activeJob struct {
	id      string
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{}
	handler JobHandler
}

type JobManager struct {
	mu       sync.Mutex
	active   *activeJob
	queue    []*activeJob
	sequence atomic.Uint64
	onEvent  EventSink
	onLog    LogSink
}

func NewJobManager(onEvent EventSink, onLog LogSink) *JobManager {
	return &JobManager{onEvent: onEvent, onLog: onLog}
}

func (m *JobManager) Start(parent context.Context, handler JobHandler) (*models.DownloadJob, *models.UserError) {
	if parent == nil {
		parent = context.Background()
	}
	m.mu.Lock()
	jobID := m.newID()
	ctx, cancel := context.WithCancel(parent)
	job := &activeJob{id: jobID, ctx: ctx, cancel: cancel, done: make(chan struct{}), handler: handler}
	startNow := m.active == nil
	if startNow {
		m.active = job
	} else {
		m.queue = append(m.queue, job)
	}
	m.mu.Unlock()

	if startNow {
		m.start(job)
		return &models.DownloadJob{ID: jobID, State: models.StatePreparing, Message: "Preparando o download…"}, nil
	}
	m.emit(models.DownloadEvent{JobID: jobID, State: models.StateQueued, Message: "Download aguardando na fila…"})
	return &models.DownloadJob{ID: jobID, State: models.StateQueued, Message: "Download aguardando na fila…"}, nil
}

func (m *JobManager) start(job *activeJob) {
	m.emit(models.DownloadEvent{JobID: job.id, State: models.StatePreparing, Stage: models.StagePreparing, Message: "Preparando o download…"})
	go m.run(job)
}

func (m *JobManager) run(job *activeJob) {
	defer close(job.done)
	defer func() {
		var next *activeJob
		m.mu.Lock()
		if m.active != nil && m.active.id == job.id {
			m.active = nil
			if len(m.queue) > 0 {
				next = m.queue[0]
				m.queue = m.queue[1:]
				m.active = next
			}
		}
		m.mu.Unlock()
		if next != nil {
			m.start(next)
		}
	}()

	result, err := job.handler(job.ctx, job.id, m.emit, m.log)
	if err == nil && job.ctx.Err() == nil {
		state := models.StateCompleted
		message := "Download concluído."
		if result.Partial {
			state = models.StatePartial
			message = fmt.Sprintf("Coleção parcialmente concluída: %d concluídos, %d falhos.", result.CompletedItems, result.FailedItems)
		} else if result.CompletedItems > 0 {
			state = models.StateCollectionCompleted
			message = fmt.Sprintf("Coleção concluída: %d itens.", result.CompletedItems)
		}
		m.emit(models.DownloadEvent{
			JobID:          job.id,
			State:          state,
			Stage:          models.StageCompleted,
			Percent:        100,
			OverallPercent: 100,
			Message:        message,
			Result:         &result,
		})
		return
	}
	if job.ctx.Err() != nil || errors.Is(err, context.Canceled) {
		event := models.DownloadEvent{JobID: job.id, State: models.StateCancelled, Message: "Download cancelado."}
		if result.CompletedItems > 0 || result.FailedItems > 0 {
			event.Result = &result
			event.Message = fmt.Sprintf("Download cancelado: %d itens preservados.", result.CompletedItems)
		}
		m.emit(event)
		return
	}
	if err == nil {
		err = errors.New("download failed without a reason")
	}
	m.emit(models.DownloadEvent{JobID: job.id, State: models.StateError, Error: AsUserError(err), Message: AsUserError(err).Message})
}

func (m *JobManager) Cancel(jobID string) *models.UserError {
	m.mu.Lock()
	job := m.active
	if job != nil && job.id == jobID {
		job.cancel()
		m.mu.Unlock()
		return nil
	}
	for index, queued := range m.queue {
		if queued.id != jobID {
			continue
		}
		m.queue = append(m.queue[:index], m.queue[index+1:]...)
		queued.cancel()
		close(queued.done)
		m.mu.Unlock()
		m.emit(models.DownloadEvent{JobID: jobID, State: models.StateCancelled, Message: "Download cancelado na fila."})
		return nil
	}
	m.mu.Unlock()
	return &models.UserError{Code: "job_not_active", Message: "Este download não está mais ativo.", Retryable: false}
}

func (m *JobManager) IsActive() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active != nil
}

func (m *JobManager) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	m.mu.Lock()
	job := m.active
	queued := m.queue
	m.queue = nil
	if job != nil {
		job.cancel()
	}
	m.mu.Unlock()
	for _, pending := range queued {
		pending.cancel()
		close(pending.done)
		m.emit(models.DownloadEvent{JobID: pending.id, State: models.StateCancelled, Message: "Download cancelado na fila."})
	}
	if job == nil {
		return nil
	}
	select {
	case <-job.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *JobManager) emit(event models.DownloadEvent) {
	if event.JobID == "" {
		return
	}
	if m.onEvent != nil {
		m.onEvent(event)
	}
}

func (m *JobManager) log(line models.TechnicalLogLine) {
	if m.onLog != nil && line.JobID != "" {
		m.onLog(line)
	}
}

func (m *JobManager) newID() string {
	sequence := m.sequence.Add(1)
	return fmt.Sprintf("job-%d-%d", time.Now().UnixNano(), sequence)
}
