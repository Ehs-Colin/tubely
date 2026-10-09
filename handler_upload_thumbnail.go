package main

import (
	"fmt"
	"io"
	"net/http"

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
	// 3.2> Get the media type from the form file's Content-Type header
	mediaType := imageHeader.Header.Get("Content-Type")
	// 4> Read all the image data into a byte slice using io.ReadAll
	videoData, err := io.ReadAll(imageData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't read image data", err)
		return
	}
	// 5> Get the video's metadata from the SQLite database.  The apiConfig's db has a GetVideo method
	dbVideo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could't retrieve video metadata", err)
		return
	}
	// 6> SDave the thumbnail to the global map
	// 6.1> Create a new thumbnail struct with the image data and media type
	videoThumbnail := thumbnail{
		data:      videoData,
		mediaType: mediaType,
	}
	// 6.2> Add the thumbnail to the global map, using the video's ID as the key
	videoThumbnails[videoID] = videoThumbnail
	// 7> Update the video metadata so that it has a new thumbnail URL, then update the record in the database by using the cfg.db.UpdateVideo function
	// 	  The thumbnail URL should have this format: http://localhost:<port>/api/thumbnails/{videoID}
	newUrl := fmt.Sprintf("http://localhost:%v/api/thumbnails/%v", cfg.port, videoIDString)
	dbVideo.ThumbnailURL = &newUrl
	cfg.db.UpdateVideo(dbVideo)

	respondWithJSON(w, http.StatusOK, dbVideo)
}
