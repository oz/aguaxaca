package collector

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestTwitterAPICollector(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/twitter/user/last_tweets":
			if r.Header.Get("X-API-Key") != "test-key" {
				t.Error("missing api key header")
			}
			fmt.Fprintf(w, `{"status":"success","data":{"tweets":[
				{"id":"1","text":"HoyLlegaElAgua mañana","extendedEntities":{"media":[{"type":"photo","media_url_https":%q}]}},
				{"id":"2","text":"no hashtag","extendedEntities":{"media":[{"type":"photo","media_url_https":%q}]}},
				{"id":"3","text":"HoyLlegaElAgua video","extendedEntities":{"media":[{"type":"video","media_url_https":%q}]}}
			]},"has_next_page":false,"next_cursor":""}`,
				srv.URL+"/media/AAA.jpg", srv.URL+"/media/BBB.jpg", srv.URL+"/media/CCC.mp4")
		case "/media/AAA.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write([]byte("fake-image-bytes"))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	c := &TwitterAPICollector{
		Account: "SOAPA_Oax", APIKey: "test-key",
		DownloadDir: dir, BaseURL: srv.URL,
		Log: slog.Default(), HTTPClient: srv.Client(),
	}
	files, err := c.DownloadImages()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		fmt.Printf("file: %s (base %s)\n", f, filepath.Base(f))
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if want := filepath.Join(dir, "media%2FAAA.jpg"); files[0] != want {
		t.Fatalf("got %q, want %q", files[0], want)
	}
	b, _ := os.ReadFile(files[0])
	if string(b) != "fake-image-bytes" {
		t.Fatalf("unexpected content %q", b)
	}
}
