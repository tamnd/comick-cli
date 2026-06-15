// Package comick is the library behind the comick command line:
// the HTTP client, request shaping, and the typed data models for comick.io.
package comick

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Comic is the record emitted for search, trending, new, and comic-detail commands.
type Comic struct {
	HID           string `json:"hid"`
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	Country       string `json:"country"`
	Status        int    `json:"status"`
	StatusText    string `json:"status_text"`
	ContentRating string `json:"content_rating"`
	LastChapter   string `json:"last_chapter"`
	Demographic   string `json:"demographic"`
	Type          string `json:"type"`
	Genres        string `json:"genres"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	URL           string `json:"url"`
}

// Chapter is the record emitted for the chapters command.
type Chapter struct {
	HID       string `json:"hid"`
	Chap      string `json:"chap"`
	Vol       string `json:"vol"`
	Lang      string `json:"lang"`
	Groups    string `json:"groups"`
	UpCount   int    `json:"up_count"`
	DownCount int    `json:"down_count"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	URL       string `json:"url"`
}

// ─── wire types (unexported, JSON decode only) ────────────────────────────────

// wireGenreList decodes the genres field, which the API returns as either
// a plain string array ["Action"] or an object array [{"name":"Action"}].
type wireGenreList []string

func (g *wireGenreList) UnmarshalJSON(b []byte) error {
	// try plain string array first
	var ss []string
	if err := json.Unmarshal(b, &ss); err == nil {
		*g = ss
		return nil
	}
	// fall back to object array {name: ...}
	var objs []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(b, &objs); err != nil {
		return err
	}
	*g = make(wireGenreList, len(objs))
	for i, o := range objs {
		(*g)[i] = o.Name
	}
	return nil
}

type wireComic struct {
	HID           string        `json:"hid"`
	Slug          string        `json:"slug"`
	Title         string        `json:"title"`
	Country       string        `json:"country"`
	Status        int           `json:"status"`
	ContentRating string        `json:"content_rating"`
	LastChapter   string        `json:"last_chapter"`
	Demographic   string        `json:"demographic"`
	Type          string        `json:"type"`
	Genres        wireGenreList `json:"genres"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type wireComicDetail struct {
	Comic    wireComic    `json:"comic"`
	Authors  []wireAuthor `json:"authors"`
	Artists  []wireArtist `json:"artists"`
	Genres   []wireGenre  `json:"genres"`
	LangList []string     `json:"langList"`
}

type wireAuthor struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type wireArtist struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type wireGenre struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type wireChapter struct {
	HID       string    `json:"hid"`
	Chap      string    `json:"chap"`
	Vol       string    `json:"vol"`
	Lang      string    `json:"lang"`
	GroupName []string  `json:"group_name"`
	UpCount   int       `json:"up_count"`
	DownCount int       `json:"down_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type wireChaptersResp struct {
	Chapters []wireChapter `json:"chapters"`
	Total    int           `json:"total"`
}

// ─── mapping helpers ──────────────────────────────────────────────────────────

func statusText(s int) string {
	switch s {
	case 1:
		return "ongoing"
	case 2:
		return "completed"
	case 3:
		return "cancelled"
	case 4:
		return "hiatus"
	default:
		return "unknown"
	}
}

func comicURL(slug string) string {
	return fmt.Sprintf("https://comick.io/comic/%s", slug)
}

func chapterURL(hid string) string {
	return fmt.Sprintf("https://comick.io/chapter/%s", hid)
}

func wireComicToComic(w wireComic) Comic {
	genres := strings.Join(w.Genres, ";")
	return Comic{
		HID:           w.HID,
		Slug:          w.Slug,
		Title:         w.Title,
		Country:       w.Country,
		Status:        w.Status,
		StatusText:    statusText(w.Status),
		ContentRating: w.ContentRating,
		LastChapter:   w.LastChapter,
		Demographic:   w.Demographic,
		Type:          w.Type,
		Genres:        genres,
		CreatedAt:     formatTime(w.CreatedAt),
		UpdatedAt:     formatTime(w.UpdatedAt),
		URL:           comicURL(w.Slug),
	}
}

func wireComicDetailToComic(d wireComicDetail) Comic {
	c := wireComicToComic(d.Comic)
	// prefer top-level genres array when populated
	if len(d.Genres) > 0 {
		names := make([]string, len(d.Genres))
		for i, g := range d.Genres {
			names[i] = g.Name
		}
		c.Genres = strings.Join(names, ";")
	}
	return c
}

func wireChapterToChapter(w wireChapter) Chapter {
	return Chapter{
		HID:       w.HID,
		Chap:      w.Chap,
		Vol:       w.Vol,
		Lang:      w.Lang,
		Groups:    strings.Join(w.GroupName, ";"),
		UpCount:   w.UpCount,
		DownCount: w.DownCount,
		CreatedAt: formatTime(w.CreatedAt),
		UpdatedAt: formatTime(w.UpdatedAt),
		URL:       chapterURL(w.HID),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
