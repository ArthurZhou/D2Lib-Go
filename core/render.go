package core

import (
	"encoding/json"
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"regexp"
	"strings"
)

/*
render.go
Handle and render normal requests
*/

type ExtraData struct {
	Title string `json:"title"`
}

func RequestHandler() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		log := GetLogger()
		// check if the user has logged in
		userName := getUserName(request)
		if userName == "" && os.Getenv("D2LIB_elogn") == "true" { // not logged in
			log.Tracef("[%s] > redirect because not logged in", request.RemoteAddr)
			http.Redirect(response, request, "/login", 302)
		} else { // logged in
			reqURL := request.URL.Path // get doc path
			if len(reqURL) > 1 {
				log.Tracef("[%s] > request for doc: %s", request.RemoteAddr, reqURL)
				if _, err := os.Stat(os.Getenv("D2LIB_root") + "/" + os.Getenv("D2LIB_sloc") + reqURL); !os.IsNotExist(err) {
					// url exists
					filePath := os.Getenv("D2LIB_root") + "/" + os.Getenv("D2LIB_sloc") + reqURL
					fileByte, _ := os.ReadFile(filePath) // read file as byte array
					fileText := string(fileByte)         // convert byte array to string
					if reqURL[len(reqURL)-3:] == ".md" { // is this a markdown file?
						if strings.Contains(fileText, "<d2lib>") && strings.Contains(fileText, "</d2lib>") {
							extraDataRaw := regexp.MustCompile(`<d2lib>\r*\n*((.|\n)*)\r*\n*</d2lib>`).FindString(fileText)
							fileByte = regexp.MustCompile(`<d2lib>\r*\n*((.|\n)*)\r*\n*</d2lib>`).ReplaceAll(fileByte, []byte(""))
							var extraData ExtraData
							_ = json.Unmarshal([]byte(extraDataRaw[7:len(extraDataRaw)-8]), &extraData)
							fileText = strings.ReplaceAll(os.Getenv("D2LIB_ipage"), "{{ TITLE }}", extraData.Title)
						} else {
							splPath := strings.Split(reqURL, "/")
							fileName := strings.Join(splPath[len(splPath)-1:], "")
							fileText = strings.ReplaceAll(os.Getenv("D2LIB_ipage"), "{{ TITLE }}", fileName)
						}
						// render markdown to html and put it into the template
						fileText = strings.ReplaceAll(fileText, "{{ CONTENT }}", string(markdown.ToHTML(fileByte, nil, nil)))
						fileText = strings.ReplaceAll(fileText, "{{ MENU }}", os.Getenv("D2LIB_menu"))
					} else if reqURL[len(reqURL)-5:] == ".html" { // is this a html file?
						splPath := strings.Split(reqURL, "/")
						fileName := strings.Join(splPath[len(splPath)-1:], "")
						// replace hooks
						fileText = strings.ReplaceAll(string(fileByte), "{{ TITLE }}", fileName)
						fileText = strings.ReplaceAll(fileText, "{{ MENU }}", os.Getenv("D2LIB_menu"))
						fileText = strings.ReplaceAll(fileText, "{{ STYLE }}", "<style>"+os.Getenv("D2LIB_istyle")+"</style>")
					}
					_, _ = fmt.Fprint(response, fileText) // output to http.ResponseWriter
				} else {
					// url does not exist
					log.Tracef("[%s] > url does not exist: %s", request.RemoteAddr, reqURL)
					FnfHandler(response, request)
				}
			} else {
				log.Tracef("[%s] > blank url", request.RemoteAddr)
				FnfHandler(response, request)
			}
		}
	}
}

func TemplatesHandler() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		log := GetLogger()
		log.Tracef("[%s] > request for assets assets: %s", request.RemoteAddr, mux.Vars(request)["path"])
		fileByte, _ := os.ReadFile(os.Getenv("D2LIB_root") + "/assets" + request.URL.Path)
		_, _ = fmt.Fprint(response, string(fileByte))
	}
}

func RedirectHandler() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		// if request for root, redirect to home page
		http.Redirect(response, request, "/docs/"+os.Getenv("D2LIB_hpage"), 302)
	}
}

func FaviconHandler() http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) { // handle favicon
		log := GetLogger()
		log.Tracef("[%s] > request for favicon", request.RemoteAddr)
		http.ServeFile(response, request, os.Getenv("D2LIB_root")+"/assets/favicon.ico")
	}
}

func FnfHandler(response http.ResponseWriter, _ *http.Request) { // handle 404 page
	response.WriteHeader(404)
	_, _ = response.Write([]byte(os.Getenv("D2LIB_fnf")))
}
