package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type Progress struct {
	total      int
	counter    int64
	totalBytes int64
	errors     int64
	startTime  time.Time
	operation  string
	mu         sync.Mutex
}

// Report holds the summary data for JSON/text report output.
type Report struct {
	Operation  string `json:"operation"`
	Files      int64  `json:"files"`
	TotalFiles int    `json:"totalFiles"`
	Bytes      int64  `json:"bytes"`
	BytesHuman string `json:"bytesHuman"`
	Duration   string `json:"duration"`
	DurationMs int64  `json:"durationMs"`
	Errors     int64  `json:"errors"`
	Status     string `json:"status"`
}

func NewProgress(operation string, total int) *Progress {
	return &Progress{
		total:     total,
		startTime: time.Now(),
		operation: operation,
	}
}

func (p *Progress) Tick(path string, size int64) {
	n := atomic.AddInt64(&p.counter, 1)
	atomic.AddInt64(&p.totalBytes, size)

	p.mu.Lock()
	fmt.Fprintf(os.Stderr, "[%d/%d] %s %s\n", n, p.total, p.operation, filepath.Base(path))
	p.mu.Unlock()
}

func (p *Progress) AddError() {
	atomic.AddInt64(&p.errors, 1)
}

func (p *Progress) Summary(failed bool) {
	elapsed := time.Since(p.startTime).Round(time.Millisecond)
	count := atomic.LoadInt64(&p.counter)
	totalBytes := atomic.LoadInt64(&p.totalBytes)

	status := "complete"
	if failed {
		status = "completed with errors"
	}

	fmt.Fprintf(os.Stderr, "\n%s %s: %d files (%s) in %s\n",
		p.operation, status, count, formatBytes(totalBytes), elapsed)
}

// GenerateReport returns a Report struct with the operation summary.
// The failed flag reflects the overall command outcome (not just tracked errors).
func (p *Progress) GenerateReport(failed bool) Report {
	elapsed := time.Since(p.startTime).Round(time.Millisecond)
	count := atomic.LoadInt64(&p.counter)
	totalBytes := atomic.LoadInt64(&p.totalBytes)
	errCount := atomic.LoadInt64(&p.errors)

	status := "success"
	if failed || errCount > 0 {
		status = "completed with errors"
	}

	return Report{
		Operation:  p.operation,
		Files:      count,
		TotalFiles: p.total,
		Bytes:      totalBytes,
		BytesHuman: formatBytes(totalBytes),
		Duration:   elapsed.String(),
		DurationMs: elapsed.Milliseconds(),
		Errors:     errCount,
		Status:     status,
	}
}

// WriteReport writes the report to the given path as JSON.
func WriteReport(path string, report Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func formatBytes(b int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.2f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.2f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
