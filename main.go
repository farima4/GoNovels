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
	"time" // Main markdown package

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"   // HTML renderer
	"github.com/gomarkdown/markdown/parser" // Markdown parser
)

// GLOBAL VARIABLES
var (
	port         = "4040"
	novels       []Novel
	mutex        sync.RWMutex
	lastScan     time.Time
	scanDelay    = time.Minute * 5 // 5 minutes by default
	defaultCover = "/static/cover.png"

	indexTemplate   *template.Template
	novelTemplate   *template.Template
	chapterTemplate *template.Template

	mdParser   *parser.Parser
	mdRenderer *html.Renderer
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
	Number  int
	Title   string
	Content template.HTML
}

// ---------- FUNCTIONS

func main() {
	fmt.Println("Starting the webserver at port :" + port)

	// Initial scan
	novels = scanNovels()
	lastScan = time.Now()

	// loading templates
	var err error
	indexTemplate, err = template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal("Failed to parse template:", err)
	}

	novelTemplate, err = template.ParseFiles("templates/novel.html")
	if err != nil {
		log.Fatal("Failed to parse template:", err)
	}

	chapterTemplate, err = template.ParseFiles("templates/chapter.html")
	if err != nil {
		log.Fatal("Failed to parse template:", err)
	}

	// initializing the parser and renderer

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	mdRenderer = html.NewRenderer(opts)

	// handling http requests
	http.HandleFunc("/", homePageHandler)
	http.HandleFunc("/novel/", novelPageHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/novels/", http.StripPrefix("/novels/", http.FileServer(http.Dir("novels"))))

	// starting server
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
	parts := strings.Split(strings.Trim(reqest.URL.Path, "/"), "/")

	/*
		println("---")
		for _, i := range parts {
			println(i)
		}
		println("---")
	*/

	slug := parts[1]

	//Handles chapter and media requests in chapters
	if len(parts) == 4 {
		if strings.Contains(parts[3], ".") { //checks if it is a file
			chapterMediaHandler(response, reqest, slug, parts[3])
		} else { //otherwise it is a chapter
			chapterPageHandler(response, reqest, slug, parts[3])
		}

		return
	}

	//Continues if it isn't a chapter
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
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func chapterPageHandler(response http.ResponseWriter, reqest *http.Request, slug string, num string) {
	filepath := filepath.Join("novels", slug, "chapter-"+num+".md")
	content, err := os.ReadFile(filepath)

	if err != nil {
		http.NotFound(response, reqest)
		return
	}

	// Split into first line and the rest
	parts := strings.SplitN(string(content), "\n", 2)
	var mdContent []byte
	if len(parts) == 2 && strings.HasPrefix(parts[0], "# ") {
		// First line is a heading – skip it
		mdContent = []byte(parts[1])
	} else {
		// No heading or unexpected format – keep whole file
		mdContent = content
	}

	// Convert to HTML without the duplicate title
	htmlContent := markdownToHTML(mdContent)

	var found bool = false
	var targetNovel *Novel
	for i := range novels {
		if novels[i].Slug == slug {
			targetNovel = &novels[i]
			found = true
			break
		}
	}
	if !found {
		http.NotFound(response, reqest)
		return
	}

	chnum, err := strconv.Atoi(num)
	if err != nil {
		http.NotFound(response, reqest)
		return
	}

	found = false
	var targetChapter *Chapter
	for i := range targetNovel.Chapters {
		if targetNovel.Chapters[i].Number == chnum {
			targetChapter = &targetNovel.Chapters[i]
			found = true
			break
		}
	}
	if !found {
		http.NotFound(response, reqest)
		return
	}

	chapter := Chapter{
		Number:  chnum,
		Title:   targetChapter.Title, // from novel.Chapters list
		Content: htmlContent,
	}

	// Execute chapter template with this chapter data
	err = chapterTemplate.Execute(response, map[string]interface{}{
		"Novel":   targetNovel,
		"Chapter": chapter,
	})
}

func chapterMediaHandler(response http.ResponseWriter, reqest *http.Request, slug string, media string) {
	filepath := filepath.Join("novels", slug, "media", media)
	//println(filepath)

	http.ServeFile(response, reqest, filepath)
}

// ---------- parsing markdown
func markdownToHTML(md []byte) template.HTML {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	mdParser = parser.NewWithExtensions(extensions)

	doc := mdParser.Parse(md)
	htmlContent := markdown.Render(doc, mdRenderer)
	return template.HTML(htmlContent)
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
				var coverPath = "novels/" + slug + "/media/" + meta.Cover
				author = meta.Author

				if _, err := os.Stat(coverPath); err != nil {
					println("cover error at" + coverPath + ", using default")
					cover = defaultCover
				} else {
					cover = "/novel/" + slug + "/chapter/" + meta.Cover
				}
			}
		}

		if title == "" {
			title = strings.ReplaceAll(slug, "-", " ")
			title = strings.ToTitle(title)
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
