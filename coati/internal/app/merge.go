package app

import (
	"os"
	"strings"
)

const (
	markerBeginOriginal = "# BEGIN ORIGINAL"
	markerEndOriginal   = "# END ORIGINAL"
	markerBeginCoati    = "# BEGIN COATI"
	markerEndCoati      = "# END COATI"
)

// mergeWithMarkers reads the existing file at path (if any), preserves its
// ORIGINAL section (or wraps the whole content on first run), and replaces
// the COATI section with newContent. Returns the merged bytes ready to write.
func mergeWithMarkers(path string, newContent []byte) ([]byte, error) {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	originalSection := extractOrWrapOriginal(string(existing))
	coatiSection := markerBeginCoati + "\n" + strings.TrimRight(string(newContent), "\n") + "\n" + markerEndCoati

	var parts []string
	if originalSection != "" {
		parts = append(parts, originalSection)
	}
	parts = append(parts, coatiSection)

	return []byte(strings.Join(parts, "\n\n") + "\n"), nil
}

// extractOrWrapOriginal returns the ORIGINAL section from existing content.
// On first run (no markers), wraps the current content as ORIGINAL.
func extractOrWrapOriginal(existing string) string {
	beginIdx := strings.Index(existing, markerBeginOriginal)
	endIdx := strings.Index(existing, markerEndOriginal)

	if beginIdx >= 0 && endIdx > beginIdx {
		return strings.TrimSpace(existing[beginIdx:endIdx+len(markerEndOriginal)])
	}

	trimmed := strings.TrimSpace(existing)
	if trimmed == "" {
		return ""
	}
	return markerBeginOriginal + "\n" + trimmed + "\n" + markerEndOriginal
}
