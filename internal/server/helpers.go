package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ikhsan3adi/gemini-web2api/internal/format"
	"github.com/ikhsan3adi/gemini-web2api/internal/multimodal"
)

func randHex(n int) string {
	b := make([]byte, (n+1)/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(data)
}

func startSSE(w http.ResponseWriter) bool {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if ok {
		flusher.Flush()
	}
	return ok
}

func writeSSEData(w http.ResponseWriter, data any) error {
	enc, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "data: %s\n\n", string(enc))
	if err != nil {
		return err
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

func writeSSEEvent(w http.ResponseWriter, event string, data any) error {
	enc, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(enc))
	if err != nil {
		return err
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

func writeSSEDone(w http.ResponseWriter) error {
	_, err := fmt.Fprintf(w, "data: [DONE]\n\n")
	if err != nil {
		return err
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

func (a *App) uploadImages(images []format.Image) []string {
	if len(images) == 0 {
		return nil
	}

	tokens := a.Tokens.Get()
	var fileRefs []string

	httpClient := a.HTTPClient
	if httpClient == nil {
		httpClient = createHTTPClient(a.Cfg)
	}

	for _, img := range images {
		data := img.Data
		if len(data) == 0 {
			continue
		}

		ref, err := multimodal.UploadImage(httpClient, tokens, data, img.MIME, a.Gem.Cookies, a.Cfg.AuthUser)
		if err != nil {
			a.Logf("Image upload failed: %v", err)
			continue
		}
		if ref != "" {
			fileRefs = append(fileRefs, ref)
		}
	}

	if len(fileRefs) == 0 {
		return nil
	}
	return fileRefs
}
