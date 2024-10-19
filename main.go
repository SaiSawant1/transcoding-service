package main

import (
	"log"

	"github.com/SaiSawant1/transcoder/utils"
)

func main() {

	// fetch message form queue
	u := utils.NewUtil()
	msg, err := u.FetchMessage()

	if err != nil {
		log.Printf("FAILED TO FETCH MESSAGE FROM QUEUE:[ERROR]:%s\n", err)
		return
	}

	// download the file
	fileBytes, err := u.DownloadFIle(msg)
	if err != nil {
		log.Printf("FAILED TO DOWNLOAD FILE FROM STORAGE:[ERROR]:%s\n", err)
		return
	}

	//create input file from the downloaded bytes

	filePath, err := u.SaveVideoInDir(fileBytes)
	if err != nil {
		log.Printf("FAILED TO SAVE VIDEO LOCALY:[ERROR]:%s\n", err)
		return
	}

	err = u.SplitVideoIntoSegments(filePath)
	if err != nil {
		log.Printf("FAILED TO TRANSCODE VIDEO:[ERROR]:%s\n", err)
		return
	}

	segments, err := u.UploadSegmentsToSupabase(msg.ID)
	if err != nil {
		log.Printf("FAILED TO UPLOAD VIDEO SEGMENTS:[ERROR]:%s\n", err)
		return
	}
	err = u.CreateM3U8File(segments, msg.ID)
	if err != nil {
		log.Printf("FAILED TO UPLOAD VIDEO SEGMENTS:[ERROR]:%s\n", err)
		return
	}

	err = u.UploadM3U8ToSupabase()
	if err != nil {
		log.Printf("FAILED TO UPLOAD VIDEO SEGMENTS:[ERROR]:%s\n", err)
		return
	}

	u.CleanUP()

}
