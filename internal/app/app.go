package app

import (
	"time"

	"postman-lite/internal/httpclient"
)

type App struct {
	Client *httpclient.Client
}

func New() *App {
	return &App{Client: httpclient.New(60 * time.Second)}
}
