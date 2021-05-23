package docparser

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/textract"
)

func TestNew(t *testing.T) {
	client := New()
	if client == nil {
		t.Error("error creating parser client")
	}
}

type mockTextractClient struct {
	textractClientOutput *textract.AnalyzeDocumentOutput
	textractClientError  error
}

func (m *mockTextractClient) AnalyzeDocument(input *textract.AnalyzeDocumentInput) (*textract.AnalyzeDocumentOutput, error) {
	return m.textractClientOutput, m.textractClientError
}

func TestParse(t *testing.T) {
	tests := []struct {
		description          string
		textractClientOutput *textract.AnalyzeDocumentOutput
		textractClientError  error
		content              Content
		error                error
	}{
		{
			description:          "textract client analyze error",
			textractClientOutput: nil,
			textractClientError:  errors.New("mock analyze error"),
			content:              Content{},
			error:                &ErrorAnalyzeDocument{},
		},
		{
			textractClientOutput: &textract.AnalyzeDocumentOutput{},
			description:          "no words/lines returned",
			textractClientError:  nil,
			content: Content{
				ID: "test_id",
			},
			error: nil,
		},
		{
			textractClientOutput: &textract.AnalyzeDocumentOutput{},
			description:          "one word and one line returned",
			textractClientError:  nil,
			content: Content{
				ID: "test_id",
				Words: []Data{
					{
						Text: "testword",
						Coordinates: Coordinates{
							TopLeft: Coordinate{
								X: 0.1,
								Y: 0.1,
							},
							TopRight: Coordinate{
								X: 0.5,
								Y: 0.1,
							},
							BottomLeft: Coordinate{
								X: 0.1,
								Y: 0.3,
							},
							BottomRight: Coordinate{
								X: 0.5,
								Y: 0.3,
							},
						},
					},
				},
				Lines: []Data{
					{
						Text: "test line",
						Coordinates: Coordinates{
							TopLeft: Coordinate{
								X: 0.1,
								Y: 0.3,
							},
							TopRight: Coordinate{
								X: 0.5,
								Y: 0.3,
							},
							BottomLeft: Coordinate{
								X: 0.1,
								Y: 0.5,
							},
							BottomRight: Coordinate{
								X: 0.5,
								Y: 0.5,
							},
						},
					},
				},
			},
			error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				textractClient: &mockTextractClient{
					textractClientOutput: test.textractClientOutput,
					textractClientError:  test.textractClientError,
				},
				convertToContent: func(input *textract.AnalyzeDocumentOutput) Content {
					return test.content
				},
			}

			ctx := context.Background()
			doc := []byte("content")

			output, err := client.Parse(ctx, doc)

			if err != nil {
				switch test.error.(type) {
				case *ErrorAnalyzeDocument:
					var testError *ErrorAnalyzeDocument
					if !errors.As(err, &testError) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, testError)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			}

			if err == nil {
				if output.ID == "" {
					t.Errorf("no content id, received: %+v", output)
				}

				if len(output.Words) != len(test.content.Words) || len(output.Lines) != len(test.content.Lines) {
					t.Errorf("unequal words/lines lengths, received: %+v, expected: %+v", output, test.content)
				} else {
					for i, word := range output.Words {
						if word.Text != output.Words[i].Text {
							t.Errorf("incorrect word text, received: %s, expected: %s", word.Text, output.Words[i].Text)
						}

						if !checkCoordinates(t, word.Coordinates, output.Words[i].Coordinates) {
							t.Errorf("incorrect word coordinates, received: %+v, expected: %+v", word.Coordinates, output.Words[i].Coordinates)
						}
					}

					for i, line := range output.Lines {
						if line.Text != output.Lines[i].Text {
							t.Errorf("incorrect line text, received: %s, expected: %s", line.Text, output.Lines[i].Text)
						}

						if !checkCoordinates(t, line.Coordinates, output.Lines[i].Coordinates) {
							t.Errorf("incorrect line coordinates, received: %+v, expected: %+v", line.Coordinates, output.Lines[i].Coordinates)
						}
					}
				}
			}
		})
	}
}

