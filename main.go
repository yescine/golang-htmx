package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/run", runHandler)
	http.HandleFunc("/stream", streamHandler)
	http.HandleFunc("/job_fields", jobFieldsHandler)

	// Serve static files (if any)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Templates
var templates = template.Must(template.ParseGlob("templates/*.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		job := r.FormValue("job")
		path := r.FormValue("path") // Additional input field for job2
		data := struct {
			Job  string
			Path string
		}{
			Job:  job,
			Path: path,
		}
		err := templates.ExecuteTemplate(w, "stream.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func streamHandler(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	path := r.URL.Query().Get("path")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	output := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		switch job {
		case "job1":
			simulateAPIPagination(output)
		case "job2":
			simulateDirectoryPagination(path, output)
		default:
			msg := "Invalid job selected."
			fmt.Println(msg) // Print to CLI
			output <- msg
		}
	}()

	// Close the output channel once the job is done
	go func() {
		wg.Wait()
		close(output)
	}()

	// Stream the output to the frontend
	for msg := range output {
		fmt.Fprintf(w, "data: %s\n\n", msg)
		flusher.Flush()
	}
}

func jobFieldsHandler(w http.ResponseWriter, r *http.Request) {
	job := r.FormValue("job")
	switch job {
	case "job2":
		// Need additional input field for path
		err := templates.ExecuteTemplate(w, "job2_fields.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		// No additional fields
		w.Write([]byte(""))
	}
}

// Dummy job functions
func simulateAPIPagination(output chan<- string) {
	for i := 1; i <= 5; i++ {
		time.Sleep(1 * time.Second) // Simulate API call delay
		msg := fmt.Sprintf("Fetched page %d from API", i)
		fmt.Println(msg) // Print to CLI
		output <- msg    // Send to frontend
	}
}

func simulateDirectoryPagination(path string, output chan<- string) {
	if path == "" {
		path = "." // default path
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		msg := fmt.Sprintf("Error reading directory: %v", err)
		fmt.Println(msg) // Print to CLI
		output <- msg    // Send to frontend
		return
	}
	for _, f := range files {
		if f.IsDir() {
			msg := fmt.Sprintf("Found directory: %s", f.Name())
			fmt.Println(msg)                   // Print to CLI
			output <- msg                      // Send to frontend
			time.Sleep(500 * time.Millisecond) // Simulate processing time
		}
	}
}
