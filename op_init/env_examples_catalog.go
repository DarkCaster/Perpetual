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
	return cloneEnvExampleCatalog(envExampleCatalog, version)
}

func cloneEnvExampleCatalog(source []EnvExampleFile, version string) []EnvExampleFile {
	result := make([]EnvExampleFile, len(source))
	for i, example := range source {
		result[i] = example
		if version != "" {
			result[i].Content = "# Example .env config, version: " + version + "\n\n" + example.Content
		}
	}
	return result
}
