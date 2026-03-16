package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type FormState struct {
	Method  string `json:"method"`
	URL     string `json:"url"`
	Headers string `json:"headers"`
	Body    string `json:"body"`
}

func Default() FormState {
	return FormState{
		Method:  "GET",
		URL:     "https://httpbin.org/anything",
		Headers: "Content-Type: application/json",
		Body:    "{\n  \"hello\": \"world\"\n}",
	}
}

func filePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "postman-lite", "state.json"), nil
}

func Load() (FormState, error) {
	path, err := filePath()
	if err != nil {
		return Default(), err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return Default(), err
	}
	st := Default()
	if err := json.Unmarshal(data, &st); err != nil {
		return Default(), err
	}
	if st.Method == "" {
		st.Method = "GET"
	}
	return st, nil
}

func Save(st FormState) error {
	path, err := filePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func Path() (string, error) {
	return filePath()
}
