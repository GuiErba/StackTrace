package services

import (
	"database/sql"
	"log"
	"time"

	"stacktrace/internal/models"
	"stacktrace/internal/repository"
)

const (
	channelBufferSize = 10000
	batchSize         = 100
	flushInterval     = 1 * time.Second
	workerCount       = 5
)

var logChannel chan models.LogEntry

func InitIngest(db *sql.DB) {
	logChannel = make(chan models.LogEntry, channelBufferSize)

	for i := 0; i < workerCount; i++ {
		go persistWorker(db, i)
	}

	log.Printf("Ingest engine started: %d workers, buffer size %d", workerCount, channelBufferSize)
}

func Enqueue(entry models.LogEntry) bool {
	select {
	case logChannel <- entry:
		if entry.Level == "error" {
			NotifyError(entry)
		}
		return true
	default:
		log.Println("WARNING: log channel full, dropping log")
		return false
	}
}

func persistWorker(db *sql.DB, workerID int) {
	batch := make([]models.LogEntry, 0, batchSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case entry := <-logChannel:
			batch = append(batch, entry)

			if len(batch) >= batchSize {
				flushBatch(db, &batch, workerID)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				flushBatch(db, &batch, workerID)
			}
		}
	}
}

func flushBatch(db *sql.DB, batch *[]models.LogEntry, workerID int) {
	if len(*batch) == 0 {
		return
	}

	err := repository.InsertLogBatch(db, *batch)
	if err != nil {
		log.Printf("Worker %d: failed to persist batch of %d logs: %v", workerID, len(*batch), err)
		return
	}

	log.Printf("Worker %d: persisted %d logs", workerID, len(*batch))
	*batch = (*batch)[:0]
}
