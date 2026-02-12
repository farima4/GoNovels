package main

import (
    "fmt"
    "html/template"      // HTML templating
    "io/fs"             // File system interface
    "net/http"          // HTTP server and client
//    "os"                // File operations
//    "path/filepath"     // Path manipulation
//    "strings"           // String utilities
	"log"
	"sync"
	"time"
//    "github.com/gomarkdown/markdown"            // Main markdown package
//    "github.com/gomarkdown/markdown/html"       // HTML renderer
//    "github.com/gomarkdown/markdown/parser"     // Markdown parser
);

// GLOBAL VARIABLES
var (
	port = "4040";
	novels []Novel;
	mutex sync.RWMutex;
	lastScan time.Time;
);

//CUSTOM TYPES
type Novel struct{
	ID string;
	Title string;
	Description string;
	Cover string;
	Chapters []Chapter;
};

type Chapter struct{
	ID string;
	Title string;
	Content template.HTML;
	Slug string;
};

// FUNCTIONS
func main(){
	fmt.Println("Starting the webserver at port :" + port);

	novels = scanNovels();
	lastScan = time.Now();

	http.HandleFunc("/", homePageHandler);

	log.Fatal(http.ListenAndServe(":" + port, nil));
}

func homePageHandler(response http.ResponseWriter, request *http.Request){
	tmpl, err := template.ParseFiles("templates/index.html");

	if err != nil { //desio se error
		http.Error(response, err.Error(), http.StatusInternalServerError);
		return;
	}

	err = tmpl.Execute(response, nil);
	if err != nil { //desio se error
		http.Error(response, err.Error(), http.StatusInternalServerError);
	}
}

func getNovels []Novel(){
	mutex.RLock();
	if time.Since(lastScan) < 5*time.Minute {
		defer mutex.RUnlock
		return novels;
	}
	mutex.RUnlock();

	mutex.Lock();
	defer mutex.Unlock();
	novels = scanNovels();
	lastScan = time.Now();
}

func scanNovels []Novel{
	
}
