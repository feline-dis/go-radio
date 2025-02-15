package utils

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseYouTubeVideoID extracts the video ID from a YouTube URL.
func ParseYouTubeVideoID(youtubeURL string) (string, error) {
	parsedURL, err := url.Parse(youtubeURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Check if the host is YouTube or youtu.be
	switch parsedURL.Host {
	case "www.youtube.com", "youtube.com":
		// Extract video ID from query parameters
		queryParams := parsedURL.Query()
		if videoID, exists := queryParams["v"]; exists && len(videoID) > 0 {
			return videoID[0], nil
		}
	case "youtu.be":
		// Extract video ID from the path
		pathParts := strings.Trim(parsedURL.Path, "/")
		if pathParts != "" {
			return pathParts, nil
		}
	}

	return "", fmt.Errorf("video ID not found in URL")
}
