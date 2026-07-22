package system_setting

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
)

const (
	defaultCanvasURL               = "http://localhost:3001"
	defaultCanvasTokenGroup        = "Image"
	defaultCanvasHandoffTTLSeconds = 60
	minCanvasHandoffTTLSeconds     = 15
	maxCanvasHandoffTTLSeconds     = 300
)

func GetCanvasURL() (string, error) {
	rawURL := strings.TrimSpace(os.Getenv("CANVAS_URL"))
	if rawURL == "" {
		rawURL = defaultCanvasURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid CANVAS_URL: %w", err)
	}
	if (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" || parsedURL.User != nil || parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return "", fmt.Errorf("CANVAS_URL must be an http(s) URL without credentials, query parameters, or a fragment")
	}

	return strings.TrimRight(parsedURL.String(), "/"), nil
}

func GetCanvasOrigin() (string, error) {
	canvasURL, err := GetCanvasURL()
	if err != nil {
		return "", err
	}
	parsedURL, err := url.Parse(canvasURL)
	if err != nil {
		return "", err
	}
	return common.NormalizeOrigin(parsedURL.Scheme + "://" + parsedURL.Host)
}

func GetCanvasAPIBaseURL() (string, error) {
	rawURL := strings.TrimSpace(os.Getenv("NEW_API_PUBLIC_URL"))
	if rawURL == "" {
		rawURL = strings.TrimSpace(ServerAddress)
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid New API public URL: %w", err)
	}
	if (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" || parsedURL.User != nil || parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return "", fmt.Errorf("New API public URL must be an http(s) URL without credentials, query parameters, or a fragment")
	}
	return strings.TrimRight(parsedURL.String(), "/"), nil
}

func GetCanvasTokenGroup() string {
	group := strings.TrimSpace(os.Getenv("CANVAS_TOKEN_GROUP"))
	if group == "" {
		return defaultCanvasTokenGroup
	}
	return group
}

func GetCanvasHandoffTTL() time.Duration {
	seconds := defaultCanvasHandoffTTLSeconds
	if rawTTL := strings.TrimSpace(os.Getenv("CANVAS_HANDOFF_TTL_SECONDS")); rawTTL != "" {
		if parsedTTL, err := strconv.Atoi(rawTTL); err == nil {
			seconds = parsedTTL
		}
	}
	if seconds < minCanvasHandoffTTLSeconds {
		seconds = minCanvasHandoffTTLSeconds
	}
	if seconds > maxCanvasHandoffTTLSeconds {
		seconds = maxCanvasHandoffTTLSeconds
	}
	return time.Duration(seconds) * time.Second
}
