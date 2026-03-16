package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"postman-lite/internal/model"
)

type Client struct {
	httpClient *http.Client
}

func New(timeout time.Duration) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}
	return &Client{httpClient: &http.Client{Timeout: timeout, Transport: transport}}
}

func (c *Client) Send(ctx context.Context, req model.Request) model.Response {
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, strings.TrimSpace(req.URL), bytes.NewBufferString(req.Body))
	if err != nil {
		return model.Response{Error: fmt.Sprintf("创建请求失败: %v", err)}
	}
	for _, h := range req.Headers {
		key := strings.TrimSpace(h.Key)
		if key == "" {
			continue
		}
		httpReq.Header.Set(key, h.Value)
	}

	started := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return model.Response{Error: fmt.Sprintf("请求失败: %v", err)}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Response{Error: fmt.Sprintf("读取响应失败: %v", err)}
	}

	return model.Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Duration:   time.Since(started),
		Headers:    resp.Header,
		Body:       string(bodyBytes),
		Size:       len(bodyBytes),
	}
}
