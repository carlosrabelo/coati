package app

import (
	"strings"
	"testing"
)

func TestSplitLines(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{"a\nb\nc\n", []string{"a", "b", "c"}},
		{"a\nb\nc", []string{"a", "b", "c"}},
		{"", []string{""}},
		{"single", []string{"single"}},
	}
	for _, c := range cases {
		got := splitLines(c.input)
		if len(got) != len(c.expected) {
			t.Errorf("splitLines(%q): len=%d want %d", c.input, len(got), len(c.expected))
			continue
		}
		for i, line := range got {
			if line != c.expected[i] {
				t.Errorf("splitLines(%q)[%d] = %q, want %q", c.input, i, line, c.expected[i])
			}
		}
	}
}

func TestComputeDiff_Identical(t *testing.T) {
	lines := []string{"aaa", "bbb", "ccc"}
	diff := computeDiff(lines, lines)
	for _, d := range diff {
		if !strings.HasPrefix(d, "  ") {
			t.Errorf("identical files: expected context line, got %q", d)
		}
	}
}

func TestComputeDiff_AllAdded(t *testing.T) {
	diff := computeDiff([]string{}, []string{"aaa", "bbb"})
	for _, d := range diff {
		if !strings.HasPrefix(d, "+ ") {
			t.Errorf("all added: expected '+ ' prefix, got %q", d)
		}
	}
}

func TestComputeDiff_AllRemoved(t *testing.T) {
	diff := computeDiff([]string{"aaa", "bbb"}, []string{})
	for _, d := range diff {
		if !strings.HasPrefix(d, "- ") {
			t.Errorf("all removed: expected '- ' prefix, got %q", d)
		}
	}
}

func TestComputeDiff_Mixed(t *testing.T) {
	old := []string{"aaa", "bbb", "ccc"}
	new := []string{"aaa", "xxx", "ccc"}
	diff := computeDiff(old, new)

	hasRemoved := false
	hasAdded := false
	hasContext := false
	for _, d := range diff {
		switch {
		case strings.HasPrefix(d, "- "):
			hasRemoved = true
		case strings.HasPrefix(d, "+ "):
			hasAdded = true
		case strings.HasPrefix(d, "  "):
			hasContext = true
		}
	}
	if !hasRemoved {
		t.Error("expected at least one removed line")
	}
	if !hasAdded {
		t.Error("expected at least one added line")
	}
	if !hasContext {
		t.Error("expected at least one context line")
	}
}
