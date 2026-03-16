package main

import (
	"log"

	appcore "postman-lite/internal/app"
	"postman-lite/internal/ui"
)

func main() {
	core := appcore.New()
	if err := ui.New(core).Run(); err != nil {
		log.Fatal(err)
	}
}
