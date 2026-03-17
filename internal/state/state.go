package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type RequestTabState struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Method  string `json:"method"`
	URL     string `json:"url"`
	Headers string `json:"headers"`
	Body    string `json:"body"`
}

type FormState struct {
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	Headers      string            `json:"headers"`
	Body         string            `json:"body"`
	Tabs         []RequestTabState `json:"tabs,omitempty"`
	SelectedTab  int               `json:"selectedTab,omitempty"`
	NextTabIndex int               `json:"nextTabIndex,omitempty"`
}

func defaultTab() RequestTabState {
	return RequestTabState{
		ID:      "tab-1",
		Title:   "请求 1",
		Method:  "GET",
		URL:     "https://httpbin.org/anything",
		Headers: "Content-Type: application/json",
		Body:    "{\n  \"hello\": \"world\"\n}",
	}
}

func Default() FormState {
	tab := defaultTab()
	return FormState{
		Method:       tab.Method,
		URL:          tab.URL,
		Headers:      tab.Headers,
		Body:         tab.Body,
		Tabs:         []RequestTabState{tab},
		SelectedTab:  0,
		NextTabIndex: 2,
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
	normalize(&st)
	return st, nil
}

func Save(st FormState) error {
	normalize(&st)
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

func normalize(st *FormState) {
	if len(st.Tabs) == 0 {
		st.Tabs = []RequestTabState{defaultTab()}
	}
	for i := range st.Tabs {
		if strings.TrimSpace(st.Tabs[i].ID) == "" {
			st.Tabs[i].ID = fmt.Sprintf("tab-%d", i+1)
		}
		if strings.TrimSpace(st.Tabs[i].Method) == "" {
			st.Tabs[i].Method = "GET"
		}
		if strings.TrimSpace(st.Tabs[i].Title) == "" {
			st.Tabs[i].Title = fmt.Sprintf("请求 %d", i+1)
		}
	}
	if st.SelectedTab < 0 || st.SelectedTab >= len(st.Tabs) {
		st.SelectedTab = 0
	}
	if st.NextTabIndex <= len(st.Tabs) {
		st.NextTabIndex = len(st.Tabs) + 1
	}

	// 向后兼容旧版本单页签字段。
	if strings.TrimSpace(st.Method) != "" || strings.TrimSpace(st.URL) != "" || strings.TrimSpace(st.Headers) != "" || strings.TrimSpace(st.Body) != "" {
		current := &st.Tabs[st.SelectedTab]
		if current.Method == "" || current.Method == "GET" {
			if strings.TrimSpace(st.Method) != "" {
				current.Method = st.Method
			}
		}
		if strings.TrimSpace(current.URL) == "" && strings.TrimSpace(st.URL) != "" {
			current.URL = st.URL
		}
		if strings.TrimSpace(current.Headers) == "" && strings.TrimSpace(st.Headers) != "" {
			current.Headers = st.Headers
		}
		if strings.TrimSpace(current.Body) == "" && strings.TrimSpace(st.Body) != "" {
			current.Body = st.Body
		}
	}

	current := st.Tabs[st.SelectedTab]
	st.Method = current.Method
	st.URL = current.URL
	st.Headers = current.Headers
	st.Body = current.Body
}
