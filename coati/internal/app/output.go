package app

import (
	"fmt"
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
