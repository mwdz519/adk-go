// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// WebPageTool represents a tool that can be used to load a web page.
type WebPageTool struct {
	hc *http.Client
}

func NewWebPageTool(hc *http.Client) *WebPageTool {
	if hc == nil {
		hc = http.DefaultClient
	}

	return &WebPageTool{
		hc: hc,
	}
}

// LoadWebPage fetches the content in the url and returns the text in it.
func (t *WebPageTool) LoadWebPage(ctx context.Context, uri string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, http.NoBody)
	if err != nil {
		return "", nil
	}

	resp, err := t.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var text string
	switch resp.StatusCode {
	case http.StatusOK:
	// TODO(zchee): use github.com/PuerkitoBio/goquery
	// soup = BeautifulSoup(response.content, 'lxml')
	// text = soup.get_text(separator='\n', strip=True)
	default:
		text = fmt.Sprintf("Failed to fetch url: %s", uri)
	}

	// Split the text into lines, filtering out very short lines
	// (e.g., single words or short subtitles)
	if len(text) <= 3 {
		return "", fmt.Errorf("too short text: %s", text)
	}

	// TODO(zchee): use [text/scanner]
	content := strings.Split(text, "\n")
	return strings.Join(content, "\n"), nil
}
