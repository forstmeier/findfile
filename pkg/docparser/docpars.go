package docparser

import "context"

// Content holds the output of parsing the provided image file.
type Content struct {
	ID    string
	Lines []Data
	Words []Data
}

// Data holds text and location coordinates retrieved from the image file.
type Data struct {
	Text        string
	Coordinates Coordinates
}

// Coordinates holds the four coordinate points for a piece of text.
type Coordinates struct {
	TopLeft     Coordinate
	TopRight    Coordinate
	BottomLeft  Coordinate
	BottomRight Coordinate
}

// Coordinate holds the X and Y values for a point in text coordinates.
type Coordinate struct {
	X float64
	Y float64
}

// Parser defines the method needed for converting the provided
// doc image into database content.
type Parser interface {
	Parse(ctx context.Context, doc []byte) (*Content, error)
}
