package test

import (
	"encoding/json"
	"strings"
)

type Logger struct {
	out []byte
}

func (log *Logger) Write(p []byte) (n int, err error) {
	log.out = append(log.out, p...)

	return len(p), nil
}

func (log *Logger) Lines() (result []map[string]interface{}) {
	lines := strings.Split(strings.TrimSpace(string(log.out)), "\n")
	for _, line := range lines {
		jsonMap := make(map[string]interface{})
		_ = json.Unmarshal([]byte(line), &jsonMap)
		result = append(result, jsonMap)
	}

	return result
}

func (log *Logger) LastLine() (result map[string]interface{}) {
	lines := log.Lines()

	return lines[len(lines)-1]
}
