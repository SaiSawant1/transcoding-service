package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
	storage_go "github.com/supabase-community/storage-go"
	"github.com/supabase-community/supabase-go"
)

type Message struct {
	FullPath string `json:"fullPath"`
	ID       string `json:"id"`
	Path     string `json:"path"`
}

func FetchMessage() (Message, error) {
	err := godotenv.Load()
	if err != nil {
		return Message{}, err
	}
	QUEUE_BASE_URL := os.Getenv("QUEUE_BASE_URL")

	res, err := http.Get(QUEUE_BASE_URL + "/get-message")
	if err != nil {
		return Message{}, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Message{}, err
	}

	var msg Message
	err = json.Unmarshal(body, &msg)
	if err != nil {
		return Message{}, err
	}

	return msg, nil

}

func DownloadFIle(msg Message) ([]byte, error) {

	err := godotenv.Load()
	if err != nil {
		log.Printf("FAILED TO LOAD ENV FILE:[ERROR]:%s", err)
	}
	API_URL := os.Getenv("API_URL")
	API_KEY := os.Getenv("API_KEY")

	client, err := supabase.NewClient(API_URL, API_KEY, &supabase.ClientOptions{})
	if err != nil {
		return nil, err
	}

	fileBytes, err := client.Storage.DownloadFile("temp", msg.Path, storage_go.UrlOptions{})
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

func SaveVideoInDir(fileBytes []byte) (string, error) {

	dirPath := "./videos"                    // Change to your desired path
	outputFilePath := dirPath + "/input.mp4" // Change the extension as needed

	// Ensure the videos directory exists
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		log.Fatalf("FAILED TO CREATE DIRECTORY:[ERROR]:%s", err)
	}

	// Create output file in the videos directory
	file, err := os.Create(outputFilePath)
	if err != nil {
		log.Println("Error creating file:", err)
		return "", err
	}
	defer file.Close()

	// Write the byte slice to the file
	_, err = file.Write(fileBytes)
	if err != nil {
		log.Println("Error writing to file:", err)
		return "", err
	}

	log.Println("MP4 file created successfully:", outputFilePath)
	return "", nil
}

// splitVideoIntoSegments splits the video into segments using ffmpeg
func SplitVideoIntoSegments(input string) error {
	cmd := exec.Command("ffmpeg", "-i", input, "-f", "segment", "-segment_time", fmt.Sprintf("%d", 10), "-c", "copy", "output%03d.ts")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return err
	}
	log.Println(out.String())
	return nil
}
