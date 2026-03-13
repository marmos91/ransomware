package cli

import (
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
	startTime  time.Time
	operation  string
	mu         sync.Mutex
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
