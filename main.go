package main

import (
	"encoding/csv"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func main() {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}
	defaultDownloadDir := path.Join(homeDir, "Downloads", strconv.Itoa(int(time.Now().Unix())))

	filePath := flag.String("file", "data.csv", "Path to the CSV file")
	comma := flag.String("comma", ",", "Delimiter used in the CSV file")
	urlField := flag.String("urlField", "商品图片", "Field number containing the URL")
	fileField := flag.String("fileField", "商品名称", "Field number containing the file name")
	downloadDir := flag.String("download", defaultDownloadDir, "Path to the download directory")
	flag.Parse()

	file, err := os.Open(*filePath)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = rune((*comma)[0])
	reader.FieldsPerRecord = -1

	// 读取文件头
	header, err := reader.Read()
	if err != nil {
		log.Fatalf("Error reading header: %v", err)
	}
	urlIndex := -1
	fileIndex := -1

	for i, field := range header {
		if field == *urlField {
			urlIndex = i
		}
		if field == *fileField {
			fileIndex = i
		}
	}

	if urlIndex == -1 {
		log.Fatalf("URL field not found: %v", *urlField)
	}
	if fileIndex == -1 {
		log.Fatalf("File field not found: %v", *fileField)
	}

	if _, e := os.Stat(*downloadDir); os.IsNotExist(e) {
		ce := os.MkdirAll(*downloadDir, os.ModePerm)
		if ce != nil {
			log.Fatalf("Error creating download directory: %v", ce)
		}
	}

	for {
		record, e := reader.Read()
		if err == io.EOF {
			break
		}
		if e != nil {
			break
		}
		log.Printf("Downloading file: %s => %s", record[urlIndex], record[fileIndex])
		fileExt := path.Ext(record[urlIndex])
		resp, e := http.DefaultClient.Get(record[urlIndex])
		if e != nil {
			log.Fatalf("Error downloading file: %v", e)
		}
		defer resp.Body.Close()
		filename := strings.ReplaceAll(record[fileIndex], " ", "_")
		filename = strings.ReplaceAll(filename, "/", "_")
		dataPath := path.Join(*downloadDir, filename+fileExt)
		filePtr, e := os.Create(dataPath)
		if e != nil {
			log.Fatalf("Error creating file: %v", e)
		}
		defer filePtr.Close()
		_, e = io.Copy(filePtr, resp.Body)
		if e != nil {
			log.Fatalf("Error writing file: %v", e)
		}
	}

	log.Println(*downloadDir)
}
