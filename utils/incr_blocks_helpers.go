package utils

import (
	"errors"
	"regexp"
)

type IncrBlock struct {
	Search  string
	Replace string
}

func ParseIncrBlocks(source string, searchTags []*regexp.Regexp) ([]IncrBlock, error) {
	if len(searchTags) != 3 {
		return nil, errors.New("searchTags must contain exactly 3 regexps")
	}

	searchStartTag := searchTags[0]
	searchEndTag := searchTags[1]
	replaceEndTag := searchTags[2]

	var blocks []IncrBlock
	remainingText := source

	for remainingText != "" {
		// Find search start tag
		searchStartMatch := searchStartTag.FindStringIndex(remainingText)
		if searchStartMatch == nil {
			if len(blocks) == 0 {
				return nil, errors.New("no search start tag found")
			}
			break // No more blocks found
		}
		// Extract text after search start tag
		remainingText = remainingText[searchStartMatch[1]:]
		// Find search end tag
		searchEndMatch := searchEndTag.FindStringIndex(remainingText)
		if searchEndMatch == nil {
			return nil, errors.New("incomplete block: search end tag not found")
		}
		// Extract search text
		searchText := remainingText[:searchEndMatch[0]]
		// Extract text after search end tag
		remainingText = remainingText[searchEndMatch[1]:]
		// Find replace end tag
		replaceEndMatch := replaceEndTag.FindStringIndex(remainingText)
		if replaceEndMatch == nil {
			return nil, errors.New("incomplete block: replace end tag not found")
		}
		// Extract replace text (everything between search end tag and replace end tag)
		replaceText := remainingText[:replaceEndMatch[0]]
		// Create and append the block
		block := IncrBlock{
			Search:  searchText,
			Replace: replaceText,
		}
		blocks = append(blocks, block)
		// Move remaining text past the current block
		remainingText = remainingText[replaceEndMatch[1]:]
	}

	if len(blocks) == 0 {
		return nil, errors.New("no valid blocks found")
	}

	//validate blocks
	for _, block := range blocks {
		// Check Search field for any of the tags
		if searchStartTag.MatchString(block.Search) {
			return nil, errors.New("search start tag found in Search field of block")
		}
		if searchEndTag.MatchString(block.Search) {
			return nil, errors.New("search end tag found in Search field of block")
		}
		if replaceEndTag.MatchString(block.Search) {
			return nil, errors.New("replace end tag found in Search field of block")
		}
		// Check Replace field for any of the tags
		if searchStartTag.MatchString(block.Replace) {
			return nil, errors.New("search start tag found in Replace field of block")
		}
		if searchEndTag.MatchString(block.Replace) {
			return nil, errors.New("search end tag found in Replace field of block")
		}
		if replaceEndTag.MatchString(block.Replace) {
			return nil, errors.New("replace end tag found in Replace field of block")
		}
	}

	return blocks, nil
}

func FilterIncrBlocks(incrBlocks []IncrBlock, startTags []*regexp.Regexp, endTags []*regexp.Regexp) []IncrBlock {
	filteredBlocks := make([]IncrBlock, len(incrBlocks))
	for i, block := range incrBlocks {
		search := GetTextAfterFirstMatchesRx(block.Search, startTags)
		search = GetTextBeforeLastMatchesRx(search, endTags)
		replace := GetTextAfterFirstMatchesRx(block.Replace, startTags)
		replace = GetTextBeforeLastMatchesRx(replace, endTags)
		filteredBlocks[i] = IncrBlock{Search: search, Replace: replace}
	}
	return filteredBlocks
}
