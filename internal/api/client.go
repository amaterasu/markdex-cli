package api

import (
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
}

type apiResponse struct {
	Items []Bookmark `json:"items"`
	Total int        `json:"total"`
}

var httpClient = &http.Client{Timeout: 12 * time.Second}

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
	// attempt decode flexible formats
	var arr []Bookmark
	if err := json.Unmarshal(b, &arr); err == nil && len(arr) > 0 {
		return arr, nil
	}
	var ar apiResponse
	if err := json.Unmarshal(b, &ar); err == nil && len(ar.Items) > 0 {
		return ar.Items, nil
	}
	// empty acceptable
	if string(b) == "[]" {
		return []Bookmark{}, nil
	}
	return nil, errors.New("unexpected response format")
}
