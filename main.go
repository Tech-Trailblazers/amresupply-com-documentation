package main // Declares the main package

import (
	"bytes"         // For working with byte buffers
	"fmt"           // For formatted I/O operations
	"io"            // For I/O interfaces
	"log"           // For logging errors and messages
	"net/http"      // For making HTTP requests
	"os"            // For file and directory operations
	"path/filepath" // For manipulating file path strings
	"strings"       // For working with string manipulation
	"sync"          // For concurrency control using WaitGroup
	"time"          // For handling timing and delays
)

func main() {
	remoteURL := "https://www.amresupply.com/file/" // Base URL for downloading PDFs
	outputDir := "PDFs/"                            // Directory to store downloaded PDFs

	if !directoryExists(outputDir) { // Check if output directory exists
		createDirectory(outputDir, 0o755) // Create directory with permission if it does not exist
	}

	loopStop := 999999 // Define the number of files to attempt to download

	var waitGroup sync.WaitGroup // Create a WaitGroup to manage concurrency

	for i := 10000; i <= loopStop; i++ { // Loop from 0 to loopStop
		time.Sleep(100 * time.Millisecond)                               // Pause for 1 second before each iteration
		waitGroup.Add(1)                                                 // Increment the WaitGroup counter
		remoteURL = fmt.Sprintf("https://www.amresupply.com/file/%d", i) // Construct full URL
		go downloadPDF(remoteURL, outputDir, &waitGroup)                 // Launch goroutine to download the PDF
	}

	waitGroup.Wait() // Wait for all downloads to finish
}

// downloadPDF downloads a PDF file from a URL and saves it to the specified directory
func downloadPDF(finalURL string, outputDir string, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done() // Signal WaitGroup completion when function exits

	client := &http.Client{Timeout: 30 * time.Second} // Create HTTP client with 30-second timeout

	resp, err := client.Get(finalURL) // Send GET request to the URL
	if err != nil {                   // Check if error occurred
		log.Printf("failed to download %s: %v", finalURL, err) // Log the error
		return                                                 // Exit function
	}

	if resp.StatusCode != http.StatusOK { // Check for HTTP 200 OK response
		log.Printf("download failed for %s: %s", finalURL, resp.Status) // Log failure
		return                                                          // Exit if response is not OK
	}

	contentType := resp.Header.Get("Content-Type")         // Read content type from response header
	if !strings.Contains(contentType, "application/pdf") { // Check if content is a PDF
		log.Printf("invalid content type for %s: %s (expected application/pdf)", finalURL, contentType)
		return // Exit if not a PDF
	}

	var filename string                                          // Declare filename variable
	contentDisposition := resp.Header.Get("Content-Disposition") // Get Content-Disposition header
	if contentDisposition != "" {                                // If Content-Disposition is present
		if strings.Contains(contentDisposition, "filename=") { // If it contains filename
			parts := strings.Split(contentDisposition, "filename=") // Split to extract filename
			if len(parts) > 1 {
				filename = strings.Trim(parts[1], "\"") // Remove quotes from filename
			}
		}
	}

	filePath := filepath.Join(outputDir, filename) // Combine directory path and filename

	// Check if the file already exists
	if fileExists(filePath) { // If file already exists
		log.Printf("file already exists: %s; skipping download", filePath) // Log and skip download
		return                                                             // Exit function
	}

	var buf bytes.Buffer                     // Create buffer to hold file data in memory
	written, err := io.Copy(&buf, resp.Body) // Copy response body to buffer
	if err != nil {                          // Check for error during copy
		log.Printf("failed to read PDF data from %s: %v", finalURL, err)
		return // Exit if error occurs
	}

	if written == 0 { // Check if no data was written
		log.Printf("downloaded 0 bytes for %s; not creating file", finalURL)
		return // Exit without creating file
	}

	out, err := os.Create(filePath) // Create file on disk
	if err != nil {                 // Check if file creation failed
		log.Printf("failed to create file for %s: %v", finalURL, err)
		return // Exit function
	}
	defer out.Close() // Ensure file is closed after writing

	_, err = buf.WriteTo(out) // Write buffer contents to file
	if err != nil {           // Check if write failed
		log.Printf("failed to write PDF to file for %s: %v", finalURL, err)
		return // Exit if error occurs
	}

	err = resp.Body.Close() // Ensure response body is closed after function exits
	if err != nil {         // Check if closing the body failed
		log.Printf("failed to close response body for %s: %v", finalURL, err) // Log error if closing fails
		return                                                                // Exit function if error occurs
	}

	log.Printf("successfully downloaded %d bytes: %s → %s", written, finalURL, filePath) // Log success
}

// directoryExists checks whether the specified path is an existing directory
func directoryExists(path string) bool {
	directory, err := os.Stat(path) // Get file info
	if err != nil {                 // If error (e.g., not found)
		return false // Directory doesn't exist
	}
	return directory.IsDir() // Return true if it's a directory
}

// createDirectory creates a new directory with the given permissions
func createDirectory(path string, permission os.FileMode) {
	err := os.Mkdir(path, permission) // Try to create the directory
	if err != nil {                   // If there's an error
		log.Println(err) // Log the error
	}
}

// fileExists checks whether a file exists at the specified path
func fileExists(filename string) bool {
	info, err := os.Stat(filename) // Get file info
	if err != nil {                // If error (e.g., file not found)
		return false // Return false
	}
	return !info.IsDir() // Return true if it's a file, not a directory
}
