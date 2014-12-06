package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// ca := &CA{"goagent", 2048}
	// ca.Issue(true, "", 365*24*time.Hour)

	addr := ":1080"
	ln, err := Listen("tcp4", addr)
	if err != nil {
		log.Fatalf("Listen(\"tcp\", %s) failed: %s", addr, err)
	}
	h := Handler{
		Listener: ln,
		Log:      log.New(os.Stderr, "INFO - ", 3),
		Net:      &SimpleNetwork{},
		RequestPlugins: map[string]RequestPlugin{
			"direct": &DirectRequestPlugin{},
			"strip":  &StripRequestPlugin{},
		},
		ResponsePlugins: map[string]ResponsePlugin{
			"direct": &DirectResponsePlugin{},
		},
		RequestFilters: []RequestFilter{
			&StripRequestFilter{},
			&DirectRequestFilter{},
		},
		ResponseFilters: []ResponseFilter{
			&DirectResponseFilter{},
		},
	}
	s := &http.Server{
		Handler:        h,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	h.Log.Printf("ListenAndServe on %s\n", h.Listener.Addr().String())
	h.Log.Fatal(s.Serve(h.Listener))
}
