package op_onboard

import (
	"flag"
	"os"

	"github.com/DarkCaster/Perpetual/logging"
	"github.com/DarkCaster/Perpetual/usage"
)

const OpName = "onboard"
const OpDesc = "Validate environment configuration and create default global LLM configuration if needed"

func initFlags() *flag.FlagSet {
	return flag.NewFlagSet(OpName, flag.ExitOnError)
}

func Run(version string, args []string, logger logging.ILogger) {
	var help, check, verbose, trace bool
	var provider, keyMethod, key string

	onboardFlags := initFlags()
	onboardFlags.BoolVar(&help, "h", false, "Show usage")
	onboardFlags.BoolVar(&check, "c", false, "Check current global environment configuration")
	onboardFlags.StringVar(&provider, "e", "", "Recreate global env config for selected provider (anthropic|openai|ollama|generic)")
	onboardFlags.StringVar(&keyMethod, "m", "", "Auth method to write with -e, if applicable (Bearer|Basic)")
	onboardFlags.StringVar(&key, "k", "", "API key or login:password auth value to write with -e")
	onboardFlags.BoolVar(&verbose, "v", false, "Enable debug logging")
	onboardFlags.BoolVar(&trace, "vv", false, "Enable debug and trace logging")
	onboardFlags.Parse(args)

	if verbose {
		logger.EnableLevel(logging.DebugLevel)
	}
	if trace {
		logger.EnableLevel(logging.DebugLevel)
		logger.EnableLevel(logging.TraceLevel)
	}

	logger.Debugln("Starting 'onboard' operation")
	logger.Traceln("Args:", args)

	if help || (!check && provider == "") {
		usage.PrintOperationUsage("", onboardFlags)
	}

	if check && provider != "" {
		usage.PrintOperationUsage("-c and -e cannot be used together", onboardFlags)
	}

	if keyMethod != "" && provider == "" {
		usage.PrintOperationUsage("-m can only be used together with -e", onboardFlags)
	}

	if key != "" && provider == "" {
		usage.PrintOperationUsage("-k can only be used together with -e", onboardFlags)
	}

	if provider != "" {
		if err := RolloutEnvironmentConfig(version, provider, keyMethod, key, logger); err != nil {
			logger.Panicln("Failed to create global env configuration:", err)
		}
	}

	env, err := DetectActiveEnvironment(logger)
	if err != nil {
		logger.Panicln("Failed to detect active environment:", err)
	}

	PrintActiveEnvironment(os.Stdout, env)
	if !ValidateEnvironment(os.Stdout, env) {
		logger.Panicln("Environment configuration validation failed")
	}
}
