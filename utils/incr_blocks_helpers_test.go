package utils

import (
	"regexp"
	"testing"
)

func TestParseIncrBlocks(t *testing.T) {
	// Test case 1: Basic functionality with valid blocks
	t.Run("BasicValidBlocks", func(t *testing.T) {
		source := `###SEARCH_START###search text 1###SEARCH_END###replace text 1###REPLACE_END###
###SEARCH_START###search text 2###SEARCH_END###replace text 2###REPLACE_END###`

		searchTags := []*regexp.Regexp{
			regexp.MustCompile(`###SEARCH_START###`),
			regexp.MustCompile(`###SEARCH_END###`),
			regexp.MustCompile(`###REPLACE_END###`),
		}

		blocks, err := ParseIncrBlocks(source, searchTags)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(blocks) != 2 {
			t.Fatalf("Expected 2 blocks, got %d", len(blocks))
		}

		expectedBlocks := []IncrBlock{
			{Search: "search text 1", Replace: "replace text 1"},
			{Search: "search text 2", Replace: "replace text 2"},
		}

		for i, block := range blocks {
			if block.Search != expectedBlocks[i].Search {
				t.Errorf("Block %d: expected search '%s', got '%s'", i, expectedBlocks[i].Search, block.Search)
			}
			if block.Replace != expectedBlocks[i].Replace {
				t.Errorf("Block %d: expected replace '%s', got '%s'", i, expectedBlocks[i].Replace, block.Replace)
			}
		}
	})

	// Test case 2: Single block
	t.Run("SingleBlock", func(t *testing.T) {
		source := `###START###find this###END###replace with this###DONE###`

		searchTags := []*regexp.Regexp{
			regexp.MustCompile(`###START###`),
			regexp.MustCompile(`###END###`),
			regexp.MustCompile(`###DONE###`),
		}

		blocks, err := ParseIncrBlocks(source, searchTags)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(blocks) != 1 {
			t.Fatalf("Expected 1 block, got %d", len(blocks))
		}

		expectedBlock := IncrBlock{
			Search:  "find this",
			Replace: "replace with this",
		}

		if blocks[0].Search != expectedBlock.Search {
			t.Errorf("Expected search '%s', got '%s'", expectedBlock.Search, blocks[0].Search)
		}
		if blocks[0].Replace != expectedBlock.Replace {
			t.Errorf("Expected replace '%s', got '%s'", expectedBlock.Replace, blocks[0].Replace)
		}
	})

	// Test case 3: No blocks found
	t.Run("NoBlocks", func(t *testing.T) {
		source := "just regular text without any tags"
		searchTags := []*regexp.Regexp{
			regexp.MustCompile(`###START###`),
			regexp.MustCompile(`###END###`),
			regexp.MustCompile(`###DONE###`),
		}

		_, err := ParseIncrBlocks(source, searchTags)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})

	// Test case 4: Incomplete block - missing search end tag
	t.Run("MissingSearchEndTag", func(t *testing.T) {
		source := `###START###search text###DONE###`
		searchTags := []*regexp.Regexp{
			regexp.MustCompile(`###START###`),
			regexp.MustCompile(`###END###`),
			regexp.MustCompile(`###DONE###`),
		}

		_, err := ParseIncrBlocks(source, searchTags)
		if err == nil {
			t.Error("Expected error for incomplete block, got nil")
		}
	})

	// Test case 5: Incomplete block - missing replace end tag
	t.Run("MissingReplaceEndTag", func(t *testing.T) {
		source := `###START###search text###END###replace text`
		searchTags := []*regexp.Regexp{
			regexp.MustCompile(`###START###`),
			regexp.MustCompile(`###END###`),
			regexp.MustCompile(`###DONE###`),
		}

		_, err := ParseIncrBlocks(source, searchTags)
		if err == nil {
			t.Error("Expected error for incomplete block, got nil")
		}
	})

	t.Run("MissingSecondReplaceTag", func(t *testing.T) {
		source := `###START###search text###END###replace text###DONE### ###START###search text`
		searchTags := []*regexp.Regexp{
			regexp.MustCompile(`###START###`),
			regexp.MustCompile(`###END###`),
			regexp.MustCompile(`###DONE###`),
		}

		_, err := ParseIncrBlocks(source, searchTags)
		if err == nil {
			t.Error("Expected error for incomplete block, got nil")
		}
	})

	t.Run("MissingSecondReplaceEndTag", func(t *testing.T) {
		source := `###START###search text###END###replace text###DONE### ###START###search text###END###`
		searchTags := []*regexp.Regexp{
			regexp.MustCompile(`###START###`),
			regexp.MustCompile(`###END###`),
			regexp.MustCompile(`###DONE###`),
		}

		_, err := ParseIncrBlocks(source, searchTags)
		if err == nil {
			t.Error("Expected error for incomplete block, got nil")
		}
	})

	// Test case 6: Invalid number of search tags
	t.Run("InvalidSearchTagsCount", func(t *testing.T) {
		source := "some text"
		searchTags := []*regexp.Regexp{
			regexp.MustCompile(`###START###`),
			regexp.MustCompile(`###END###`),
			// Missing third tag
		}

		_, err := ParseIncrBlocks(source, searchTags)
		if err == nil {
			t.Error("Expected error for invalid search tags count, got nil")
		}
	})

	// Test case 7: Tags in search field
	t.Run("TagsInSearchField", func(t *testing.T) {
		source := `###START###search ###START### text###END###replace text###DONE###`
		searchTags := []*regexp.Regexp{
			regexp.MustCompile(`###START###`),
			regexp.MustCompile(`###END###`),
			regexp.MustCompile(`###DONE###`),
		}

		_, err := ParseIncrBlocks(source, searchTags)
		if err == nil {
			t.Error("Expected error for tags in search field, got nil")
		}
	})

	// Test case 8: Tags in replace field
	t.Run("TagsInReplaceField", func(t *testing.T) {
		source := `###START###search text###END###replace ###END### text###DONE###`
		searchTags := []*regexp.Regexp{
			regexp.MustCompile(`###START###`),
			regexp.MustCompile(`###END###`),
			regexp.MustCompile(`###DONE###`),
		}

		_, err := ParseIncrBlocks(source, searchTags)
		if err == nil {
			t.Error("Expected error for tags in replace field, got nil")
		}
	})

	// Test case 9: Complex regex patterns
	t.Run("ComplexRegexPatterns", func(t *testing.T) {
		source := `<!-- SEARCH -->pattern1<!-- /SEARCH --><!-- REPLACE -->replacement1<!-- /REPLACE -->
<!-- SEARCH -->pattern2<!-- /SEARCH --><!-- REPLACE -->replacement2<!-- /REPLACE -->`

		searchTags := []*regexp.Regexp{
			regexp.MustCompile(`<!-- SEARCH -->`),
			regexp.MustCompile(`<!-- /SEARCH -->`),
			regexp.MustCompile(`<!-- /REPLACE -->`),
		}

		blocks, err := ParseIncrBlocks(source, searchTags)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(blocks) != 2 {
			t.Fatalf("Expected 2 blocks, got %d", len(blocks))
		}

		expectedBlocks := []IncrBlock{
			{Search: "pattern1", Replace: "<!-- REPLACE -->replacement1"},
			{Search: "pattern2", Replace: "<!-- REPLACE -->replacement2"},
		}

		for i, block := range blocks {
			if block.Search != expectedBlocks[i].Search {
				t.Errorf("Block %d: expected search '%s', got '%s'", i, expectedBlocks[i].Search, block.Search)
			}
			if block.Replace != expectedBlocks[i].Replace {
				t.Errorf("Block %d: expected replace '%s', got '%s'", i, expectedBlocks[i].Replace, block.Replace)
			}
		}
	})

	// Test case 10: Empty search and replace content
	t.Run("EmptyContent", func(t *testing.T) {
		source := `###S######E######R###`
		searchTags := []*regexp.Regexp{
			regexp.MustCompile(`###S###`),
			regexp.MustCompile(`###E###`),
			regexp.MustCompile(`###R###`),
		}

		blocks, err := ParseIncrBlocks(source, searchTags)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(blocks) != 1 {
			t.Fatalf("Expected 1 block, got %d", len(blocks))
		}

		if blocks[0].Search != "" {
			t.Errorf("Expected empty search, got '%s'", blocks[0].Search)
		}
		if blocks[0].Replace != "" {
			t.Errorf("Expected empty replace, got '%s'", blocks[0].Replace)
		}
	})
}
