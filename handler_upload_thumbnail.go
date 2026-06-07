package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

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
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid media type", err)
		return
	}

	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusUnsupportedMediaType, "unsupported media type. Please use png or jpeg", fmt.Errorf("%s media type is not supported", mediaType))
		return
	}

	videoMetaData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "video not found", err)
		return
	}

	if videoMetaData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "not video owner", err)
		return
	}

	extensions, err := mime.ExtensionsByType(mediaType)
	if err != nil || len(extensions) == 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid or unknown Content-Type", err)
		return
	}

	filename := videoIDString + extensions[0]
	filePath := filepath.Join(cfg.assetsRoot, filename)
	newFile, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to create file", err)
		return
	}
	_, err = io.Copy(newFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to save file", err)
		return
	}

	url := fmt.Sprint("http://localhost:", cfg.port, "/assets/", filename)
	videoMetaData.ThumbnailURL = &url
	err = cfg.db.UpdateVideo(videoMetaData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to save new video data", err)
	}

	respondWithJSON(w, http.StatusOK, videoMetaData)
}
