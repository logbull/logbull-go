package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	uuidPattern   = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	apiKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]{10,}$`)
)

const (
	maxMessageLength = 10_000
	maxFieldsCount   = 100
	maxFieldKeyLen   = 100
)

func ValidateProjectID(projectID string) error {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return fmt.Errorf("project ID cannot be empty")
	}

	if !uuidPattern.MatchString(projectID) {
		return fmt.Errorf(
			"invalid project ID format '%s'. Must be a valid UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
			projectID,
		)
	}

	return nil
}

func ValidateHostURL(host string) error {
	host = strings.TrimSpace(host)
	if host == "" {
		return fmt.Errorf("host URL cannot be empty")
	}

	parsedURL, err := url.Parse(host)
	if err != nil {
		return fmt.Errorf("invalid host URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("host URL must use http or https scheme, got: %s", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("host URL must have a host component")
	}

	return nil
}

func ValidateAPIKey(apiKey string) error {
	apiKey = strings.TrimSpace(apiKey)
	if len(apiKey) < 10 {
		return fmt.Errorf("API key must be at least 10 characters long")
	}

	if !apiKeyPattern.MatchString(apiKey) {
		return fmt.Errorf(
			"invalid API key format. API key must contain only alphanumeric characters, underscores, hyphens, and dots",
		)
	}

	return nil
}

func ValidateLogMessage(message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		return fmt.Errorf("log message cannot be empty")
	}

	if len(message) > maxMessageLength {
		return fmt.Errorf("log message too long (%d chars). Maximum allowed: %d", len(message), maxMessageLength)
	}

	return nil
}

func ValidateLogFields(fields map[string]any) error {
	if fields == nil {
		return nil
	}

	if len(fields) > maxFieldsCount {
		return fmt.Errorf("too many fields (%d). Maximum allowed: %d", len(fields), maxFieldsCount)
	}

	for key := range fields {
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("field key cannot be empty")
		}

		if len(key) > maxFieldKeyLen {
			return fmt.Errorf("field key too long (%d chars). Maximum: %d", len(key), maxFieldKeyLen)
		}
	}

	return nil
}
