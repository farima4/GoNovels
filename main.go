package main

import (
    "fmt"
    "html/template"      // HTML templating
//    "io/fs"             // File system interface
    "net/http"          // HTTP server and client
//    "os"                // File operations
//    "path/filepath"     // Path manipulation
//    "strings"           // String utilities
	"log"
//    "github.com/gomarkdown/markdown"            // Main markdown package
//    "github.com/gomarkdown/markdown/html"       // HTML renderer
//    "github.com/gomarkdown/markdown/parser"     // Markdown parser
)

var (
	port = "4040"
)

func main(){
	fmt.Println("Starting the webserver");

	http.HandleFunc("/", homePageHandler);

	log.Fatal(http.ListenAndServe(":" + port, nil));
}

func homePageHandler(response http.ResponseWriter, request *http.Request){
	tmpl, err := template.ParseFiles("templates/index.html");

	if err != nil { //desio se error
		http.Error(response, err.Error(), http.StatusInternalServerError);
		return;
	}

	err = tmpl.Execute(response, nil)
	if err != nil { //desio se error
		http.Error(response, err.Error(), http.StatusInternalServerError);
	}
}


