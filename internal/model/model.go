package model

import (
	"net/http"
	"time"
)

type HeaderKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Request struct {
	Method  string     `json:"method"`
	URL     string     `json:"url"`
	Headers []HeaderKV `json:"headers"`
	Body    string     `json:"body"`
}

type Response struct {
	StatusCode int         `json:"statusCode"`
	Status     string      `json:"status"`
	Duration   time.Duration `json:"duration"`
	Headers    http.Header `json:"headers"`
	Body       string      `json:"body"`
	Size       int         `json:"size"`
	Error      string      `json:"error,omitempty"`
}
