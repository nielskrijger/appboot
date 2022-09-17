package test

import (
	"encoding/json"
	"strings"
)

type Logger struct {
	out []byte
}

func (log *Logger) Write(p []byte) (int, error) {
	log.out = append(log.out, p...)

	return len(p), nil
}

func (log *Logger) Lines() []map[string]any {
	lines := strings.Split(strings.TrimSpace(string(log.out)), "\n")
	result := make([]map[string]any, 0, len(lines))

	for _, line := range lines {
		jsonMap := make(map[string]any)
		_ = json.Unmarshal([]byte(line), &jsonMap)
		result = append(result, jsonMap)
	}

	return result
}

func (log *Logger) LastLine() map[string]any {
	lines := log.Lines()

	return lines[len(lines)-1]
}
