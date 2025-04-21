package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Directory where images are stored
// const imageDir = "/root/nfa_files/"
const imageDir = "/var/nfa/uploads/"

//const imageDir = "/Users/riteshrai/Documents/GitHub/nfa/images/"

func ServeNFAFile(c *gin.Context) {
	// Get the file name from the query parameter
	fileName := c.Query("file") // Use ?file=filename in the URL
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file parameter is required"})
		return
	}

	// Secure file path to prevent directory traversal attacks
	cleanFileName := filepath.Clean(fileName)
	if cleanFileName != fileName || strings.Contains(cleanFileName, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file path"})
		return
	}

	// Get absolute image directory path
	absoluteImageDir, err := filepath.Abs(imageDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	// Construct full file path
	filePath := filepath.Join(absoluteImageDir, cleanFileName)

	// Ensure the requested file is within the allowed directory
	if !strings.HasPrefix(filePath, absoluteImageDir+string(os.PathSeparator)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Check if the file exists and is not a directory
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) || info.IsDir() {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found", "path": filePath})
		return
	}

	// Open file to detect MIME type
	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	defer file.Close()

	// Read part of the file to determine its MIME type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	// Detect content type and set response header
	contentType := http.DetectContentType(buffer)
	c.Writer.Header().Set("Content-Type", contentType)

	// Serve the file
	c.File(filePath)
}
func UploadFiles(c *gin.Context) {
	// Get the list of uploaded files
	file, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error retrieving the files",
		})
		return
	}

	files := file.File["file"] // Get files under the key "image"

	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No files uploaded",
		})
		return
	}

	//Validate and ensure the directory exists
	if err := os.MkdirAll(imageDir, 0755); err != nil {
		log.Println("Error creating directory:", err) // Log the error to your server logs
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Unable to create directory",
			"details": err.Error(), // This will display the error details
		})
		return
	}

	// Prepare to store the uploaded file info
	var uploadedFiles []map[string]string

	// Process each file
	for _, file := range files {
		// Ensure the file name is sanitized
		filename := filepath.Base(file.Filename)
		if filename == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Invalid file name for %s", file.Filename),
			})
			return
		}

		// Create a unique file name
		uniqueName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), filename)
		dstPath := filepath.Join(imageDir, uniqueName)

		// Open the uploaded file (file is a *multipart.FileHeader)
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Unable to open uploaded file",
				"details": err.Error(),
			})
			return
		}
		defer src.Close()

		// Create the destination file
		dst, err := os.Create(dstPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Unable to create file",
				"details": err.Error(),
			})
			return
		}
		defer dst.Close()

		// Copy the file content from src (multipart.File) to the destination file
		if _, err := io.Copy(dst, src); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Unable to save file %s", file.Filename),
			})
			return
		}

		// Generate a unique URL for the uploaded file

		// Store the file information in the response
		uploadedFiles = append(uploadedFiles, map[string]string{
			"file_name": uniqueName,
		})
	}

	// Success response with all uploaded file information
	c.JSON(http.StatusOK, gin.H{
		"message": "Files uploaded successfully",
		"files":   uploadedFiles,
	})
}
