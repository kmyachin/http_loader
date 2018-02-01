package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type FileStatEntry struct {
	Name string
	Date string
	Size int64
}

var ModifiedTime string

func listDirectory(w http.ResponseWriter, r *http.Request) {
	LastModified := r.Header.Get("If-Modified-Since")

	if LastModified == ModifiedTime {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	Dir, err := os.Open(".")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "{\"code\": 400}")
		return
	}

	paths, err := Dir.Readdir(0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "{\"code\": 400}")
		return
	}

	w.Header().Add("Last-Modified", ModifiedTime)

	tmpl := template.Must(template.ParseFiles("files.html"))

	files := make([]FileStatEntry, 0, len(paths))
	for _, p := range paths {
		files = append(files, FileStatEntry{p.Name(), p.ModTime().Format("02-Jan-2006 15:04"), p.Size()})
	}

	tmpl.Execute(w, struct{ Files []FileStatEntry }{files})
}

func fileStat(w http.ResponseWriter, r *http.Request) {
	filename := r.FormValue("file")
	stat, err := os.Stat(filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "{\"code\": 400}")
		return
	}

	io.WriteString(w, "name: ")
	io.WriteString(w, stat.Name())
	if !stat.IsDir() {
		io.WriteString(w, " , size: ")
		io.WriteString(w, strconv.Itoa(int(stat.Size())))
		io.WriteString(w, "b")
	}
}

func accessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("[server] access: ", r.URL.Path, r.URL.RawQuery)
		next.ServeHTTP(w, r)
	})
}

func getRemoteDirectory(w http.ResponseWriter, r *http.Request) {
	var m FileListLoader_ModDir
	m.Index.Url = "http://localhost:8080/"
	ch := make(chan int)
	go m.reload(ch)
	time.Sleep(20 * time.Second)
	ch <- 1
	files := m.Index.Files
	io.WriteString(w, "remote directory file list: \n")
	for _, f := range files {
		io.WriteString(w, f.Name)
		io.WriteString(w, "\n")
	}
}

func main() {
	ModifiedTime = time.Now().Format(time.RFC822)

	mux := http.NewServeMux()
	mux.HandleFunc("/", listDirectory)
	mux.HandleFunc("/stat", fileStat)
	mux.HandleFunc("/remote_dir/", getRemoteDirectory)

	wrapper := accessMiddleware(mux)

	log.Fatal(http.ListenAndServe(":8080", wrapper))
}
