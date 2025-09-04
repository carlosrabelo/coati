package app

import (
	"fmt"
	"os"
	"strings"
)

func printColoredDiff(title string, content []byte) {
	green := "\033[32m"
	reset := "\033[0m"

	fmt.Printf("\n--- %s ---\n", title)

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			fmt.Println(line)
			continue
		}
		fmt.Printf("%s%s%s\n", green, line, reset)
	}
}

// printFileDiff reads the current file at path and prints a unified-style diff
// against newContent. Lines removed are shown in red, added in green.
func printFileDiff(title, path string, newContent []byte) {
	red := "\033[31m"
	green := "\033[32m"
	reset := "\033[0m"

	fmt.Printf("\n--- %s ---\n", title)

	var oldLines []string
	if data, err := os.ReadFile(path); err == nil {
		oldLines = splitLines(string(data))
	}
	newLines := splitLines(string(newContent))

	diff := computeDiff(oldLines, newLines)

	unchanged := true
	for _, d := range diff {
		switch {
		case strings.HasPrefix(d, "- "):
			unchanged = false
			fmt.Printf("%s%s%s\n", red, d, reset)
		case strings.HasPrefix(d, "+ "):
			unchanged = false
			fmt.Printf("%s%s%s\n", green, d, reset)
		default:
			fmt.Println(d)
		}
	}
	if unchanged {
		fmt.Println("(no changes)")
	}
}

func splitLines(s string) []string {
	return strings.Split(strings.TrimRight(s, "\n"), "\n")
}

// computeDiff returns a unified-style diff using LCS.
func computeDiff(old, new []string) []string {
	m, n := len(old), len(new)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if old[i-1] == new[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] > dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	var result []string
	i, j := m, n
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && old[i-1] == new[j-1]:
			result = append([]string{"  " + old[i-1]}, result...)
			i--
			j--
		case j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]):
			result = append([]string{"+ " + new[j-1]}, result...)
			j--
		default:
			result = append([]string{"- " + old[i-1]}, result...)
			i--
		}
	}
	return result
}
