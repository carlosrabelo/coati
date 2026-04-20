package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeWithMarkers_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts")

	result, err := mergeWithMarkers(path, []byte("192.168.1.1\tweb\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := string(result)
	if strings.Contains(s, markerBeginOriginal) {
		t.Error("should not have ORIGINAL section for empty file")
	}
	if !strings.Contains(s, markerBeginCoati) {
		t.Error("missing BEGIN COATI marker")
	}
	if !strings.Contains(s, "192.168.1.1\tweb") {
		t.Error("missing new content in COATI section")
	}
}

func TestMergeWithMarkers_FirstRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts")
	existing := "127.0.0.1\tlocalhost\n"
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := mergeWithMarkers(path, []byte("10.0.0.1\tdb\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := string(result)
	if !strings.Contains(s, markerBeginOriginal) {
		t.Error("missing BEGIN ORIGINAL marker")
	}
	if !strings.Contains(s, "127.0.0.1\tlocalhost") {
		t.Error("original content not preserved")
	}
	if !strings.Contains(s, markerBeginCoati) {
		t.Error("missing BEGIN COATI marker")
	}
	if !strings.Contains(s, "10.0.0.1\tdb") {
		t.Error("missing new content in COATI section")
	}
	originalPos := strings.Index(s, markerBeginOriginal)
	coatiPos := strings.Index(s, markerBeginCoati)
	if originalPos > coatiPos {
		t.Error("ORIGINAL section should appear before COATI section")
	}
}

func TestMergeWithMarkers_SubsequentRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hosts")
	existing := markerBeginOriginal + "\n127.0.0.1\tlocalhost\n" + markerEndOriginal +
		"\n\n" + markerBeginCoati + "\n10.0.0.1\told-entry\n" + markerEndCoati + "\n"
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := mergeWithMarkers(path, []byte("10.0.0.2\tnew-entry\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := string(result)
	if !strings.Contains(s, "127.0.0.1\tlocalhost") {
		t.Error("original content should be preserved")
	}
	if strings.Contains(s, "10.0.0.1\told-entry") {
		t.Error("old COATI content should be replaced")
	}
	if !strings.Contains(s, "10.0.0.2\tnew-entry") {
		t.Error("new content missing from COATI section")
	}
	count := strings.Count(s, markerBeginOriginal)
	if count != 1 {
		t.Errorf("expected 1 BEGIN ORIGINAL marker, got %d", count)
	}
}

func TestExtractOrWrapOriginal_Empty(t *testing.T) {
	result := extractOrWrapOriginal("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestExtractOrWrapOriginal_NoMarkers(t *testing.T) {
	result := extractOrWrapOriginal("127.0.0.1\tlocalhost")
	if !strings.HasPrefix(result, markerBeginOriginal) {
		t.Error("should wrap with BEGIN ORIGINAL")
	}
	if !strings.HasSuffix(result, markerEndOriginal) {
		t.Error("should wrap with END ORIGINAL")
	}
	if !strings.Contains(result, "127.0.0.1\tlocalhost") {
		t.Error("original content missing")
	}
}

func TestExtractOrWrapOriginal_WithMarkers(t *testing.T) {
	input := markerBeginOriginal + "\n127.0.0.1\tlocalhost\n" + markerEndOriginal +
		"\n\n" + markerBeginCoati + "\n10.0.0.1\tdb\n" + markerEndCoati
	result := extractOrWrapOriginal(input)
	if !strings.HasPrefix(result, markerBeginOriginal) {
		t.Error("should start with BEGIN ORIGINAL")
	}
	if !strings.HasSuffix(result, markerEndOriginal) {
		t.Error("should end with END ORIGINAL")
	}
	if strings.Contains(result, markerBeginCoati) {
		t.Error("COATI section should not appear in ORIGINAL extraction")
	}
}
