package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	// 1> Set an upload limit of 1GB
	const uploadLimit = 1 << 30
	r.Body = http.MaxBytesReader(w, r.Body, uploadLimit)
	// 2> Extract the videoID from the URL
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}
	// 3> Authenticate the user to get a userID
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}
	// 4> Get video data from the database and verify it belongs to the user
	dbVideo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't find video", err)
		return
	}
	if dbVideo.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to update this video", nil)
	}
	// 5> Parse the uploaded video file from the form data
	videoData, videoHeader, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't get video data from platform", err)
		return
	}
	defer videoData.Close()
	// 6> Validate the uploaded file to ensure it's an MP4 video
	mediaType, _, err := mime.ParseMediaType(videoHeader.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type", nil)
		return
	}
	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Invalid file type, only MP4 is allowed", nil)
		return
	}
	// 7> Save the uploaded file to a tempiorary file on the disk
	tempFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't create file on server", err)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	if _, err = io.Copy(tempFile, videoData); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error saving file", err)
		return
	}
	// 8> Reset tempFile's pointer to the beginning with Seek
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error resetting file", err)
		return
	}
	aspectRatio, err := getVideoAspectRatio(tempFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error finding aspect ratio", err)
		return
	}
	// 9> Put the object into S3 using PutObject
	key := aspectRatio + "/" + getAssetPath(mediaType)
	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      aws.String(cfg.s3Bucket),
		Key:         aws.String(key),
		Body:        tempFile,
		ContentType: aws.String(mediaType),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error uploading file to S3", err)
		return
	}
	// 10> Update VideoURL in database with S3 bucket and key
	newUrl := cfg.getObjectURL(key)
	dbVideo.VideoURL = &newUrl
	err = cfg.db.UpdateVideo(dbVideo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, dbVideo)
}

func getVideoAspectRatio(filePath string) (string, error) {
	var buffer bytes.Buffer
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	cmd.Stdout = &buffer
	cmd.Run()

	videoFileJSON := FfprobeFile{}
	err := json.Unmarshal(buffer.Bytes(), &videoFileJSON)
	if err != nil {
		return "", err
	}

	height := videoFileJSON.Streams[0].Height
	width := videoFileJSON.Streams[0].Width
	if height == 16*width/9 {
		return "portrait", nil
	} else if width == 16*height/9 {
		return "landscape", nil
	}
	return "other", nil
}
