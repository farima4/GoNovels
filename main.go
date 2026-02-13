package main

import (
	"encoding/json"
	"fmt"
	"html/template" // HTML templating

	// File system interface
	"net/http"      // HTTP server and client
	"os"            // File operations
	"path/filepath" // Path manipulation
	"strings"       // String utilities

	"log"
	"sync"
	"time"
	// "github.com/gomarkdown/markdown"            // Main markdown package
	// "github.com/gomarkdown/markdown/html"       // HTML renderer
	// "github.com/gomarkdown/markdown/parser"     // Markdown parser
)

// GLOBAL VARIABLES
var (
	port     = "4040"
	novels   []Novel
	mutex    sync.RWMutex
	lastScan time.Time
)

// CUSTOM TYPES
type Novel struct {
	Slug        string
	Title       string
	Description string
	Cover       string
	Chapters    []Chapter
}

type Chapter struct {
	Slug    string
	Title   string
	Content template.HTML
}

// ---------- FUNCTIONS

func main() {
	fmt.Println("Starting the webserver at port :" + port)

	novels = scanNovels()
	lastScan = time.Now()

	http.HandleFunc("/", homePageHandler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homePageHandler(response http.ResponseWriter, request *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")

	if err != nil { //desio se error
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(response, nil)
	if err != nil { //desio se error
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

// ---------- loading novels
func getNovels() {
	mutex.RLock()
	if time.Since(lastScan) < 5*time.Minute {
		defer mutex.RUnlock()
		return
	}
	mutex.RUnlock()

	mutex.Lock()
	defer mutex.Unlock()
	novels = scanNovels()
	lastScan = time.Now()
}

func scanNovels() []Novel {
	var novels []Novel

	entries, err := os.ReadDir("novels")
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		slug := entry.Name()
		novelPath := filepath.Join("novels", slug)
		metadataPath := filepath.Join(novelPath, "metadata.json")

		var title, description string

		title = description // so it compiles, remove later

		if data, err := os.ReadFile(metadataPath); err == nil {
			var meta struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			}

			if err := json.Unmarshal(data, &meta); err == nil {
				title = meta.Title
				description = meta.Description
			}
		}

		if title == "" {
			title = strings.ReplaceAll(slug, "-", " ")
			title = strings.Title(title)
		}
	}

	return novels
}
