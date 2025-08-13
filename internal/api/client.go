package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Bookmark struct {
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Section     string   `json:"section,omitempty"`
	Hash        string   `json:"hash,omitempty"`
}

type Usage struct {
	Hash       string `json:"hash"`
	UserId     string `json:"user_id"`
	Usage      int    `json:"usage"`
	TotalUsage int    `json:"total_usage"`
}

type apiResponse struct {
	Items []Bookmark `json:"items"`
	Total int        `json:"total"`
}

type UsageRequest struct {
	Hash   string `json:"hash"`
	UserId string `json:"user_id"`
}

var httpClient = &http.Client{Timeout: 12 * time.Second}

func UseBookmark(base string, q url.Values) (Usage, error) {

	endpoint := base + "/api/usage"

	reqBody := UsageRequest{
		Hash:   q.Get("hash"),
		UserId: q.Get("user_id"),
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return Usage{}, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return Usage{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return Usage{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return Usage{}, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return Usage{}, err
	}

	var usage Usage
	if err := json.Unmarshal(b, &usage); err != nil {
		return Usage{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return usage, nil
}

func FetchBookmarks(base string, q url.Values) ([]Bookmark, error) {
	endpoint := base + "/api/bookmarks"
	if qs := q.Encode(); qs != "" {
		endpoint += "?" + qs
	}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// attempt decode flexible formats (either a raw array or an object with items field)
	var arr []Bookmark
	if err := json.Unmarshal(b, &arr); err == nil {
		// Successfully decoded slice (could be empty)
		if arr == nil { // normalize nil slice to empty
			arr = []Bookmark{}
		}
		return arr, nil
	}
	var ar apiResponse
	if err := json.Unmarshal(b, &ar); err == nil {
		if ar.Items == nil {
			return []Bookmark{}, nil
		}
		return ar.Items, nil
	}
	return nil, errors.New("unexpected response format")
}

// SearchAI performs an AI-powered natural language search using /api/ai/search?q=...
// It returns bookmarks similar in structure to FetchBookmarks.
func SearchAI(base, query string) ([]Bookmark, error) {
	// Use PathEscape so spaces become %20 (some backends are picky about '+' for spaces)
	endpoint := base + "/api/ai/search?q=" + url.PathEscape(query)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var arr []Bookmark
	if err := json.Unmarshal(b, &arr); err == nil {
		if arr == nil {
			arr = []Bookmark{}
		}
		return arr, nil
	}
	var ar apiResponse
	if err := json.Unmarshal(b, &ar); err == nil {
		if ar.Items == nil {
			return []Bookmark{}, nil
		}
		return ar.Items, nil
	}
	return nil, errors.New("unexpected response format")
}
