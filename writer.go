package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/logger"
	"github.com/nicolasbonnici/gorest/query"
)

const (
	defaultBufferCapacity = 4096
	defaultBatchSize      = 256
	defaultFlushInterval  = 250 * time.Millisecond
	defaultWriteTimeout   = 5 * time.Second
)

// batchWriterOptions tunes the async writer. Zero values fall back to the
// package defaults, so tests can override only the knobs they care about.
type batchWriterOptions struct {
	bufferCapacity int
	batchSize      int
	flushInterval  time.Duration
	writeTimeout   time.Duration
}

// batchWriter keeps metric inserts off the request hot path by buffering them
// and persisting them from a single background goroutine in portable multi-row
// batches. A metric insert on the hot path costs one channel send instead of a
// synchronous INSERT round trip (plus, previously, a follow-up SELECT).
type batchWriter struct {
	db        database.Database
	buf       chan Metric
	batchSize int
	interval  time.Duration
	timeout   time.Duration

	wg sync.WaitGroup

	// mu serialises enqueue against shutdown so the buffer is never closed
	// while a producer is mid-send (which would panic).
	mu     sync.RWMutex
	closed bool
}

func newBatchWriter(db database.Database, opts batchWriterOptions) *batchWriter {
	if opts.bufferCapacity <= 0 {
		opts.bufferCapacity = defaultBufferCapacity
	}
	if opts.batchSize <= 0 {
		opts.batchSize = defaultBatchSize
	}
	if opts.flushInterval <= 0 {
		opts.flushInterval = defaultFlushInterval
	}
	if opts.writeTimeout <= 0 {
		opts.writeTimeout = defaultWriteTimeout
	}

	w := &batchWriter{
		db:        db,
		buf:       make(chan Metric, opts.bufferCapacity),
		batchSize: opts.batchSize,
		interval:  opts.flushInterval,
		timeout:   opts.writeTimeout,
	}

	w.wg.Add(1)
	go w.run()

	return w
}

// enqueue hands a metric to the background writer. It blocks only when the
// bounded buffer is saturated, applying backpressure rather than dropping
// events. Once the writer is shut down it silently discards the event: the
// HTTP server stops serving before shutdown, so no live request reaches here.
func (w *batchWriter) enqueue(m Metric) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.closed {
		return
	}
	w.buf <- m
}

func (w *batchWriter) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	batch := make([]Metric, 0, w.batchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		w.writeBatch(batch)
		batch = batch[:0]
	}

	for {
		select {
		case m, ok := <-w.buf:
			if !ok {
				// Shutdown closed the buffer; the range above has already
				// drained every accepted event, so a final flush guarantees
				// no loss on a normal shutdown.
				flush()
				return
			}
			batch = append(batch, m)
			if len(batch) >= w.batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

// shutdown stops accepting events, waits for the background goroutine to drain
// and persist everything already buffered, and honours ctx as a deadline.
func (w *batchWriter) shutdown(ctx context.Context) error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	close(w.buf)
	w.mu.Unlock()

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *batchWriter) writeBatch(batch []Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), w.timeout)
	defer cancel()

	if err := w.execInsert(ctx, batch); err == nil {
		return
	}

	// A single offending row (e.g. a unique-constraint violation) fails the
	// whole multi-row statement, so retry row by row to isolate the bad one
	// and still persist every valid event.
	for i := range batch {
		if err := w.execInsert(ctx, batch[i:i+1]); err != nil {
			logger.Log.Error("metrics: failed to persist metric",
				"error", err,
				"resource", batch[i].Resource,
				"key", batch[i].Key,
			)
		}
	}
}

func (w *batchWriter) execInsert(ctx context.Context, batch []Metric) error {
	qb := query.New(w.db.Dialect()).
		Insert(Metric{}.TableName()).
		Columns("id", "resource", "resource_id", "name", "value")

	for _, m := range batch {
		qb = qb.Values(m.Id, m.Resource, m.ResourceId, m.Key, m.Value)
	}

	sqlStr, args, err := qb.Build()
	if err != nil {
		return err
	}

	_, err = w.db.Exec(ctx, sqlStr, args...)
	return err
}
