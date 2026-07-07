package metrics

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/database"
	_ "github.com/nicolasbonnici/gorest/database/sqlite"
)

func newTestDB(t *testing.T) database.Database {
	t.Helper()

	// busy_timeout + WAL let the pooled writer/reader connections coexist on a
	// file-backed SQLite DB without spurious SQLITE_BUSY during the test.
	path := filepath.Join(t.TempDir(), "metrics_test.db")
	dsn := "file:" + path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := database.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	_, err = db.Exec(ctx, `CREATE TABLE metrics (
		id TEXT PRIMARY KEY,
		resource TEXT NOT NULL,
		resource_id TEXT NOT NULL,
		name TEXT NOT NULL,
		value INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		UNIQUE (resource, resource_id, name)
	)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	return db
}

func countMetrics(t *testing.T, db database.Database) int {
	t.Helper()
	var n int
	if err := db.QueryRow(context.Background(), "SELECT COUNT(*) FROM metrics").Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func sampleMetric() Metric {
	return Metric{
		Id:         uuid.New().String(),
		Resource:   "post",
		ResourceId: uuid.New().String(),
		Key:        "view_count",
		Value:      1,
	}
}

func TestBatchWriter_FlushesOnShutdownWithoutLoss(t *testing.T) {
	db := newTestDB(t)
	w := newBatchWriter(db, batchWriterOptions{
		bufferCapacity: 512,
		batchSize:      64,
		flushInterval:  time.Hour, // force the drain to happen on shutdown, not the ticker
	})

	const total = 300
	for i := 0; i < total; i++ {
		w.enqueue(sampleMetric())
	}

	if err := w.shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if got := countMetrics(t, db); got != total {
		t.Fatalf("persisted %d metrics, want %d", got, total)
	}
}

func TestBatchWriter_ConcurrentEnqueueRaceFree(t *testing.T) {
	db := newTestDB(t)
	w := newBatchWriter(db, batchWriterOptions{
		bufferCapacity: 256,
		batchSize:      32,
		flushInterval:  5 * time.Millisecond,
	})

	const producers = 16
	const perProducer = 50

	var wg sync.WaitGroup
	wg.Add(producers)
	for p := 0; p < producers; p++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perProducer; i++ {
				w.enqueue(sampleMetric())
			}
		}()
	}
	wg.Wait()

	if err := w.shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if got := countMetrics(t, db); got != producers*perProducer {
		t.Fatalf("persisted %d metrics, want %d", got, producers*perProducer)
	}
}

func TestBatchWriter_PeriodicFlush(t *testing.T) {
	db := newTestDB(t)
	w := newBatchWriter(db, batchWriterOptions{
		bufferCapacity: 64,
		batchSize:      1000, // never reached, so only the ticker can flush
		flushInterval:  10 * time.Millisecond,
	})
	t.Cleanup(func() { _ = w.shutdown(context.Background()) })

	w.enqueue(sampleMetric())

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if countMetrics(t, db) == 1 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("periodic flush did not persist the buffered metric")
}

func TestBatchWriter_BadRowDoesNotPoisonBatch(t *testing.T) {
	db := newTestDB(t)
	w := newBatchWriter(db, batchWriterOptions{flushInterval: time.Hour})

	good1 := sampleMetric()
	dup := sampleMetric()
	dupClash := dup
	dupClash.Id = uuid.New().String() // distinct PK, same unique (resource, resource_id, name)
	good2 := sampleMetric()

	for _, m := range []Metric{good1, dup, dupClash, good2} {
		w.enqueue(m)
	}

	if err := w.shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	// The clashing duplicate is dropped; the three valid rows still persist.
	if got := countMetrics(t, db); got != 3 {
		t.Fatalf("persisted %d metrics, want 3", got)
	}
}

func TestBatchWriter_ShutdownIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	w := newBatchWriter(db, batchWriterOptions{})

	if err := w.shutdown(context.Background()); err != nil {
		t.Fatalf("first shutdown: %v", err)
	}
	if err := w.shutdown(context.Background()); err != nil {
		t.Fatalf("second shutdown: %v", err)
	}

	// Enqueue after shutdown must not panic and must be a no-op.
	w.enqueue(sampleMetric())
	if got := countMetrics(t, db); got != 0 {
		t.Fatalf("expected no metrics after post-shutdown enqueue, got %d", got)
	}
}
