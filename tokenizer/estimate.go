package tokenizer

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// EstimateTokenCount estimates the number of tokens in a text string
// This is a simplified estimation based on English text patterns
// A more accurate count would require a proper tokenizer like tiktoken
func EstimateTokenCount(text string) int {
	if text == "" {
		return 0
	}

	// Count Unicode characters
	runeCount := utf8.RuneCountInString(text)

	// Average tokens per character varies by language and content
	// For English text, roughly 4 characters per token is a common approximation
	// We'll use a more refined approach based on word and character patterns

	// Count words (sequences separated by whitespace)
	words := strings.Fields(text)
	wordCount := len(words)

	// Estimate based on combined heuristics
	// Rule of thumb: ~4 characters per token for English
	// But punctuation and special characters add overhead

	// Adjust for code-like content (often has more tokens per character)
	isCode := strings.Contains(text, "func ") ||
		strings.Contains(text, "function ") ||
		strings.Contains(text, "def ") ||
		strings.Contains(text, "class ") ||
		strings.Contains(text, "{") ||
		strings.Contains(text, "}")

	// Adjust for non-English content
	isChinese := regexp.MustCompile(`[\u4e00-\u9fa5]`).MatchString(text)
	isJapanese := regexp.MustCompile(`[\u3040-\u309f\u30a0-\u30ff]`).MatchString(text)
	isKorean := regexp.MustCompile(`[\uac00-\ud7af]`).MatchString(text)

	if isCode {
		// Code tends to have more tokens per character
		// Use 2.5 characters per token for code
		return max(1, int(float64(len(text))/2.5))
	}

	if isChinese || isJapanese || isKorean {
		// CJK characters typically represent one token each
		return runeCount
	}

	// For English and other Latin-script languages
	// Use a combination of character count and word count
	// This accounts for punctuation and special characters
	charBased := int(float64(len(text)) / 4.0)
	wordBased := wordCount * 2 // Average 2 tokens per word

	// Average of both estimates
	return max(1, (charBased+wordBased)/2)
}

// EstimateMessagesTokenCount estimates token count for a list of messages
func EstimateMessagesTokenCount(messages []map[string]interface{}) int {
	total := 0

	for _, msg := range messages {
		// Add overhead for message structure (role, etc.)
		total += 4 // ~4 tokens for message wrapper

		if role, ok := msg["role"].(string); ok {
			total += EstimateTokenCount(role)
		}

		content := msg["content"]
		switch c := content.(type) {
		case string:
			total += EstimateTokenCount(c)
		case []interface{}:
			for _, part := range c {
				if partMap, ok := part.(map[string]interface{}); ok {
					if text, ok := partMap["text"].(string); ok {
						total += EstimateTokenCount(text)
					}
					// Images add significant token overhead
					if _, ok := partMap["image_url"]; ok {
						total += 100 // Rough estimate for image tokens
					}
				}
			}
		}
	}

	// Add completion overhead
	total += 3

	return total
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// EstimateEmbeddingInput estimates token count for embedding input
func EstimateEmbeddingInput(input interface{}) int {
	switch v := input.(type) {
	case string:
		return EstimateTokenCount(v)
	case []interface{}:
		total := 0
		for _, item := range v {
			if s, ok := item.(string); ok {
				total += EstimateTokenCount(s)
			}
		}
		return total
	default:
		return 0
	}
}
