package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	storage_go "github.com/supabase-community/storage-go"
	"github.com/supabase-community/supabase-go"
)

type Message struct {
	FullPath string `json:"fullPath"`
	ID       string `json:"id"`
	Path     string `json:"path"`
}

type UtilConfig struct {
	API_URL        string
	API_KEY        string
	Client         *supabase.Client
	QUEUE_BASE_URL string
}

func NewUtil() *UtilConfig {
	// load env
	err := godotenv.Load(".ENVIRONMENT_VARIABLE")
	if err != nil {
		err = godotenv.Load("ENVIRONMENT_VARIABLE")
	}
	QUEUE_BASE_URL := os.Getenv("QUEUE_BASE_URL")
	API_URL := os.Getenv("API_URL")
	API_KEY := os.Getenv("API_KEY")
	// create supabase client
	client, err := supabase.NewClient(API_URL, API_KEY, &supabase.ClientOptions{})
	if err != nil {
		return nil
	}

	return &UtilConfig{
		API_URL:        API_URL,
		API_KEY:        API_KEY,
		Client:         client,
		QUEUE_BASE_URL: QUEUE_BASE_URL,
	}
}

func (util *UtilConfig) FetchMessage() (Message, error) {
	res, err := http.Get(util.QUEUE_BASE_URL + "/get-message")
	if res.Status == "204 No Content" {
		return Message{}, errors.New(res.Status)
	}
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

func (util *UtilConfig) DownloadFIle(msg Message) ([]byte, error) {

	fileBytes, err := util.Client.Storage.DownloadFile("temp", msg.Path, storage_go.UrlOptions{})
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

func (util *UtilConfig) SaveVideoInDir(fileBytes []byte) (string, error) {

	videoPath := "./videos"                   // Change to your desired path
	inputFilePath := videoPath + "/input.mp4" // Change the extension as needed

	// Ensure the videos directory exists
	if err := os.MkdirAll(videoPath, os.ModePerm); err != nil {
		log.Fatalf("FAILED TO CREATE DIRECTORY:[ERROR]:%s", err)
	}

	// Create input file in the videos directory
	inputfile, err := os.Create(inputFilePath)
	if err != nil {
		log.Println("Error creating file:", err)
		return "", err
	}
	defer inputfile.Close()

	// Write the byte slice to the file
	_, err = inputfile.Write(fileBytes)
	if err != nil {
		log.Println("Error writing to file:", err)
		return "", err
	}

	log.Println("MP4 file created successfully:", inputFilePath)
	return "", nil
}

// splitVideoIntoSegments splits the video into segments using ffmpeg
// NOTE:- ffmpeg -i ../videos/input.mp4 -f segment -segment_time 10 -c copy ../output/output%03d.ts
func (util *UtilConfig) SplitVideoIntoSegments(input string) error {

	// Get the absolute path of the input and output
	inputPath, err := filepath.Abs("./videos/input.mp4")
	if err != nil {
		return err
	}
	outputFolder, err := filepath.Abs("./output")
	if err != nil {
		return err
	}
	outputPath, err := filepath.Abs("./output/output%03d.ts")
	if err != nil {
		return err
	}

	err = os.MkdirAll(outputFolder, 0755)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("ffmpeg", "-i", inputPath, "-f", "segment", "-segment_time", "10", "-c", "copy", outputPath)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("Command output: %s\n", out.String())
		log.Printf("Command error: %s\n", stderr.String())
		return err
	}
	log.Println(out.String())
	return nil
}

func (util *UtilConfig) UploadSegmentsToSupabase(id string) ([]string, error) {
	var segmentFiles []string

	// Read the segment files created by ffmpeg
	files, err := filepath.Glob("output/output*.ts")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		// Read the file content
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		// Upload the file to Supabase Storage
		filePath := id + strings.Split(file, "/")[1]
		_, err = util.Client.Storage.UploadFile("transcoded-videos", filePath, bytes.NewReader(data), storage_go.FileOptions{})
		if err != nil {
			return nil, err
		}
		res := util.Client.Storage.GetPublicUrl("transcoded-videos", filePath, storage_go.UrlOptions{})
		segmentFiles = append(segmentFiles, res.SignedURL)

	}
	return segmentFiles, nil
}

// createM3U8File creates an M3U8 file with the given public URLs
func (util *UtilConfig) CreateM3U8File(segmentFiles []string, id string) error {

	indexFolderPath, err := filepath.Abs("./index")
	if err != nil {
		return err
	}

	filePath := fmt.Sprintf("./index/%s.m3u8", id)
	indexFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(indexFolderPath, 0755)
	if err != nil {
		log.Fatal(err)
	}
	m3u8Content := "#EXTM3U\n#EXT-X-VERSION:3\n"
	for _, url := range segmentFiles {
		m3u8Content += fmt.Sprintf("#EXTINF:%d,\n%s\n", 10, url)
	}
	m3u8Content += "#EXT-X-ENDLIST\n"

	// Write the M3U8 content to a file
	if err := os.WriteFile(indexFilePath, []byte(m3u8Content), 0644); err != nil {
		return err
	}

	return nil
}

func (util *UtilConfig) UploadM3U8ToSupabase() error {

	// Read the segment files created by ffmpeg
	files, err := filepath.Glob("index/*.m3u8")
	if err != nil {
		return err
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil
		}
		//upload to supabase
		supabasePath := strings.Split(file, "/")[1]
		_, err = util.Client.Storage.UploadFile("m3u8_index", supabasePath, bytes.NewReader(data), storage_go.FileOptions{})
		// get public url
		res := util.Client.Storage.GetPublicUrl("m3u8_index", supabasePath, storage_go.UrlOptions{})
		log.Println(res.SignedURL)
	}

	return nil
}

func (util *UtilConfig) CleanUP() {
	folderPaths := []string{"./index", "./output", "./videos"}

	for _, folder := range folderPaths {
		folderPath, err := filepath.Abs(folder)
		if err != nil {
			log.Printf("FAILED TO DELETE FOLDER:[ERROR] %s\n", err)
		}
		err = os.RemoveAll(folderPath)
		if err != nil {
			fmt.Println("Error removing folder:", err)
			return
		}

	}

	fmt.Println("Folder deleted successfully.")
}
