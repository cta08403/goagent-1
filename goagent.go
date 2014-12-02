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
	h := Handler{}
	h.RegisterPlugin("direct", &DirectPlugin{})
	h.RegisterPlugin("strip", &StripPlugin{})
	h.StackFilter(&DirectFilter{})
	h.StackFilter(&StripFilter{})
	h.Net = &SimpleNetwork{}
	h.L = ln
	h.Log = log.New(os.Stderr, "INFO - ", 3)
	s := &http.Server{
		Handler:        h,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	h.Log.Printf("ListenAndServe on %s\n", h.L.Addr().String())
	h.Log.Fatal(s.Serve(h.L))
}
