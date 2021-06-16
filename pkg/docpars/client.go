package docpars

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
	"github.com/google/uuid"
)

const (
	documentEntity = "document"
	pageEntity     = "page"
	lineEntity     = "line"
)

var _ Parser = &Client{}

// Client implements the docparser.Parser methods using AWS Textract.
type Client struct {
	textractClient    textractClient
	convertToDocument func(input *textract.AnalyzeDocumentOutput, accountID, filename, filepath string) Document
}

type textractClient interface {
	AnalyzeDocument(input *textract.AnalyzeDocumentInput) (*textract.AnalyzeDocumentOutput, error)
}

// New generates a Client pointer instance with an AWS Textract client.
func New() *Client {
	newSession := session.Must(session.NewSession())
	service := textract.New(newSession)

	return &Client{
		textractClient:    service,
		convertToDocument: convertToDocument,
	}
}

// Parse implements the docparser.Parser.Parse interface method.
func (c *Client) Parse(ctx context.Context, accountID, filename, filepath string, doc []byte) (*Document, error) {
	input := &textract.AnalyzeDocumentInput{
		Document: &textract.Document{
			Bytes: doc,
		},
		FeatureTypes: []*string{
			aws.String(textract.FeatureTypeTables),
			aws.String(textract.FeatureTypeForms),
		},
	}

	output, err := c.textractClient.AnalyzeDocument(input)
	if err != nil {
		return nil, &ErrorAnalyzeDocument{err: err}
	}

	document := c.convertToDocument(output, accountID, filename, filepath)

	return &document, nil
}

func convertToDocument(input *textract.AnalyzeDocumentOutput, accountID, filename, filepath string) Document {
	document := Document{
		ID:        uuid.NewString(),
		Entity:    documentEntity,
		AccountID: accountID,
		Filename:  filename,
		Filepath:  filepath,
	}

	pages := []*textract.Block{}
	lines := map[string]*textract.Block{}

	for _, block := range input.Blocks {
		if *block.BlockType == textract.BlockTypePage {
			pages = append(pages, block)
		}

		if *block.BlockType == textract.BlockTypeLine {
			lines[*block.Id] = block
		}
	}

	for _, pageBlock := range pages {
		page := Page{
			ID:     uuid.NewString(),
			Entity: pageEntity,
			Lines:  []Data{},
		}

		if pageBlock.Page == nil {
			page.PageNumber = 1
		} else {
			page.PageNumber = *pageBlock.Page
		}

		for _, id := range pageBlock.Relationships[0].Ids {
			// not all child IDs are "lines" which requires an ok check
			if lineBlock, ok := lines[*id]; ok {
				left := *lineBlock.Geometry.BoundingBox.Left
				top := *lineBlock.Geometry.BoundingBox.Top
				height := *lineBlock.Geometry.BoundingBox.Height
				width := *lineBlock.Geometry.BoundingBox.Width

				data := Data{
					ID:     uuid.NewString(),
					Entity: lineEntity,
					Text:   *lineBlock.Text,
					Coordinates: Coordinates{
						TopLeft: Point{
							X: left,
							Y: top,
						},
						TopRight: Point{
							X: left + width,
							Y: top,
						},
						BottomLeft: Point{
							X: left,
							Y: top + height,
						},
						BottomRight: Point{
							X: left + width,
							Y: top + height,
						},
					},
				}

				page.Lines = append(page.Lines, data)
			}
		}

		document.Pages = append(document.Pages, page)
	}

	return document
}