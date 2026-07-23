package utils

import (
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

func ParseRequest(payload string, target interface{}) error {
	return json.Unmarshal([]byte(payload), target)
}

func MarshalResponse(data interface{}) (string, error) {
	output, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func LogAndReturnRuntime(logger runtime.Logger, message string, err error, code int) (string, error) {
	if err != nil {
		logger.Error("%s: %v", message, err)
	} else {
		logger.Error(message)
	}
	return "", runtime.NewError(message, code)
}
