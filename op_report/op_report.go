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
	var help, verbose, trace bool

	flags := flag.NewFlagSet(OpName, flag.ExitOnError)
	flags.BoolVar(&help, "h", false, "Show usage")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if verbose {
		logger.SetLevel(logging.DebugLevel)
	}
	if trace {
		logger.SetLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'report' operation")
	logger.Traceln("Args:", args)

	if help {
		usage.PrintOperationUsage("", flags)
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
