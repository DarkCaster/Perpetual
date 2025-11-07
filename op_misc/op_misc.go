package op_misc

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/DarkCaster/Perpetual/config"
	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/usage"
	"github.com/DarkCaster/Perpetual/utils"
)

const OpName = "misc"
const OpDesc = "Helper functions not covered by other operations. All human-readable logging goes to stderr and machine parsable output to stdout."

func miscFlags() *flag.FlagSet {
	flags := flag.NewFlagSet(OpName, flag.ExitOnError)

	return flags
}

func Run(args []string, stdErrLogger logging.ILogger) {
	var help, projTest, listFiles, verbose, trace, includeTests bool
	var descFile, userFilterFile string

	// Parse flags for the "misc" operation
	flags := miscFlags()
	flags.BoolVar(&help, "h", false, "Show usage")
	// Main flags to perform particular function
	flags.BoolVar(&projTest, "p", false, "Search for .perpetual dir, starting from curdir and check json configs inside it. Output full path of .perpetual dir on success.")
	flags.BoolVar(&listFiles, "l", false, "List project files accesible by utility, can work with '-x' and '-u' flags.")
	// Extra options, may be used with flags above to alter its behavior or test some more things
	flags.StringVar(&descFile, "df", "", "Optional path to project description file (valid values: file-path|disabled)")
	flags.BoolVar(&includeTests, "u", false, "Do not exclude unit-tests source files from processing")
	flags.StringVar(&userFilterFile, "x", "", "Path to user-supplied regex filter-file for filtering out certain files from processing")
	flags.BoolVar(&verbose, "v", false, "Enable debug logging")
	flags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	flags.Parse(args)

	if verbose {
		stdErrLogger.EnableLevel(logging.DebugLevel)
	}
	if trace {
		stdErrLogger.EnableLevel(logging.DebugLevel)
		stdErrLogger.EnableLevel(logging.TraceLevel)
	}

	stdErrLogger.Debugln("Starting 'misc' operation")
	stdErrLogger.Traceln("Args:", args)

	fc := 0
	if projTest {
		fc++
	}
	if listFiles {
		fc++
	}

	if help {
		usage.PrintOperationUsage("", flags)
	} else if fc > 1 {
		usage.PrintOperationUsage("Only one of the following flags must be provided: -p, -l", flags)
	} else if fc < 1 {
		usage.PrintOperationUsage("One of the following flags must be provided: -p, -l", flags)
	}

	// Initialize: detect work directories, load .env file with LLM settings, load file filtering regexps
	projectRootDir, perpetualDir, err := utils.FindProjectRoot(stdErrLogger)
	if err != nil {
		stdErrLogger.Panicln("Error finding project root directory:", err)
	}

	stdErrLogger.Infoln("Project root directory:", projectRootDir)
	stdErrLogger.Debugln("Perpetual directory:", perpetualDir)

	//load json config files for project and operations, will panic if it cannot be loaded or parsed
	projectConfig := config.LoadProjectConfig(perpetualDir, stdErrLogger)
	config.LoadOpAnnotateConfig(perpetualDir, stdErrLogger)
	config.LoadOpDocConfig(perpetualDir, stdErrLogger)
	config.LoadOpExplainConfig(perpetualDir, stdErrLogger)
	config.LoadOpImplementConfig(perpetualDir, stdErrLogger)
	config.LoadOpReportConfig(perpetualDir, stdErrLogger)

	//test load of project description file
	wrn := ""
	if descFile == "" {
		_, wrn, err = utils.LoadTextFile(filepath.Join(perpetualDir, config.ProjectDescriptionFile))
		if err != nil {
			if os.IsNotExist(err) {
				stdErrLogger.Infoln("Not loading missing project description file (description.md)")
			} else {
				stdErrLogger.Panicln("Failed to load project description file:", err)
			}
		}
		if wrn != "" {
			stdErrLogger.Warnf("%s: %s", config.ProjectDescriptionFile, wrn)
		}
	} else if strings.ToLower(descFile) != "disabled" {
		_, wrn, err = utils.LoadTextFile(descFile)
		if err != nil {
			stdErrLogger.Panicln("Failed to load project description file:", err)
		}
		if wrn != "" {
			stdErrLogger.Warnf("%s: %s", descFile, wrn)
		}
	} else {
		stdErrLogger.Infoln("Loading of project description file (description.md) is disabled")
	}

	//if we are only checking for .perpetual directory validity, output it here
	if projTest {
		fmt.Println(perpetualDir)
		return
	}

	// Preparation of project files
	stdErrLogger.Infoln("Fetching project files")
	fileNames, _, err := utils.GetProjectFileList(
		projectRootDir,
		perpetualDir,
		projectConfig.RegexpArray(config.K_ProjectFilesWhitelist),
		projectConfig.RegexpArray(config.K_ProjectFilesBlacklist))

	if err != nil {
		stdErrLogger.Panicln("Error getting project file-list:", err)
	}

	// Check fileNames array for case collisions
	if !utils.CheckFilenameCaseCollisions(fileNames) {
		//list current files
		if listFiles {
			for _, file := range fileNames {
				fmt.Println(file)
			}
		}
		stdErrLogger.Panicln("Filename case collisions detected in project files")
	}

	// File names and dir-names must not contain path separators characters
	if !utils.CheckForPathSeparatorsInFilenames(fileNames) {
		//list current files
		if listFiles {
			for _, file := range fileNames {
				fmt.Println(file)
			}
		}
		stdErrLogger.Panicln("Invalid characters detected in project filenames or directories: / and \\ characters are not allowed!")
	}

	// Filter project files with unittest- and user- filters
	var userBlacklist []*regexp.Regexp
	if userFilterFile != "" {
		userBlacklist, err = utils.AppendUserFilterFromFile(userFilterFile, userBlacklist)
		if err != nil {
			stdErrLogger.Panicln("Error processing user blacklist-filter:", err)
		}
	}

	if !includeTests {
		userBlacklist = append(userBlacklist, projectConfig.RegexpArray(config.K_ProjectTestFilesBlacklist)...)
	}

	fileNames, droppedFiles := utils.FilterFilesWithBlacklist(fileNames, userBlacklist)
	if len(droppedFiles) > 0 {
		stdErrLogger.Infoln("Number of blacklisted files with unit-tests and/or user-provided filters:", len(droppedFiles))
		slices.Sort(droppedFiles)
		for _, file := range droppedFiles {
			stdErrLogger.Debugln("Filtered-out:", file)
		}
	}

	//list currently available files
	slices.Sort(fileNames)
	if listFiles {
		for _, file := range fileNames {
			fmt.Println(file)
		}
		return
	}
}
