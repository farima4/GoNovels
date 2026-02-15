package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	// "github.com/gomarkdown/markdown"            // Main markdown package
	// "github.com/gomarkdown/markdown/html"       // HTML renderer
	// "github.com/gomarkdown/markdown/parser"     // Markdown parser
)

// GLOBAL VARIABLES
var (
	port         = "4040"
	novels       []Novel
	mutex        sync.RWMutex
	lastScan     time.Time
	scanDelay    = time.Minute * 5 // should be 5 minutes,
	defaultCover = "/static/cover.png"

	indexTemplate *template.Template
	novelTemplate *template.Template
)

// CUSTOM TYPES
type Novel struct {
	Slug         string
	Title        string
	Description  string
	Cover        string
	Author       string
	NovelPath    string
	Chapters     []Chapter
	ChapterCount int
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

	var err error
	indexTemplate, err = template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal("Failed to parse template:", err)
	}

	novelTemplate, err = template.ParseFiles("templates/novel.html")
	if err != nil {
		log.Fatal("Failed to parse template:", err)
	}

	http.HandleFunc("/", homePageHandler)
	http.HandleFunc("/novel/", novelPageHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/novels/", http.StripPrefix("/novels/", http.FileServer(http.Dir("novels"))))

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homePageHandler(response http.ResponseWriter, request *http.Request) {
	getNovels()

	err := indexTemplate.Execute(response, novels)
	if err != nil { //desio se error
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func novelPageHandler(response http.ResponseWriter, reqest *http.Request) {
	slug := strings.TrimPrefix(reqest.URL.Path, "/novel/")
	if slug == "" {
		http.NotFound(response, reqest)
		return
	}

	var found *Novel
	for i := range novels {
		if novels[i].Slug == slug {
			found = &novels[i]
			break
		}
	}

	if found == nil {
		http.NotFound(response, reqest)
		return
	}

	err := novelTemplate.Execute(response, found)
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

		var title, description, cover, author string

		if data, err := os.ReadFile(metadataPath); err == nil {
			var meta struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				Cover       string `json:"cover"`
				Author      string `json:"author"`
			}

			if err := json.Unmarshal(data, &meta); err == nil {
				title = meta.Title
				description = meta.Description
				cover = "/novels/" + slug + "/media/" + meta.Cover
				author = meta.Author

				if _, err := os.Stat(cover); err != nil {
					cover = defaultCover
				}
			}
		}

		if title == "" {
			title = strings.ReplaceAll(slug, "-", " ")
			title = strings.ToTitle(title)
			cover = "/static/cover.png"
			author = "unavailable"
		}

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
			Title:        title,
			Description:  description,
			Cover:        cover,
			Slug:         slug,
			Author:       author,
			NovelPath:    novelPath,
			Chapters:     chapters,
			ChapterCount: len(chapters),
		})

		//debugging
		fmt.Println("Title: " + title)
		fmt.Println("Description: " + description)
		fmt.Println("slug: " + slug)
		fmt.Println("Author: " + author)
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
