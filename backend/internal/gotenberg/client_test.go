package gotenberg

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Render(t *testing.T) {
	var gotHTML, gotWidth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, convertHTMLPath, r.URL.Path)
		file, header, err := r.FormFile("files")
		require.NoError(t, err)
		assert.Equal(t, "index.html", header.Filename)
		data, _ := io.ReadAll(file)
		gotHTML = string(data)
		gotWidth = r.FormValue("paperWidth")
		_, _ = w.Write([]byte("%PDF-1.7"))
	}))
	defer server.Close()

	pdf, err := NewClient(server.URL+"/").Render(context.Background(), []byte("<html>hola</html>"))

	require.NoError(t, err)
	assert.Equal(t, "%PDF-1.7", string(pdf))
	assert.Equal(t, "<html>hola</html>", gotHTML)
	assert.Equal(t, "11", gotWidth)
}

func TestClient_Render_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "chromium crashed", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := NewClient(server.URL).Render(context.Background(), []byte("<html></html>"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "503")
	assert.Contains(t, err.Error(), "chromium crashed")
}
