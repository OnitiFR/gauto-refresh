package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func serveScript(w http.ResponseWriter, req *http.Request) {
	str := `
	const es = new EventSource(sse_url)
	es.addEventListener("ask-refresh", (ev) => {
		###
	})
	`
	str = strings.Replace(str, "###", Action, 1)

	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "no-cache")
	fmt.Fprintf(w, "var sse_url='http://%s/sse'", ListenPort)
	w.Write([]byte(str))
}

func serveSSE(w http.ResponseWriter, req *http.Request) {
	log.Println("client connected")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	for _, file := range Files {
		stat, err := os.Stat(file)
		if err != nil {
			log.Fatal(err)
		}

		if stat.Mode().IsDir() {
			filepath.Walk(file, func(path string, fi fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if fi.Mode().IsDir() {
					err := watcher.Add(path)
					if err != nil {
						log.Printf("error: %s", err)
					}
				}
				return nil
			})
		} else {
			err = watcher.Add(file)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)

				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("ask client refresh")
					fmt.Fprintf(w, "event: ask-refresh\n")
					fmt.Fprintf(w, "data: {}\n")
					fmt.Fprintf(w, "\n\n")
					flusher.Flush()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			case <-req.Context().Done():
				done <- true
			}
		}
	}()
	<-done
	log.Println("client disconnected")
}

func StartServer() {
	http.HandleFunc("/sse", serveSSE)
	http.HandleFunc("/refresh", serveScript)
	err := http.ListenAndServe(ListenPort, nil)
	if err != nil {
		log.Fatal(err)
	}
}
