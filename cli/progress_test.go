package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1536, "1.50 KB"},
		{2411724, "2.30 MB"},
		{1181116006, "1.10 GB"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := formatBytes(tc.bytes)
			if got != tc.want {
				t.Fatalf("formatBytes(%d) = %q, want %q", tc.bytes, got, tc.want)
			}
		})
	}
}

func TestProgress_TickAndReport(t *testing.T) {
	p := NewProgress("Testing", 5)

	for i := 0; i < 5; i++ {
		p.Tick("/tmp/file", 100)
	}
	p.AddError()

	report := p.GenerateReport(false)

	if report.Operation != "Testing" {
		t.Fatalf("operation = %q, want %q", report.Operation, "Testing")
	}
	if report.Files != 5 {
		t.Fatalf("files = %d, want 5", report.Files)
	}
	if report.TotalFiles != 5 {
		t.Fatalf("totalFiles = %d, want 5", report.TotalFiles)
	}
	if report.Bytes != 500 {
		t.Fatalf("bytes = %d, want 500", report.Bytes)
	}
	if report.BytesHuman != "500 B" {
		t.Fatalf("bytesHuman = %q, want %q", report.BytesHuman, "500 B")
	}
	if report.Errors != 1 {
		t.Fatalf("errors = %d, want 1", report.Errors)
	}
	if report.Status != "completed with errors" {
		t.Fatalf("status = %q, want %q", report.Status, "completed with errors")
	}
}

func TestProgress_SuccessReport(t *testing.T) {
	p := NewProgress("Encrypting", 3)

	for i := 0; i < 3; i++ {
		p.Tick("/tmp/file", 1024)
	}

	report := p.GenerateReport(false)

	if report.Status != "success" {
		t.Fatalf("status = %q, want %q", report.Status, "success")
	}
}

func TestWriteReport(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.json")

	report := Report{
		Operation:  "Testing",
		Files:      10,
		TotalFiles: 10,
		Bytes:      1024,
		BytesHuman: "1.00 KB",
		Duration:   "100ms",
		DurationMs: 100,
		Errors:     0,
		Status:     "success",
	}

	if err := WriteReport(path, report); err != nil {
		t.Fatalf("write report: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}

	var got Report
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal report: %v", err)
	}

	if got != report {
		t.Fatalf("report mismatch:\n got %+v\nwant %+v", got, report)
	}
}
