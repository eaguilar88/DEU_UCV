package gotenberg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const (
	convertHTMLPath = "/forms/chromium/convert/html"
	requestTimeout  = 60 * time.Second
	// maxErrorBody caps how much of an error response is read into the returned error.
	maxErrorBody = 1024
)

// Client renders HTML documents to PDF through a Gotenberg server.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: requestTimeout},
	}
}

// Render converts a self-contained HTML document (styles and images inlined) into a landscape
// letter-size PDF with no margins, so the document's own layout fills the page.
func (c *Client) Render(ctx context.Context, html []byte) ([]byte, error) {
	var body bytes.Buffer
	form := multipart.NewWriter(&body)

	part, err := form.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(html); err != nil {
		return nil, err
	}
	fields := map[string]string{
		"paperWidth":      "11",
		"paperHeight":     "8.5",
		"marginTop":       "0",
		"marginBottom":    "0",
		"marginLeft":      "0",
		"marginRight":     "0",
		"printBackground": "true",
	}
	for name, value := range fields {
		if err := form.WriteField(name, value); err != nil {
			return nil, err
		}
	}
	if err := form.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+convertHTMLPath, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", form.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling gotenberg: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		if err != nil {
			return nil, fmt.Errorf("gotenberg returned %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("gotenberg returned %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	pdf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading gotenberg response: %w", err)
	}
	return pdf, nil
}