func Test_convertToContent(t *testing.T) {
	tests := []struct {
		description string
		input       *textract.AnalyzeDocumentOutput
		output      Content
	}{
		{
			description: "one word no lines",
			input: &textract.AnalyzeDocumentOutput{
				Blocks: []*textract.Block{
					{
						BlockType: aws.String(textract.BlockTypeWord),
						Text:      aws.String("testword"),
						Geometry: &textract.Geometry{
							BoundingBox: &textract.BoundingBox{
								Height: aws.Float64(0.2),
								Width:  aws.Float64(0.5),
								Top:    aws.Float64(0.1),
								Left:   aws.Float64(0.1),
							},
						},
					},
				},
			},
			output: Content{
				ID: "test_id",
				Words: []Data{
					{
						Text: "testword",
						Coordinates: Coordinates{
							TopLeft: Coordinate{
								X: 0.1,
								Y: 0.1,
							},
							TopRight: Coordinate{
								X: 0.6,
								Y: 0.1,
							},
							BottomLeft: Coordinate{
								X: 0.1,
								Y: 0.3,
							},
							BottomRight: Coordinate{
								X: 0.6,
								Y: 0.3,
							},
						},
					},
				},
			},
		},
		{
			description: "on words one line",
			input: &textract.AnalyzeDocumentOutput{
				Blocks: []*textract.Block{
					{
						BlockType: aws.String(textract.BlockTypeLine),
						Text:      aws.String("test line"),
						Geometry: &textract.Geometry{
							BoundingBox: &textract.BoundingBox{
								Height: aws.Float64(0.2),
								Width:  aws.Float64(0.5),
								Top:    aws.Float64(0.1),
								Left:   aws.Float64(0.1),
							},
						},
					},
				},
			},
			output: Content{
				ID: "test_id",
				Words: []Data{
					{
						Text: "test line",
						Coordinates: Coordinates{
							TopLeft: Coordinate{
								X: 0.1,
								Y: 0.1,
							},
							TopRight: Coordinate{
								X: 0.6,
								Y: 0.1,
							},
							BottomLeft: Coordinate{
								X: 0.1,
								Y: 0.3,
							},
							BottomRight: Coordinate{
								X: 0.6,
								Y: 0.3,
							},
						},
					},
				},
			},
		},
		{
			description: "one word one line",
			input: &textract.AnalyzeDocumentOutput{
				Blocks: []*textract.Block{
					{
						BlockType: aws.String(textract.BlockTypeWord),
						Text:      aws.String("testword"),
						Geometry: &textract.Geometry{
							BoundingBox: &textract.BoundingBox{
								Height: aws.Float64(0.2),
								Width:  aws.Float64(0.5),
								Top:    aws.Float64(0.1),
								Left:   aws.Float64(0.1),
							},
						},
					},
					{
						BlockType: aws.String(textract.BlockTypeLine),
						Text:      aws.String("test line"),
						Geometry: &textract.Geometry{
							BoundingBox: &textract.BoundingBox{
								Height: aws.Float64(0.2),
								Width:  aws.Float64(0.5),
								Top:    aws.Float64(0.3),
								Left:   aws.Float64(0.1),
							},
						},
					},
				},
			},
			output: Content{
				ID: "test_id",
				Words: []Data{
					{
						Text: "testword",
						Coordinates: Coordinates{
							TopLeft: Coordinate{
								X: 0.1,
								Y: 0.1,
							},
							TopRight: Coordinate{
								X: 0.6,
								Y: 0.1,
							},
							BottomLeft: Coordinate{
								X: 0.1,
								Y: 0.3,
							},
							BottomRight: Coordinate{
								X: 0.6,
								Y: 0.3,
							},
						},
					},
				},
				Lines: []Data{
					{
						Text: "test line",
						Coordinates: Coordinates{
							TopLeft: Coordinate{
								X: 0.3,
								Y: 0.1,
							},
							TopRight: Coordinate{
								X: 0.6,
								Y: 0.3,
							},
							BottomLeft: Coordinate{
								X: 0.1,
								Y: 0.5,
							},
							BottomRight: Coordinate{
								X: 0.6,
								Y: 0.5,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			output := convertToContent(test.input)

			for i, word := range output.Words {
				if word.Text != output.Words[i].Text {
					t.Errorf("incorrect word text, received: %s, expected: %s", word.Text, output.Words[i].Text)
				}

				if !checkCoordinates(t, word.Coordinates, output.Words[i].Coordinates) {
					t.Errorf("incorrect word coordinates, received: %+v, expected: %+v", word.Coordinates, output.Words[i].Coordinates)
				}
			}

			for i, line := range output.Lines {
				if line.Text != output.Lines[i].Text {
					t.Errorf("incorrect line text, received: %s, expected: %s", line.Text, output.Lines[i].Text)
				}

				if !checkCoordinates(t, line.Coordinates, output.Lines[i].Coordinates) {
					t.Errorf("incorrect line coordinates, received: %+v, expected: %+v", line.Coordinates, output.Lines[i].Coordinates)
				}
			}
		})
	}
}

func checkCoordinates(t *testing.T, a, b Coordinates) bool {
	t.Helper()

	if !checkCoordinate(t, a.TopLeft, b.TopLeft) {
		return false
	}

	if !checkCoordinate(t, a.TopRight, b.TopRight) {
		return false
	}

	if !checkCoordinate(t, a.BottomLeft, b.BottomLeft) {
		return false
	}

	if !checkCoordinate(t, a.BottomRight, b.BottomRight) {
		return false
	}

	return true
}

func checkCoordinate(t *testing.T, a, b Coordinate) bool {
	t.Helper()

	if a.X != b.X || a.Y != b.Y {
		return false
	}

	return true
}
