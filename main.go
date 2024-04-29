// main.go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/DarkCaster/Perpetual/op_annotate"
	"github.com/DarkCaster/Perpetual/op_implement"
	"github.com/DarkCaster/Perpetual/op_init"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func getOperations() map[string]string {
	return map[string]string{
		op_init.OpName:      op_init.OpDesc,
		op_annotate.OpName:  op_annotate.OpDesc,
		op_implement.OpName: op_implement.OpDesc,
	}
}

func main() {
	operations := getOperations()

	if len(os.Args) < 2 {
		usage.PrintMainUsage("Operation is required", operations)
		return
	}

	operation := os.Args[1]
	args := os.Args[2:]

	if _, ok := operations[operation]; !ok {
		usage.PrintMainUsage(fmt.Sprintf("Unknown operation: %s", operation), operations)
		return
	}

	switch strings.ToLower(operation) {
	case op_init.OpName:
		op_init.Run(args, logger)
	case op_annotate.OpName:
		op_annotate.Run(args, logger)
	case op_implement.OpName:
		op_implement.Run(args, logger)
	}
}
