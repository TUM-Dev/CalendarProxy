package main

import (
	"github.com/tum-dev/calendar-proxy/internal"
	"log"
)

func main() {
	app := &internal.App{}
	log.Println(app.Run())
}
