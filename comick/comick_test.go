package comick_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/comick-cli/comick"
)

func newTestClient(t *testing.T, mux *http.ServeMux) *comick.Client {
	t.Helper()
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	cfg := comick.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return comick.NewClient(cfg)
}

func TestSearch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/search/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"hid":"abc123","slug":"one-piece","title":"One Piece","country":"jp","status":1,"content_rating":"safe","last_chapter":"1110","created_at":"2021-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z","genres":["Action","Adventure"]},
			{"hid":"def456","slug":"naruto","title":"Naruto","country":"jp","status":2,"content_rating":"safe","last_chapter":"700","created_at":"2020-01-01T00:00:00Z","updated_at":"2022-01-01T00:00:00Z","genres":["Action"]}
		]`))
	})
	c := newTestClient(t, mux)
	comics, err := c.Search(context.Background(), "one piece", 20, "view", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(comics) != 2 {
		t.Fatalf("len = %d, want 2", len(comics))
	}
	if comics[0].Title != "One Piece" {
		t.Errorf("title = %q, want %q", comics[0].Title, "One Piece")
	}
	if comics[0].HID != "abc123" {
		t.Errorf("hid = %q, want %q", comics[0].HID, "abc123")
	}
	if comics[0].StatusText != "ongoing" {
		t.Errorf("status_text = %q, want %q", comics[0].StatusText, "ongoing")
	}
	if comics[1].StatusText != "completed" {
		t.Errorf("status_text = %q, want %q", comics[1].StatusText, "completed")
	}
}

func TestTrending(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/search/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("sort") != "trending" {
			t.Errorf("sort = %q, want %q", r.URL.Query().Get("sort"), "trending")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"hid":"t1","slug":"trending-manga","title":"Trending Manga","country":"kr","status":1,"content_rating":"safe","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-06-01T00:00:00Z"},
			{"hid":"t2","slug":"trending-manga-2","title":"Trending Manga 2","country":"jp","status":1,"content_rating":"safe","created_at":"2024-02-01T00:00:00Z","updated_at":"2024-06-01T00:00:00Z"}
		]`))
	})
	c := newTestClient(t, mux)
	comics, err := c.Trending(context.Background(), 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(comics) != 2 {
		t.Fatalf("len = %d, want 2", len(comics))
	}
	if comics[0].HID != "t1" {
		t.Errorf("hid = %q, want %q", comics[0].HID, "t1")
	}
}

func TestNew(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1.0/search/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("sort") != "new" {
			t.Errorf("sort = %q, want %q", r.URL.Query().Get("sort"), "new")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"hid":"n1","slug":"newest-comic","title":"Newest Comic","country":"us","status":1,"content_rating":"safe","created_at":"2024-06-01T00:00:00Z","updated_at":"2024-06-01T00:00:00Z"}
		]`))
	})
	c := newTestClient(t, mux)
	comics, err := c.New(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(comics) != 1 {
		t.Fatalf("len = %d, want 1", len(comics))
	}
	if comics[0].Title != "Newest Comic" {
		t.Errorf("title = %q, want %q", comics[0].Title, "Newest Comic")
	}
}

func TestGetComic(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/comic/abc123/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"comic": {
				"hid":"abc123","slug":"one-piece","title":"One Piece","country":"jp",
				"status":1,"content_rating":"safe","last_chapter":"1110","demographic":"shounen",
				"created_at":"2021-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"
			},
			"genres": [
				{"name":"Action","slug":"action"},
				{"name":"Adventure","slug":"adventure"}
			]
		}`))
	})
	c := newTestClient(t, mux)
	comic, err := c.GetComic(context.Background(), "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if comic.Title != "One Piece" {
		t.Errorf("title = %q, want %q", comic.Title, "One Piece")
	}
	if comic.Demographic != "shounen" {
		t.Errorf("demographic = %q, want %q", comic.Demographic, "shounen")
	}
	if comic.Genres != "Action;Adventure" {
		t.Errorf("genres = %q, want %q", comic.Genres, "Action;Adventure")
	}
	if comic.StatusText != "ongoing" {
		t.Errorf("status_text = %q, want %q", comic.StatusText, "ongoing")
	}
}

func TestGetChapters(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/comic/abc123/chapters", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("lang") != "en" {
			t.Errorf("lang = %q, want %q", r.URL.Query().Get("lang"), "en")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"chapters": [
				{"hid":"ch1","chap":"1110","vol":"","lang":"en","group_name":["MangaPlus"],"up_count":100,"down_count":2,"created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"},
				{"hid":"ch2","chap":"1109","vol":"","lang":"en","group_name":["MangaPlus"],"up_count":95,"down_count":1,"created_at":"2023-12-01T00:00:00Z","updated_at":"2023-12-01T00:00:00Z"}
			],
			"total": 1110
		}`))
	})
	c := newTestClient(t, mux)
	chapters, err := c.GetChapters(context.Background(), "abc123", 2, 1, "en", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(chapters) != 2 {
		t.Fatalf("len = %d, want 2", len(chapters))
	}
	if chapters[0].Chap != "1110" {
		t.Errorf("chap = %q, want %q", chapters[0].Chap, "1110")
	}
	if chapters[0].Groups != "MangaPlus" {
		t.Errorf("groups = %q, want %q", chapters[0].Groups, "MangaPlus")
	}
}

func TestGetComicNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/comic/nonexistent/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"statusCode":404,"message":"Not found"}`))
	})
	c := newTestClient(t, mux)
	_, err := c.GetComic(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, comick.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
