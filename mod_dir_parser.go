package main

import (
	"bufio"
	"bytes"
	"io"
)

// @brief строчка в индексном файле директории
type DirectoryIndexEntry struct {
	Name      string
	Timestamp uint32
	Length    uint32
}

// @brief слайс индекс-файл
type FullDirectoryIndexList []DirectoryIndexEntry

const (
	prefix string = "<a href=\""
	a_end  string = "</a>"
)

/*
 * <a href="20120503062942.dat.gz">20120503062942.dat.gz</a>                              03-May-2012 02:31           146869497
 * <a href="20120503063142.dat.gz">20120503063142.dat.gz</a>                              03-May-2012 02:33           149908508
 */
// @brief разбирает индексный файл и возращает результат разбора
func ParseAutoIndex(reader bufio.Reader) FullDirectoryIndexList {
	var result FullDirectoryIndexList
	for {
		buf, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}

		name_start := bytes.Index(buf, []byte(prefix))
		if name_start == -1 {
			continue
		}
		name_start += len(prefix)

		name_end := bytes.Index(buf[name_start:], []byte{'"'})
		if name_end == -1 {
			continue
		}

		result = append(result, DirectoryIndexEntry{
			Name: string(buf[name_start : name_start+name_end]),
		})
		//TODO разбирать время
	}

	return result
}
