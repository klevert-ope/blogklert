package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

type uploadResponse struct {
	URL string `json:"url"`
}

func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the file from the request
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to parse file")
		return
	}
	defer closeFileWithLog(file)

	storageAccount := os.Getenv("STORAGE_ACCOUNT_ENDPOINT")
	if storageAccount == "" {
		writeError(w, http.StatusInternalServerError,
			"Storage Account endpoint environment variable (STORAGE_ACCOUNT_ENDPOINT) is not set")
		return
	}

	// Load the container name from environment variable
	containerName := os.Getenv("CONTAINER_NAME")
	if containerName == "" {
		writeError(w, http.StatusInternalServerError,
			"Container Name environment variable (CONTAINER_NAME) is not set")
		return
	}

	// Load the CDN endpoint URL from environment variable
	cdnEndpoint := os.Getenv("CDN_ENDPOINT_URL")
	if cdnEndpoint == "" {
		writeError(w, http.StatusInternalServerError,
			"CDN Endpoint URL environment variable (CDN_ENDPOINT_URL) is not set")
		return
	}

	// Create a client for the specified storage account
	client, err := azblob.NewClientFromConnectionString(storageAccount, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create storage client")
		return
	}

	// Generate a unique name for the blob
	blobName := header.Filename

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
