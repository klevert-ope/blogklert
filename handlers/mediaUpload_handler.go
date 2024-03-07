package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/gorilla/mux"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

type uploadResponse struct {
	URL string `json:"url"`
}

func SetupUploadRoute(router *mux.Router, storageAccount, cdnEndpoint, containerName string) {
	router.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		UploadFileHandler(w, r, storageAccount, cdnEndpoint, containerName)
	}).Methods("POST")
}

func UploadFileHandler(w http.ResponseWriter, r *http.Request, storageAccount, cdnEndpoint, containerName string) {
	// Parse the file from the request
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to parse file")
		return
	}
	defer closeFileWithLog(file)

	// Check file size
	const maxFileSize = 100 << 20 // 100 MB
	if header.Size > maxFileSize {
		writeError(w, http.StatusBadRequest, "File size exceeds maximum allowed size")
		return
	}

	// Validate file extension
	ext := getFileExtension(header.Filename)
	if !isValidFileExtension(ext) {
		writeError(w, http.StatusBadRequest, "Invalid file extension")
		return
	}

	// Create a client for the specified storage account
	client, err := azblob.NewClientFromConnectionString(storageAccount, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create storage client")
		return
	}

	// Generate a random token for the blob name
	token := generateToken(4) // Adjust the token length as needed
	extension := getFileExtension(header.Filename)
	blobName := token + extension

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "upload-")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create temporary file")
		return
	}
	defer removeTempFileWithLog(tmpFile.Name())

	// Write the uploaded file to the temporary file
	_, err = io.Copy(tmpFile, file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to copy file to temporary storage")
		return
	}

	// Close the temporary file before uploading it to Azure Blob Storage
	if err := tmpFile.Close(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to close temporary file")
		return
	}

	// Open the temporary file for reading
	tmpFile, err = os.Open(tmpFile.Name())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to open temporary file")
		return
	}

	// Upload the temporary file to Azure Blob Storage
	_, err = client.UploadFile(context.TODO(), containerName, blobName, tmpFile, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload file to Azure Blob Storage: %v", err))
		return
	}

	defer func() {
		// Close the temporary file after the upload is complete
		if err := tmpFile.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	// Construct the CDN URL for the uploaded file
	cdnURL := fmt.Sprintf("%s/%s/%s", cdnEndpoint, containerName, blobName)

	// Respond with the upload URL in JSON format
	response := uploadResponse{URL: cdnURL}
	w.Header().Set("Content-Type", "application/json")

	// Encode the response to JSON and handle potential errors
	if err := json.NewEncoder(w).Encode(response); err != nil {
		handleError(w, http.StatusInternalServerError, "Failed to encode response as JSON", err)
		return
	}
}

func isValidFileExtension(ext string) bool {
	// Define a whitelist of allowed file extensions
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".tiff": true,
		".webp": true,
		".svg":  true,
		".mp4":  true,
		".mov":  true,
		".avi":  true,
		".wmv":  true,
		".mkv":  true,
	}
	return allowedExtensions[ext]
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	_, err := fmt.Fprint(w, message)
	if err != nil {
		log.Printf("Failed to write error response: %v", err)
	}
}

func closeFileWithLog(file multipart.File) {
	if err := file.Close(); err != nil {
		log.Printf("Error closing file: %v", err)
	}
}

func removeTempFileWithLog(name string) {
	if err := os.Remove(name); err != nil {
		log.Printf("Failed to remove temporary file: %v", err)
	}
}

func handleError(w http.ResponseWriter, code int, message string, err error) {
	log.Printf("%s: %v", message, err)
	http.Error(w, message, code)
}

func generateToken(length int) string {
	const charset = "0123456789"
	token := make([]byte, length)
	for i := range token {
		token[i] = charset[rand.Intn(len(charset))]
	}
	return "MD_" + string(token)
}

func getFileExtension(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return strings.ToLower(filename[i:])
		}
	}
	return ""
}
