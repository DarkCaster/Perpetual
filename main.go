// main.go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/op_annotate"
	"github.com/DarkCaster/Perpetual/op_doc"
	"github.com/DarkCaster/Perpetual/op_embed"
	"github.com/DarkCaster/Perpetual/op_explain"
	"github.com/DarkCaster/Perpetual/op_implement"
	"github.com/DarkCaster/Perpetual/op_init"
	"github.com/DarkCaster/Perpetual/op_misc"
	"github.com/DarkCaster/Perpetual/op_onboard"
	"github.com/DarkCaster/Perpetual/op_report"
	"github.com/DarkCaster/Perpetual/op_stash"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

func getOperations() map[string]string {
	return map[string]string{
		op_init.OpName:      op_init.OpDesc,
		op_onboard.OpName:   op_onboard.OpDesc,
		op_annotate.OpName:  op_annotate.OpDesc,
		op_embed.OpName:     op_embed.OpDesc,
		op_implement.OpName: op_implement.OpDesc,
		op_stash.OpName:     op_stash.OpDesc,
		op_report.OpName:    op_report.OpDesc,
		op_doc.OpName:       op_doc.OpDesc,
		op_explain.OpName:   op_explain.OpDesc,
		op_misc.OpName:      op_misc.OpDesc,
	}
}

var Version = "development"

func main() {
	operations := getOperations()

	if len(os.Args) < 2 {
		usage.PrintMainUsage(fmt.Sprintf("Operation is required\nVersion: %s", Version), operations)
		return
	}

	operation := os.Args[1]
	args := os.Args[2:]

	if _, ok := operations[operation]; !ok {
		usage.PrintMainUsage(fmt.Sprintf("Unknown operation: %s", operation), operations)
		return
	}

	stdErrLogger, err := logging.NewStdErrSimpleLogger(logging.InfoLevel)
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGPIPE)
	go func() {
		defer shutdown()
		<-c
		panic(logging.LoggerPanic{Message: "Interrupted\n"})
	}()

	defer shutdown()

	switch strings.ToLower(operation) {
	case op_init.OpName:
		op_init.Run(Version, args, stdErrLogger)
	case op_onboard.OpName:
		op_onboard.Run(Version, args, stdErrLogger)
	case op_annotate.OpName:
		op_annotate.Run(args, false, stdErrLogger)
	case op_embed.OpName:
		op_embed.Run(args, false, stdErrLogger)
	case op_implement.OpName:
		op_implement.Run(args, stdErrLogger)
	case op_stash.OpName:
		op_stash.Run(args, false, stdErrLogger)
	case op_report.OpName:
		op_report.Run(args, stdErrLogger)
	case op_explain.OpName:
		op_explain.Run(args, stdErrLogger)
	case op_doc.OpName:
		op_doc.Run(args, stdErrLogger)
	case op_misc.OpName:
		op_misc.Run(args, stdErrLogger)
	}
}

func shutdown() {
	utils.RunGlobalCleanup()
	if r := recover(); r != nil {
		if lp, ok := r.(logging.LoggerPanic); ok {
			fmt.Fprintln(os.Stderr, lp.Message)
			os.Exit(-1)
		}
		fmt.Fprintf(os.Stderr, "panic: %v\n\n", r)
		debug.PrintStack()
		os.Exit(1)
	}
}
