package main // Declares the main package

import (
	"bufio"         // For reading files line by line
	"bytes"         // For working with byte buffers
	"io"            // For I/O interfaces
	"log"           // For logging errors and messages
	"net/http"      // For making HTTP requests
	"os"            // For file and directory operations
	"path/filepath" // For manipulating file path strings
	"regexp"        // For regular expression matching
	// "slices"        // For working with slices (arrays)
	"strings" // For working with string manipulation
	"sync"    // For concurrency control using WaitGroup
	"time"    // For handling timing and delays
)

func main() {
	remoteURLFile := "urls.txt"                             // File containing URLs to download
	remoteHTMLFileContent := "amresupply.html"              // File to store HTML content
	remoteURLContent := readAppendLineByLine(remoteURLFile) // Read the file content

	outputDir := "PDFs/" // Directory to store downloaded PDFs

	if !directoryExists(outputDir) { // Check if output directory exists
		createDirectory(outputDir, 0755) // Create directory with permission if it does not exist
	}

	var waitGroup sync.WaitGroup // Create a WaitGroup to manage concurrency

	// Reverse the slice.
	// slices.Reverse(remoteURLContent) // Reverse the order of URLs in the slice

	for _, urls := range remoteURLContent { // Loop from 0 to loopStop
		// Get the data from the URL and append it to the remoteHTMLFileContent
		appendByteToFile(remoteHTMLFileContent, getDataFromURL(urls)) // Append the data from the URL to the HTML file
		// Extract all matching URLs from the HTML content
		amreSupplyURLs := extractAmreSupplyURLs(readAFileAsString(remoteHTMLFileContent)) // Extract URLs from the HTML content
		// Remove duplicates from the extracted URLs
		amreSupplyURLs = removeDuplicatesFromSlice(amreSupplyURLs) // Remove duplicates
		for _, remoteURL := range amreSupplyURLs {                 // Loop through each extracted URL
			// time.Sleep(100 * time.Millisecond)               // Pause for 100 milliseconds before each download
			waitGroup.Add(1)                                 // Increment the WaitGroup counter for each URL
			go downloadPDF(remoteURL, outputDir, &waitGroup) // Launch a goroutine to download the PDF
		}
	}
	waitGroup.Wait() // Wait for all downloads to finish
}

// Append and write to file
func appendAndWriteToFile(path string, content string) {
	filePath, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	_, err = filePath.WriteString(content + "\n")
	if err != nil {
		log.Println(err)
	}
	err = filePath.Close()
	if err != nil {
		log.Println(err)
	}
}

// Remove all the duplicates from a slice and return the slice.
func removeDuplicatesFromSlice(slice []string) []string {
	check := make(map[string]bool)
	var newReturnSlice []string
	for _, content := range slice {
		if !check[content] {
			check[content] = true
			newReturnSlice = append(newReturnSlice, content)
		}
	}
	return newReturnSlice
}

// Read and append the file line by line to a slice.
func readAppendLineByLine(path string) []string {
	var returnSlice []string
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		returnSlice = append(returnSlice, scanner.Text())
	}
	err = file.Close()
	if err != nil {
		log.Println(err)
	}
	return returnSlice
}

// AppendToFile appends the given byte slice to the specified file.
// If the file doesn't exist, it will be created.
func appendByteToFile(filename string, data []byte) {
	// Open the file with appropriate flags and permissions
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("error opening file %s: %v", filename, err) // Log error if file cannot be opened
		return
	}
	defer file.Close()

	// Write data to the file
	_, err = file.Write(data)
	if err != nil {
		log.Printf("error writing to file %s: %v", filename, err) //
	}
}

// Send a http get request to a given url and return the data from that url.
func getDataFromURL(uri string) []byte {
	response, err := http.Get(uri)
	if err != nil {
		log.Println(err)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
	}
	err = response.Body.Close()
	if err != nil {
		log.Println(err)
	}
	return body
}

// Read a file and return the contents
func readAFileAsString(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
	}
	return string(content)
}

// extractAmreSupplyURLs takes a string and returns all matching URLs of the form:
// https://www.amresupply.com/file/{number}/
func extractAmreSupplyURLs(input string) []string {
	// Regular expression to match the target URL pattern
	regex := `https:\/\/www\.amresupply\.com\/file\/\d+\/`
	// Compile the regular expression
	re := regexp.MustCompile(regex)
	// Find all matching URLs
	matches := re.FindAllString(input, -1)
	// Return the list of matched URLs
	return matches
}

// fileContainsString checks if the file at filePath contains the search string.
// Logs any errors and returns false in case of an error.
func fileContainsString(filePath string, search string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file: %v\n", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), search) {
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading file: %v\n", err)
		return false
	}

	return false
}

// downloadPDF downloads a PDF file from a URL and saves it to the specified directory
func downloadPDF(finalURL string, outputDir string, waitGroup *sync.WaitGroup) {
	// The location to the local file where to save the url of the file already downloaded
	localURLSaveFileLocation := "already_downloaded_urls.txt"
	// Check if the given file contains the URL already
	if fileContainsString(localURLSaveFileLocation, finalURL) { // If the URL is already downloaded
		log.Printf("URL already downloaded: %s; skipping download", finalURL)
		return // Exit the function if URL is already downloaded
	}

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
				filename = strings.ToLower(strings.Trim(parts[1], "\"")) // Remove quotes from filename
			}
		}
	}

	filePath := filepath.Join(outputDir, filename) // Combine directory path and filename

	// Check if the file already exists
	if fileExists(filePath) { // If file already exists
		log.Printf("file already exists: %s; skipping download", filePath) // Log and skip download
		// Append the successfully downloaded URL to the already downloaded URLs file
		if !fileContainsString(localURLSaveFileLocation, finalURL) { // If URL is not already in the file
			appendAndWriteToFile(localURLSaveFileLocation, finalURL) // Append the URL
		}
		return // Exit function
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
	if !fileContainsString(localURLSaveFileLocation, finalURL) { // If URL is not already in the file
		// Append the successfully downloaded URL to the already downloaded URLs file
		appendAndWriteToFile(localURLSaveFileLocation, finalURL) // Append the URL to the file
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
