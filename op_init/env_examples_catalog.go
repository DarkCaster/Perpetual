package op_init

type EnvExampleFile struct {
	Filename string
	Content  string
	Provider string
}

var envExampleCatalog = []EnvExampleFile{
	{
		Filename: dotEnvExampleFileName,
		Content:  dotEnvExample,
	},
	{
		Filename: ollamaEnvExampleFileName,
		Content:  ollamaEnvExample,
		Provider: "ollama",
	},
	{
		Filename: openAiEnvExampleFileName,
		Content:  openAiEnvExample,
		Provider: "openai",
	},
	{
		Filename: anthropicEnvExampleFileName,
		Content:  anthropicEnvExample,
		Provider: "anthropic",
	},
	{
		Filename: genericEnvExampleFileName,
		Content:  genericEnvExample,
		Provider: "generic",
	},
}

func GetEnvExampleCatalogWithVersion(version string) []EnvExampleFile {
	topString := ""
	if version != "" {
		topString = "# Example .env config, version: " + version
	}
	return cloneEnvExampleCatalog(envExampleCatalog, topString)
}

func GetEnvCatalogWithTopString(topString string) []EnvExampleFile {
	return cloneEnvExampleCatalog(envExampleCatalog, topString)
}

func cloneEnvExampleCatalog(source []EnvExampleFile, topString string) []EnvExampleFile {
	result := make([]EnvExampleFile, len(source))
	for i, example := range source {
		result[i] = example
		if topString != "" {
			result[i].Content = topString + "\n\n" + example.Content
		}
	}
	return result
}
