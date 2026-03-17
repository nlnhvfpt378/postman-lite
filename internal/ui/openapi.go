package ui

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"postman-lite/internal/state"
)

type openAPIDocument struct {
	OpenAPI string                                 `json:"openapi"`
	Servers []openAPIServer                        `json:"servers"`
	Paths   map[string]map[string]openAPIOperation `json:"paths"`
}

type openAPIServer struct {
	URL string `json:"url"`
}

type openAPIOperation struct {
	Summary     string              `json:"summary"`
	OperationID string              `json:"operationId"`
	RequestBody *openAPIRequestBody `json:"requestBody"`
}

type openAPIRequestBody struct {
	Content map[string]openAPIMediaType `json:"content"`
}

type openAPIMediaType struct {
	Example  any                       `json:"example"`
	Examples map[string]openAPIExample `json:"examples"`
	Schema   *openAPISchema            `json:"schema"`
}

type openAPIExample struct {
	Value any `json:"value"`
}

type openAPISchema struct {
	Example any `json:"example"`
}

func ParseOpenAPIJSON(input string, nextIndex *int) ([]state.RequestTabState, error) {
	var doc openAPIDocument
	if err := json.Unmarshal([]byte(input), &doc); err != nil {
		return nil, fmt.Errorf("OpenAPI JSON 解析失败: %w", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(doc.OpenAPI), "3.") {
		return nil, fmt.Errorf("仅支持 OpenAPI 3 JSON")
	}
	if len(doc.Paths) == 0 {
		return nil, fmt.Errorf("未发现 paths")
	}

	baseURL := ""
	if len(doc.Servers) > 0 {
		baseURL = strings.TrimSpace(doc.Servers[0].URL)
	}

	paths := make([]string, 0, len(doc.Paths))
	for path := range doc.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	methods := []string{"get", "post", "put", "patch", "delete", "head", "options"}
	items := make([]state.RequestTabState, 0)
	for _, path := range paths {
		opMap := doc.Paths[path]
		for _, method := range methods {
			op, ok := opMap[method]
			if !ok {
				continue
			}
			title := strings.TrimSpace(op.Summary)
			if title == "" {
				title = strings.TrimSpace(op.OperationID)
			}
			if title == "" {
				title = fmt.Sprintf("%s %s", strings.ToUpper(method), path)
			}
			body, bodyContentType := extractExampleBody(op.RequestBody)
			headers := ""
			if bodyContentType != "" {
				headers = fmt.Sprintf("Content-Type: %s", bodyContentType)
			}
			items = append(items, state.RequestTabState{
				ID:      fmt.Sprintf("tab-%d", *nextIndex),
				Title:   title,
				Method:  strings.ToUpper(method),
				URL:     joinBaseURL(baseURL, path),
				Headers: headers,
				Body:    body,
			})
			*nextIndex = *nextIndex + 1
		}
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("未发现可导入的 HTTP 操作")
	}
	return items, nil
}

func extractExampleBody(body *openAPIRequestBody) (string, string) {
	if body == nil || len(body.Content) == 0 {
		return "", ""
	}
	preferred := []string{"application/json", "application/*+json", "text/plain"}
	for _, contentType := range preferred {
		for actualType, media := range body.Content {
			if contentType == actualType || (contentType == "application/*+json" && strings.HasSuffix(actualType, "+json")) {
				return formatExample(media), actualType
			}
		}
	}
	keys := make([]string, 0, len(body.Content))
	for k := range body.Content {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	first := body.Content[keys[0]]
	return formatExample(first), keys[0]
}

func formatExample(media openAPIMediaType) string {
	candidate := media.Example
	if candidate == nil {
		for _, ex := range media.Examples {
			candidate = ex.Value
			break
		}
	}
	if candidate == nil && media.Schema != nil {
		candidate = media.Schema.Example
	}
	if candidate == nil {
		return ""
	}
	data, err := json.MarshalIndent(candidate, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", candidate)
	}
	return string(data)
}

func joinBaseURL(baseURL, path string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return path
	}
	if parsed, err := url.Parse(baseURL); err == nil {
		ref, refErr := url.Parse(path)
		if refErr == nil {
			return parsed.ResolveReference(ref).String()
		}
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
}
