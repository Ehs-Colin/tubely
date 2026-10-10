package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}
	// 1> Authentication was already complete
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)
	// TODO: implement the upload here
	// 2> Parse the form data
	// 2.1> Set a const maxMemory to 10MB. I just bit-shifted the number 10 to the left 20 times to get an int that stores the proper number of bytes.
	const maxMemory = 10 << 20
	// 2.2> Use (http.Request).ParseMultipartForm with the maxMemory const as an argument
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse request", err)
		return
	}
	// 3> Get the image data from the platform
	// 3.1> Use r.FormFile to get the imageData data and imageData headers. The Key the web browser is using is called "thumbnail"
	imageData, imageHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't get image data from platform", err)
		return
	}
	defer imageData.Close()
	// 3.2> Get the media type from the form file's Content-Type header
	mediaType, _, err := mime.ParseMediaType(imageHeader.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type", nil)
		return
	}
	if mediaType != "image/jpg" && mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Invalid file type", nil)
		return
	}
	// CH1::L7> Thanks for screwing with my neat comment chains.  Anyway, updated to write to file instead of previous nonsense
	imagePath := getAssetPath(mediaType) //getAssetPath(videoID, mediaType)
	imageDiskPath := cfg.getAssetDiskPath(imagePath)
	f, err := os.Create(imageDiskPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't create file on server", err)
		return
	}
	defer f.Close()
	// copy the file. duh
	if _, err = io.Copy(f, imageData); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error saving file", err)
		return
	}
	// 5> Get the video's metadata from the SQLite database.  The apiConfig's db has a GetVideo method
	dbVideo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't retrieve video metadata", err)
		return
	}
	if dbVideo.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to update this video", nil)
		return
	}

	newUrl := cfg.getAssetURL(imagePath)
	dbVideo.ThumbnailURL = &newUrl
	err = cfg.db.UpdateVideo(dbVideo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video info", err)
		return
	}

	respondWithJSON(w, http.StatusOK, dbVideo)
}
