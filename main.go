package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	storage_go "github.com/supabase-community/storage-go"
	"github.com/supabase-community/supabase-go"
)

type Messgae struct {
	FullPath string `json:"fullPath"`
	ID       string `json:"id"`
	Path     string `json:"path"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("FAILED TO LOAD ENV FILE:[ERROR]:%s", err)
	}
	API_URL := os.Getenv("API_URL")
	API_KEY := os.Getenv("API_KEY")
	QUEUE_BASE_URL := os.Getenv("QUEUE_BASE_URL")

	client, err := supabase.NewClient(API_URL, API_KEY, &supabase.ClientOptions{})
	if err != nil {
		fmt.Println("cannot initialize client", err)
	}

	res, err := http.Get(QUEUE_BASE_URL + "/get-message")
	if err != nil {
		log.Printf("FAILED TO GET MESSAGE FROM QUEUE:[ERROR]:%s", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("FAILED TO read the body:[ERROR]:%s", err)
	}

	var msg Messgae
	err = json.Unmarshal(body, &msg)
	if err != nil {
		log.Printf("FAILED TO UNMARSHAL JSON:[ERROR]:%s", err)
	}

	// Download the file
	fileByte, err := client.Storage.DownloadFile("temp", msg.Path, storage_go.UrlOptions{})
	if err != nil {
		log.Printf("FAILED TO DOWNLOAD THE FILE:[ERROR]:%s", err)
	}

	// Create file
	dirPath := "./videos"                     // Change to your desired path
	outputFilePath := dirPath + "/output.mp4" // Change the extension as needed

	// Ensure the videos directory exists
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		log.Fatalf("FAILED TO CREATE DIRECTORY:[ERROR]:%s", err)
	}

	// Create output file in the videos directory
	file, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Write the byte slice to the file
	_, err = file.Write(fileByte)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("MP4 file created successfully:", outputFilePath)
}
