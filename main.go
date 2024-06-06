package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/render"
)

func uploadFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	// Parse the form data
	err := r.ParseMultipartForm(1000 << 20) // 10 MB limit
	if err != nil {
		log.Println("Error 1: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the file from the request
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println("Error 2: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a new file on the server
	newFile, err := os.Create("./uploads/" + handler.Filename)
	if err != nil {
		log.Println("Error 3: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer newFile.Close()

	// Copy the file content to the new file
	_, err = io.Copy(newFile, file)
	if err != nil {
		log.Println("Error 4: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	fmt.Fprint(w, "File uploaded successfully")
}

func getVideo(w http.ResponseWriter, r *http.Request) {
	// Allow CORS from any origin
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract video file name from URL parameters
	videoFileName := r.URL.Query().Get("filename")
	if videoFileName == "" {
		http.Error(w, "Video filename not provided", http.StatusBadRequest)
		return
	}

	// Open the video file
	videoFile, err := os.Open("./uploads/" + videoFileName + ".mp4")
	if err != nil {
		fmt.Println("Error opening video file:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer videoFile.Close()

	// Get file information
	fileInfo, err := videoFile.Stat()
	if err != nil {
		fmt.Println("Error getting file information:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Length", fmt.Sprint(fileInfo.Size()))
	w.Header().Set("Accept-Ranges", "bytes")

	// Copy the video content to the response writer
	_, err = io.Copy(w, videoFile)
	if err != nil {
		fmt.Println("Error serving video:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	res := map[string]interface{}{
		"data": "done",
	}

	render.JSON(w, r, res)
}

func main() {
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/video", getVideo)
	http.ListenAndServe(":8080", nil)
}
