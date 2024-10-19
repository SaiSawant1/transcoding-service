package main

import (
	"log"

	"github.com/SaiSawant1/transcoder/utils"
)

func main() {

	// fetch message form queue
	msg, err := utils.FetchMessage()

	if err != nil {
		log.Printf("FAILED TO FETCH MESSAGE FROM QUEUE:[ERROR]:%s", err)
	}

	// download the file
	fileBytes, err := utils.DownloadFIle(msg)
	if err != nil {
		log.Printf("FAILED TO DOWNLOAD FILE FROM STORAGE:[ERROR]:%s", err)
		return
	}

	//create input file from the downloaded bytes

	filePath, err := utils.SaveVideoInDir(fileBytes)
	if err != nil {
		log.Printf("FAILED TO SAVE VIDEO LOCALY:[ERROR]:%s", err)
		return
	}

	err = utils.SplitVideoIntoSegments(filePath)
	if err != nil {
		log.Printf("FAILED TO TRANSCODE VIDEO:[ERROR]:%s", err)
		return
	}

}
