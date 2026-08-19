package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dracory/logstore"
)

// seedDB populates the store with sample log entries so the admin panel
// has enough rows to exercise pagination, filtering, and sorting.
//
// Creates 120 log entries with mixed levels and messages.
func seedDB(store logstore.StoreInterface, logger *slog.Logger) {
	ctx := context.Background()

	levels := []string{"trace", "debug", "info", "warning", "error"}

	for i := 1; i <= 120; i++ {
		level := levels[(i-1)%len(levels)]
		l := logstore.NewLog()
		l.SetLevel(level)
		l.SetMessage(fmt.Sprintf("Sample log entry %03d — operation completed with status code %d", i, 200+(i%5)))
		l.SetContext(fmt.Sprintf(`{"request_id":"req-%03d","user_id":"user-%d","duration_ms":%d}`, i, i%20, 10+i*3))
		if err := store.LogCreate(ctx, l); err != nil {
			logger.Error("seedDB: failed to create log", "index", i, "error", err)
		}
	}

	logger.Info("seedDB complete", "logs", 120)
}
