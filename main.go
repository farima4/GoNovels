package main

import (
	"encoding/json"
	"fmt"
	"html/template" // HTML templating
	"sort"
	"strconv"

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
	port      = "4040"
	novels    []Novel
	mutex     sync.RWMutex
	lastScan  time.Time
	scanDelay = time.Minute * 5
)

// CUSTOM TYPES
type Novel struct {
	Slug        string
	Title       string
	Description string
	Cover       string
	Chapters    []Chapter
}

// all chapters must be named "chapter-n.md" where n is an int
type Chapter struct {
	Number int
	Title  string
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
	getNovels()

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
	if time.Since(lastScan) < scanDelay {
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
	var tmpnovels []Novel

	entries, err := os.ReadDir("novels")
	if err != nil {
		return novels
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		//Slug, Title, Description
		slug := entry.Name()
		novelPath := filepath.Join("novels", slug)
		metadataPath := filepath.Join(novelPath, "metadata.json")

		var title, description string

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
			title = strings.ToTitle(title)
		}

		//Cover:
		var cover string = filepath.Join(novelPath, "media", "cover.jpg")

		// ---------- Chapters:
		var chapters []Chapter

		chentries, err := os.ReadDir(novelPath)
		if err != nil {
			return novels
		}

		for _, chentry := range chentries {
			if chentry.IsDir() || !strings.HasSuffix(chentry.Name(), ".md") || !strings.HasPrefix(chentry.Name(), "chapter-") {
				continue
			}

			var number string = strings.TrimSuffix(strings.TrimPrefix(chentry.Name(), "chapter-"), ".md")

			var chtitle string
			content, _ := os.ReadFile(filepath.Join(novelPath, chentry.Name()))
			lines := strings.SplitN(string(content), "\n", 2)

			if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
				chtitle = strings.TrimPrefix(lines[0], "# ")
			} else {
				chtitle = strings.TrimSuffix(chentry.Name(), ".md")
			}

			n, _ := strconv.Atoi(number)
			chapters = append(chapters, Chapter{
				Title:  chtitle,
				Number: n,
			})
		}

		sort.Slice(chapters, func(i, j int) bool {
			return chapters[i].Number < chapters[j].Number
		})

		//Finishing the object
		tmpnovels = append(tmpnovels, Novel{
			Title:       title,
			Description: description,
			Cover:       cover,
			Slug:        slug,
			Chapters:    chapters,
		})

		//debugging
		fmt.Println("Title: " + title)
		fmt.Println("Description: " + description)
		fmt.Println("slug: " + slug)
		fmt.Println("Cover: " + cover)
		fmt.Println("ch number: " + strconv.Itoa(len(chapters)))
		fmt.Println("Chapter list:")
		for _, ch := range chapters {
			fmt.Println(strconv.Itoa(ch.Number) + " - " + ch.Title)
		}
		fmt.Println("------------------")
	}

	return tmpnovels
}
