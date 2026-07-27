// Package utils provides common utility functions for JSON processing and logging.
package utils

import (
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

// ParseRequest parses a JSON string payload into the specified target interface.
func ParseRequest(payload string, target interface{}) error {
	return json.Unmarshal([]byte(payload), target)
}

// MarshalResponse marshals the given data interface into a JSON string.
func MarshalResponse(data interface{}) (string, error) {
	output, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// LogAndReturnRuntime logs the error message via the logger and returns a standard Nakama runtime error with the specified code.
func LogAndReturnRuntime(logger runtime.Logger, message string, err error, code int) (string, error) {
	if err != nil {
		logger.Error("%s: %v", message, err)
	} else {
		logger.Error(message)
	}
	return "", runtime.NewError(message, code)
}
