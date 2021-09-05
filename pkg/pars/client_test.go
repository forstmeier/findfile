package pars

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
)

func TestNew(t *testing.T) {
	client := New(session.New())
	if client == nil {
		t.Error("error creating parser client")
	}
}

type mockTextractClient struct {
	textractClientOutput *textract.DetectDocumentTextOutput
	textractClientError  error
}

func (m *mockTextractClient) DetectDocumentText(input *textract.DetectDocumentTextInput) (*textract.DetectDocumentTextOutput, error) {
	return m.textractClientOutput, m.textractClientError
}

func TestParse(t *testing.T) {
	accountID := "account_id"
	filename := "test.pdf"
	filepath := "s3://bucket/path"

	tests := []struct {
		description          string
		textractClientOutput *textract.DetectDocumentTextOutput
		textractClientError  error
		document             Document
		error                error
	}{
		{
			description:          "textract client analyze error",
			textractClientOutput: nil,
			textractClientError:  errors.New("mock analyze error"),
			document:             Document{},
			error:                &ErrorAnalyzeDocument{},
		},
		{
			textractClientOutput: &textract.DetectDocumentTextOutput{},
			description:          "no pages/lines returned",
			textractClientError:  nil,
			document: Document{
				ID: "document_0",
			},
			error: nil,
		},
		{
			textractClientOutput: &textract.DetectDocumentTextOutput{},
			description:          "one page and one line returned",
			textractClientError:  nil,
			document: Document{
				ID:        "document_0",
				AccountID: accountID,
				Filename:  filename,
				Filepath:  filepath,
				Pages: []Page{
					{
						Lines: []Line{
							{
								Text: "test line",
								Coordinates: Coordinates{
									TopLeft: Point{
										X: 0.1,
										Y: 0.3,
									},
									TopRight: Point{
										X: 0.5,
										Y: 0.3,
									},
									BottomLeft: Point{
										X: 0.1,
										Y: 0.5,
									},
									BottomRight: Point{
										X: 0.5,
										Y: 0.5,
									},
								},
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
				convertToDocument: func(input *textract.DetectDocumentTextOutput, accountID, filename, filepath string) Document {
					test.document.AccountID = accountID
					test.document.Filename = filename
					test.document.Filepath = filepath
					return test.document
				},
			}

			ctx := context.Background()

			document, err := client.Parse(ctx, accountID, filename, filepath, nil)

			if err != nil {
				switch e := test.error.(type) {
				case *ErrorAnalyzeDocument:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				if document.ID == "" {
					t.Errorf("no document id, received: %+v", document)
				}

				if len(document.Pages) != len(test.document.Pages) {
					t.Errorf("unequal pages lengths, received: %+v, expected: %+v", document.Pages, test.document.Pages)
				} else {

					for i, receivedPage := range document.Pages {
						expectedPage := test.document.Pages[i]
						if len(receivedPage.Lines) != len(expectedPage.Lines) {
							t.Errorf("unequal lines lengths, received: %+v, expected: %+v", receivedPage.Lines, expectedPage.Lines)
						} else {
							for j, receivedLine := range receivedPage.Lines {
								expectedLine := expectedPage.Lines[j]
								if expectedLine.Text != receivedLine.Text {
									t.Errorf("incorrect line text, received: %s, expected: %s", receivedLine.Text, expectedLine.Text)
								}
								if !checkCoordinates(t, receivedLine.Coordinates, expectedLine.Coordinates) {
									t.Errorf("incorrect line coordinates, received: %+v, expected: %+v", receivedLine.Coordinates, expectedLine.Coordinates)
								}
							}
						}
					}
				}
			}
		})
	}
}

func Test_convertToContent(t *testing.T) {
	accountID := "account_id"
	filename := "test.pdf"
	filepath := "s3://bucket/path"

	tests := []struct {
		description string
		input       *textract.DetectDocumentTextOutput
		document    Document
	}{
		{
			description: "one page and one line",
			input: &textract.DetectDocumentTextOutput{
				Blocks: []*textract.Block{
					{
						Id:        aws.String("page_0"),
						BlockType: aws.String(textract.BlockTypePage),
						Relationships: []*textract.Relationship{
							{
								Ids: []*string{
									aws.String("line_0"),
								},
							},
						},
					},
					{
						Id:        aws.String("line_0"),
						BlockType: aws.String(textract.BlockTypeLine),
						Text:      aws.String("test words"),
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
			document: Document{
				AccountID: accountID,
				Filename:  filename,
				Filepath:  filepath,
				Pages: []Page{
					{
						PageNumber: 1,
						Lines: []Line{
							{
								Text: "test words",
								Coordinates: Coordinates{
									TopLeft: Point{
										X: 0.1,
										Y: 0.1,
									},
									TopRight: Point{
										X: 0.6,
										Y: 0.1,
									},
									BottomLeft: Point{
										X: 0.1,
										Y: 0.3,
									},
									BottomRight: Point{
										X: 0.6,
										Y: 0.3,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			description: "one page and two lines",
			input: &textract.DetectDocumentTextOutput{
				Blocks: []*textract.Block{
					{
						Id:        aws.String("page_0"),
						BlockType: aws.String(textract.BlockTypePage),
						Relationships: []*textract.Relationship{
							{
								Ids: []*string{
									aws.String("line_0"),
								},
							},
						},
					},
					{
						Id:        aws.String("line_0"),
						BlockType: aws.String(textract.BlockTypeLine),
						Text:      aws.String("test words"),
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
						Id:        aws.String("line_1"),
						BlockType: aws.String(textract.BlockTypeLine),
						Text:      aws.String("another"),
						Geometry: &textract.Geometry{
							BoundingBox: &textract.BoundingBox{
								Height: aws.Float64(0.2),
								Width:  aws.Float64(0.3),
								Top:    aws.Float64(0.1),
								Left:   aws.Float64(0.6),
							},
						},
					},
				},
			},
			document: Document{
				AccountID: accountID,
				Filename:  filename,
				Filepath:  filepath,
				Pages: []Page{
					{
						PageNumber: 1,
						Lines: []Line{
							{
								Text: "test words",
								Coordinates: Coordinates{
									TopLeft: Point{
										X: 0.1,
										Y: 0.1,
									},
									TopRight: Point{
										X: 0.6,
										Y: 0.1,
									},
									BottomLeft: Point{
										X: 0.1,
										Y: 0.3,
									},
									BottomRight: Point{
										X: 0.6,
										Y: 0.3,
									},
								},
							},
							{
								Text: "another",
								Coordinates: Coordinates{
									TopLeft: Point{
										X: 0.6,
										Y: 0.1,
									},
									TopRight: Point{
										X: 0.9,
										Y: 0.1,
									},
									BottomLeft: Point{
										X: 0.6,
										Y: 0.3,
									},
									BottomRight: Point{
										X: 0.9,
										Y: 0.3,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			description: "two pages and two lines",
			input: &textract.DetectDocumentTextOutput{
				Blocks: []*textract.Block{
					{
						Id:        aws.String("page_0"),
						BlockType: aws.String(textract.BlockTypePage),
						Relationships: []*textract.Relationship{
							{
								Ids: []*string{
									aws.String("line_0"),
								},
							},
						},
						Page: aws.Int64(1),
					},
					{
						Id:        aws.String("page_1"),
						BlockType: aws.String(textract.BlockTypePage),
						Relationships: []*textract.Relationship{
							{
								Ids: []*string{
									aws.String("line_1"),
								},
							},
						},
						Page: aws.Int64(2),
					},
					{
						Id:        aws.String("line_0"),
						BlockType: aws.String(textract.BlockTypeLine),
						Text:      aws.String("test words"),
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
						Id:        aws.String("line_1"),
						BlockType: aws.String(textract.BlockTypeLine),
						Text:      aws.String("another"),
						Geometry: &textract.Geometry{
							BoundingBox: &textract.BoundingBox{
								Height: aws.Float64(0.2),
								Width:  aws.Float64(0.3),
								Top:    aws.Float64(0.1),
								Left:   aws.Float64(0.6),
							},
						},
					},
				},
			},
			document: Document{
				AccountID: accountID,
				Filename:  filename,
				Filepath:  filepath,
				Pages: []Page{
					{
						PageNumber: 1,
						Lines: []Line{
							{
								Text: "test words",
								Coordinates: Coordinates{
									TopLeft: Point{
										X: 0.1,
										Y: 0.1,
									},
									TopRight: Point{
										X: 0.6,
										Y: 0.1,
									},
									BottomLeft: Point{
										X: 0.1,
										Y: 0.3,
									},
									BottomRight: Point{
										X: 0.6,
										Y: 0.3,
									},
								},
							},
						},
					},
					{
						PageNumber: 2,
						Lines: []Line{
							{
								Text: "another",
								Coordinates: Coordinates{
									TopLeft: Point{
										X: 0.6,
										Y: 0.1,
									},
									TopRight: Point{
										X: 0.9,
										Y: 0.1,
									},
									BottomLeft: Point{
										X: 0.6,
										Y: 0.3,
									},
									BottomRight: Point{
										X: 0.9,
										Y: 0.3,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			document := convertToDocument(test.input, accountID, filename, filepath)

			if document.AccountID != accountID {
				t.Errorf("incorrect document account id, received: %s, expected: %s", document.AccountID, accountID)
			}

			if document.Filename != filename {
				t.Errorf("incorrect document filename, received: %s, expected: %s", document.Filename, filename)
			}

			if document.Filepath != filepath {
				t.Errorf("incorrect document filepath, received: %s, expected: %s", document.Filepath, filepath)
			}

			if len(document.Pages) != len(test.document.Pages) {
				t.Errorf("incorrect pages count, received: %d, expected: %d", len(document.Pages), len(test.document.Pages))
			}

			for i, receivedPage := range document.Pages {
				expectedPage := test.document.Pages[i]
				if receivedPage.PageNumber != expectedPage.PageNumber {
					t.Errorf("incorrect page number, received: %d, expected: %d", receivedPage.PageNumber, expectedPage.PageNumber)
				}

				for j, receivedLine := range receivedPage.Lines {
					expectedLine := expectedPage.Lines[j]
					if receivedLine.Text != expectedLine.Text {
						t.Errorf("incorrect line text, received: %s, expected: %s", receivedLine.Text, expectedLine.Text)
					}

					if !checkCoordinates(t, receivedLine.Coordinates, expectedLine.Coordinates) {
						t.Errorf("incorrect line coordinates, received: %+v, expected: %+v", receivedLine.Coordinates, expectedLine.Coordinates)
					}
				}
			}
		})
	}
}

func checkCoordinates(t *testing.T, a, b Coordinates) bool {
	t.Helper()

	if !checkPoint(t, a.TopLeft, b.TopLeft) {
		return false
	}

	if !checkPoint(t, a.TopRight, b.TopRight) {
		return false
	}

	if !checkPoint(t, a.BottomLeft, b.BottomLeft) {
		return false
	}

	if !checkPoint(t, a.BottomRight, b.BottomRight) {
		return false
	}

	return true
}

func checkPoint(t *testing.T, a, b Point) bool {
	t.Helper()

	tolerance := 0.00000001

	toleranceCheck := func(a, b float64) bool {
		return math.Abs(a-b) < tolerance
	}

	if !toleranceCheck(a.X, b.X) || !toleranceCheck(a.Y, b.Y) {
		return false
	}

	return true
}
