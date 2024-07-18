package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	response := os.Getenv("RESPONSE")
	if response == "" {
		response = "response"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	m := http.NewServeMux()
	c := 0
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		id := c
		c++
		log.Println("Start: /, id:", id)
		duration := 10 * time.Second
		if d, err := time.ParseDuration(r.FormValue("sleep")); err == nil {
			duration = d
		}
		log.Println("duration:", duration.String())
		time.Sleep(duration)
		w.Write([]byte(response))
		log.Println("Finish: /, id:", id)
	})

	if err := http.ListenAndServe(":"+port, m); err != nil {
		panic(err)
	}
}
