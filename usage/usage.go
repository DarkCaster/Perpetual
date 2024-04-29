package usage

import (
	"flag"
	"fmt"
	"os"
	"sort"
)

func PrintMainUsage(msg string, operations map[string]string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	}

	if operations != nil {
		fmt.Fprintf(os.Stderr, "Available operations:\n")

		// Sort operation names alphabetically
		var sortedOps []string
		for op := range operations {
			sortedOps = append(sortedOps, op)
		}
		sort.Strings(sortedOps)

		// Print sorted operations
		for _, op := range sortedOps {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", op, operations[op])
		}
		fmt.Fprintf(os.Stderr, "Usage: %s <operation> [args...] [-h - show help for selected operation] \n", os.Args[0])
	}

	os.Exit(1)
}

func PrintOperationUsage(msg string, flags *flag.FlagSet) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	}

	flags.Usage()
	os.Exit(1)
}
