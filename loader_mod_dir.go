package main

import (
	"bufio"
	"log"
	"net/http"
	"time"
)

// @bried структура списка файлов в удаленной директории
// Url - удаленная директория
// Files - список файлов с датой модификации
// LastModified - последнее время обновления страницы удаленной директории
type DirectoryIndex struct {
	Url          string
	Files        FullDirectoryIndexList
	LastModified string
}

// @brief структура loader`а списка файлов в удаленной директории
type FileListLoader_ModDir struct {
	Index DirectoryIndex
}

// @brief метод обновления индексного списка удаленной директории
// Запрашивает индекс с заголовком If-Modified-Since
// Проверяет на наличие заголовка Last-Modified
func (m *FileListLoader_ModDir) ReloadIndexList() {
	// Перечитываем директорию
	req, err := http.NewRequest("GET", m.Index.Url, nil)
	if err != nil {
		log.Println("[Loader_ModDir] error on creating new request: ", err)
		return
	}

	if len(m.Index.LastModified) > 0 {
		req.Header.Add("If-Modified-Since", m.Index.LastModified)
	}

	//TODO добавить таймаут при запросе к удаленной директории
	client := http.Client{}
	r, err := client.Do(req)
	if err != nil {
		log.Println("[Loader_ModDir] error on sending request: ", err)
		return
	}
	defer r.Body.Close()

	if r.StatusCode == http.StatusNotModified {
		log.Println("[Loader_ModDir] directory was not modified")
		return
	}

	lastModified := r.Header.Get("Last-Modified")
	if len(lastModified) > 0 {
		m.Index.LastModified = lastModified
		log.Println("[Loader_ModDir] Last modified: ", m.Index.LastModified)
	}

	reader := bufio.NewReader(r.Body)
	filelist := ParseAutoIndex(*reader)

	m.Index.Files = filelist
}

// @brief раз в секунду обновляет список файлов по урлу m.Url
func (m *FileListLoader_ModDir) reload(in <-chan int) {
	ticker := time.NewTicker(1 * time.Second).C
	for {
		select {
		case <-ticker:
			m.ReloadIndexList()
		case <-in:
			return
		}
	}
}
