package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"ajm188.scratchpad/vtadmin-demo-insert-data/stats"
)

type counter struct {
	*stats.Counter
	lastReportedMark int
}

func newCounter() *counter {
	return &counter{
		Counter:          &stats.Counter{},
		lastReportedMark: 0,
	}
}

var Stats = &struct {
	RowsInserted *counter
	Errors       *counter

	startTime      time.Time
	lastReportTime time.Time
}{
	RowsInserted: newCounter(),
	Errors:       newCounter(),
}

func ReportStats() {
	rows := Stats.RowsInserted.Get()
	errors := Stats.Errors.Get()
	now := time.Now()

	defer func() {
		Stats.lastReportTime = now
		Stats.RowsInserted.lastReportedMark = rows
		Stats.Errors.lastReportedMark = errors
	}()

	statsMap := map[string]struct {
		Total       int
		Current     int
		TotalRate   float64
		CurrentRate float64
	}{
		"Rows Inserted": {
			Total:       rows,
			TotalRate:   float64(rows) / now.Sub(Stats.startTime).Seconds(),
			Current:     rows - Stats.RowsInserted.lastReportedMark,
			CurrentRate: float64(rows-Stats.RowsInserted.lastReportedMark) / now.Sub(Stats.lastReportTime).Seconds(),
		},
		"Errors": {
			Total:       errors,
			TotalRate:   float64(errors) / now.Sub(Stats.startTime).Seconds(),
			Current:     errors - Stats.Errors.lastReportedMark,
			CurrentRate: float64(errors-Stats.Errors.lastReportedMark) / now.Sub(Stats.lastReportTime).Seconds(),
		},
	}

	b, err := json.MarshalIndent(statsMap, "", "    ")
	if err != nil {
		log.Printf("Error reporting stats: %s", err)
		return
	}

	fmt.Printf("%s\n", b)
}
