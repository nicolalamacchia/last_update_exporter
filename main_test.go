package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestProbeHandler(t *testing.T) {
	// Create a temporary directory for testing.
	tempDir, err := ioutil.TempDir("", "exporter_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after the test.

	// Create a temporary file to simulate a folder update.
	tempFile, err := ioutil.TempFile(tempDir, "testfile_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.WriteString("test content")
	tempFile.Close()

	// Wait a moment to ensure a noticeable time difference.
	time.Sleep(1 * time.Second)

	// Simulate a request to the /probe endpoint with the temp directory as the target.
	req, err := http.NewRequest("GET", "/probe?target="+tempDir, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(probeHandler)
	handler.ServeHTTP(rr, req)

	// Check the status code.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the body for the expected content.
	// This part is tricky since the exact output depends on the metric's value,
	// which is the time since the temp file was created.
	// Instead, look for the presence of the metric's name.
	responseBody := rr.Body.String()
	if !strings.Contains(responseBody, "folder_last_update_seconds") {
		t.Errorf("Handler returned unexpected body: %v", responseBody)
	}
}

func TestGetMostRecentFileModTime(t *testing.T) {
	// Similar setup to TestProbeHandler.
	tempDir, err := ioutil.TempDir("", "exporter_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary file.
	_, err = ioutil.TempFile(tempDir, "testfile_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	mostRecent, err := getMostRecentFileModTime(tempDir)
	if err != nil {
		t.Fatalf("getMostRecentFileModTime returned an error: %v", err)
	}

	if time.Since(mostRecent) > time.Second {
		t.Errorf("The most recent modification time is too far in the past: %v", mostRecent)
	}
}
