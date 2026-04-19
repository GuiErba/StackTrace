package services

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"stacktrace/internal/models"
	"stacktrace/internal/repository"
	"stacktrace/pkg/notify"
)

var alertChannel chan models.LogEntry

const alertChannelSize = 5000

type alertWorker struct {
	db       *sql.DB
	notifier *notify.EmailNotifier
	mu       sync.Mutex
	windows  map[uuid.UUID][]time.Time
}

func InitAlertWorker(db *sql.DB, notifier *notify.EmailNotifier) {
	alertChannel = make(chan models.LogEntry, alertChannelSize)

	worker := &alertWorker{
		db:       db,
		notifier: notifier,
		windows:  make(map[uuid.UUID][]time.Time),
	}

	go worker.run()
	log.Println("Alert worker started")
}

func NotifyError(entry models.LogEntry) {
	if alertChannel == nil {
		return
	}

	select {
	case alertChannel <- entry:
	default:
		log.Println("WARNING: alert channel full, skipping error notification")
	}
}

func (w *alertWorker) run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case entry := <-alertChannel:
			w.recordError(entry)

		case <-ticker.C:
			w.evaluate()
		}
	}
}

func (w *alertWorker) recordError(entry models.LogEntry) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.windows[entry.ProjectID] = append(w.windows[entry.ProjectID], entry.Timestamp)
}

func (w *alertWorker) evaluate() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for projectID, timestamps := range w.windows {
        if len(timestamps) == 0 {
            continue
        }
		rules, err := repository.GetAlertRulesByProjectID(w.db, projectID)
		if err != nil {
			log.Printf("alert: failed to fetch rules for project %s: %v", projectID, err)
			continue
		}
        if len(rules) == 0 {
            continue
        }

		maxWindow := 0
		for _, rule := range rules {
			if rule.Condition == "error_count" && rule.WindowSeconds > maxWindow {
				maxWindow = rule.WindowSeconds
			}
		}

		widestCutoff := time.Now().UTC().Add(-time.Duration(maxWindow) * time.Second)
		filtered := make([]time.Time, 0, len(timestamps))
		for _, ts := range timestamps {
			if ts.After(widestCutoff) {
				filtered = append(filtered, ts)
			}
		}
		w.windows[projectID] = filtered

		triggered := false
		for _, rule := range rules {
			if rule.Condition != "error_count" {
				continue
			}

			cutoff := time.Now().UTC().Add(-time.Duration(rule.WindowSeconds) * time.Second)
			count := 0
			for _, ts := range filtered {
				if ts.After(cutoff) {
					count++
				}
			}

			if count >= rule.Threshold {
				w.triggerAlert(projectID, rule, count)
				triggered = true
			}
		}

		if triggered || len(filtered) == 0 {
			delete(w.windows, projectID)
		}
	}
}

func (w *alertWorker) triggerAlert(projectID uuid.UUID, rule models.AlertRule, errorCount int) {
	existing, err := repository.GetOpenIncident(w.db, projectID)
	if err != nil {
		log.Printf("Failed to check open incident for project %s: %v", projectID, err)
		return
	}
	if existing != nil {
		return
	}

	title := fmt.Sprintf("High error rate detected (%d errors)", errorCount)
	description := fmt.Sprintf(
		"%d errors in the last %d seconds exceeded threshold of %d",
		errorCount, rule.WindowSeconds, rule.Threshold,
	)

	incident := &models.Incident{
		ProjectID:   projectID,
		Title:       title,
		Description: &description,
	}

	err = repository.CreateIncident(w.db, incident)
	if err != nil {
		log.Printf("Failed to create incident for project %s: %v", projectID, err)
		return
	}

	log.Printf("Incident created for project %s: %s", projectID, title)

	if w.notifier != nil && rule.Channel == "email" {
		go w.notifier.SendAlert(
			rule.Destination,
			projectID.String(),
			title,
			description,
			errorCount,
		)
	}
}
