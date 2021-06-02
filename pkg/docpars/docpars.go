package docpars

import "context"

// Document holds the output of parsing the provided image file.
type Document struct {
	ID       string `json:"id"`
	Filename string `json:"filename"` // NOTE: maybe include/exclude
	Filepath string `json:"filepath"` // NOTE: maybe include/exclude
	Pages    []Page `json:"pages"`
}

// Page holds the output of parsing the pages of the provided image file.
type Page struct {
	ID         string `json:"id"`
	DocumentID string `json:"document_id"` // NOTE: maybe include/exclude
	PageNumber int64  `json:"page_number"`
	Lines      []Data `json:"lines"`
}

// Data holds text and location coordinates retrieved from the image file.
type Data struct {
	ID          string      `json:"id"`
	DocumentID  string      `json:"document_id"` // NOTE: maybe include/exclude
	PageNumber  int64       `json:"page_number"` // NOTE: maybe include/exclude
	Text        string      `json:"text"`
	Coordinates Coordinates `json:"coordinates"`
}

// Coordinates holds the four coordinate points for a piece of text.
type Coordinates struct {
	TopLeft     Point `json:"top_left"`
	TopRight    Point `json:"top_right"`
	BottomLeft  Point `json:"bottom_left"`
	BottomRight Point `json:"bottom_right"`
}

// Point holds the X and Y values for a point in text coordinates.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Parser defines the method needed for converting the provided
// doc image into database content.
type Parser interface {
	Parse(ctx context.Context, doc []byte) (*Document, error)
}
