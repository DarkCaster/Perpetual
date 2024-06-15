package op_report

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const (
	OpName = "report"
	OpDesc = "Create report from project source code, that can be manually copypasted into the LLM user-interface for further manual analisys"
)

func Run(args []string, logger logging.ILogger) {
	//###IMPLEMENT###
	//add help, verbose, trace flags
	//you must use op_init/op_init.go as reference of how to organize flag-handling code

	// Define and parse command-line flags
	flags := flag.NewFlagSet(OpName, flag.ExitOnError)
	flags.Usage = func() {
		usage.PrintOperationUsage("", flags)
	}

	// Parse flags
	err := flags.Parse(args)
	if err != nil {
		logger.Errorln(err)
		return
	}

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		logger.Errorln(err)
		return
	}

	// Get the project root directory
	projectRootDir := filepath.Dir(cwd)

	// Get the list of project files
	fileChecksums, projectFiles, allFiles, err := utils.GetProjectFileList(projectRootDir, "", nil, nil)
	if err != nil {
		logger.Errorln(err)
		return
	}

	// Create the report
	report := createReport(projectRootDir, projectFiles, allFiles, fileChecksums)

	// Print the report
	fmt.Println(report)
}

func createReport(projectRootDir string, projectFiles, allFiles []string, fileChecksums map[string]string) string {
	report := "Project Root Directory: " + projectRootDir + "\n\n"
	report += "Project Files:\n"
	for _, file := range projectFiles {
		report += file + "\n"
	}
	report += "\nAll Files:\n"
	for _, file := range allFiles {
		report += file + "\n"
	}
	report += "\nFile Checksums:\n"
	for file, checksum := range fileChecksums {
		report += file + ": " + checksum + "\n"
	}
	return report
}
