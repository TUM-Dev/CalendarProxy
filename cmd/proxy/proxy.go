package main

import (
	"fmt"
	"github.com/tum-dev/calendar-proxy/internal"
	"log"
	"net/http"
	"os"
)

func main() {
	// call /health if healthcheck is the first argument
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		resp, err := http.Get("http://cal-proxy:6001/health")
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != 200 {
			log.Fatal(fmt.Sprintf("healthcheck failed: %d %s", resp.StatusCode, resp.Body))
		}
		return
	}

	app := &internal.App{}
	log.Println(app.Run())
}
