package docpars

import "context"

// Content holds the output of parsing the provided image file.
type Content struct {
	ID    string `json:"id"`
	Lines []Data `json:"lines"`
	Words []Data `json:"words"`
}

// Data holds text and location coordinates retrieved from the image file.
type Data struct {
	Text        string      `json:"text"`
	Coordinates Coordinates `json:"coordinates"`
}

// Coordinates holds the four coordinate points for a piece of text.
type Coordinates struct {
	TopLeft     Coordinate `json:"top_left"`
	TopRight    Coordinate `json:"top_right"`
	BottomLeft  Coordinate `json:"bottom_left"`
	BottomRight Coordinate `json:"bottom_right"`
}

// Coordinate holds the X and Y values for a point in text coordinates.
type Coordinate struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Parser defines the method needed for converting the provided
// doc image into database content.
type Parser interface {
	Parse(ctx context.Context, doc []byte) (*Content, error)
}
