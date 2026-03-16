package security

import "github.com/microcosm-cc/bluemonday"

var msgSanitizer = bluemonday.UGCPolicy()

// SanitizeMessageContent strips unsafe HTML from user input.
func SanitizeMessageContent(content string) string {
	return msgSanitizer.Sanitize(content)
}
