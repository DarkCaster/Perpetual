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

func isValidProvider(provider string) bool {
	switch provider {
	case "anthropic", "openai", "ollama", "generic":
		return true
	default:
		return false
	}
}

func Run(version string, args []string, logger logging.ILogger) {
	var help, verbose, trace bool
	var mode, provider, keyMethod, key string

	onboardFlags := initFlags()
	onboardFlags.BoolVar(&help, "h", false, "Show usage")
	onboardFlags.StringVar(&mode, "m", "", "Select operation mode: check, install")
	onboardFlags.StringVar(&provider, "p", "", "Provider to install with 'install' mode (anthropic|openai|ollama|generic). Mandatory for 'install' mode if LLM_PROVIDER env variable is not set to a valid value")
	onboardFlags.StringVar(&keyMethod, "km", "", "Auth method to write in 'install' mode, if applicable (Bearer|Basic)")
	onboardFlags.StringVar(&key, "k", "", "API key or login:password auth value to write in 'install' mode")
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

	var check, install bool
	switch mode {
	case "check":
		check = true
	case "install":
		install = true
	case "":
		usage.PrintOperationUsage("You must provide a valid operation mode with the '-m' flag (valid values: check|install)", onboardFlags)
	default:
		logger.Errorln("Invalid operation mode:", mode)
		usage.PrintOperationUsage("You must provide a valid operation mode with the '-m' flag (valid values: check|install)", onboardFlags)
	}

	if help {
		usage.PrintOperationUsage("", onboardFlags)
	}

	if check {
		if provider != "" {
			usage.PrintOperationUsage("-p can only be used with 'install' mode", onboardFlags)
		}
		if keyMethod != "" {
			usage.PrintOperationUsage("-km can only be used with 'install' mode", onboardFlags)
		}
		if key != "" {
			usage.PrintOperationUsage("-k can only be used with 'install' mode", onboardFlags)
		}
	}

	if install {
		resolvedProvider := provider
		if resolvedProvider == "" {
			resolvedProvider = os.Getenv("LLM_PROVIDER")
		}
		if !isValidProvider(resolvedProvider) {
			usage.PrintOperationUsage("'install' mode requires a valid provider set with -p or LLM_PROVIDER env variable (anthropic|openai|ollama|generic)", onboardFlags)
		}
		if err := RolloutEnvironmentConfig(version, resolvedProvider, keyMethod, key, logger); err != nil {
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
