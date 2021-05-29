package docpars

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
	"github.com/google/uuid"
)

var _ Parser = &Client{}

// Client implements the docparser.Parser methods using AWS Textract.
type Client struct {
	textractClient   textractClient
	convertToContent func(input *textract.AnalyzeDocumentOutput) Content
}

type textractClient interface {
	AnalyzeDocument(input *textract.AnalyzeDocumentInput) (*textract.AnalyzeDocumentOutput, error)
}

// New generates a Client pointer instance with an AWS Textract client.
func New() *Client {
	newSession := session.Must(session.NewSession())
	service := textract.New(newSession)

	return &Client{
		textractClient:   service,
		convertToContent: convertToContent,
	}
}

// Parse implements the docparser.Parser.Parse interface method.
func (c *Client) Parse(ctx context.Context, doc []byte) (*Content, error) {
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

	content := c.convertToContent(output)

	return &content, nil
}

func convertToContent(input *textract.AnalyzeDocumentOutput) Content {
	content := Content{
		ID: uuid.NewString(),
	}

	for _, block := range input.Blocks {
		left := *block.Geometry.BoundingBox.Left
		top := *block.Geometry.BoundingBox.Top
		height := *block.Geometry.BoundingBox.Height
		width := *block.Geometry.BoundingBox.Width

		data := Data{
			Text: *block.Text,
			Coordinates: Coordinates{
				TopLeft: Coordinate{
					X: left,
					Y: top,
				},
				TopRight: Coordinate{
					X: left + width,
					Y: top,
				},
				BottomLeft: Coordinate{
					X: left,
					Y: top + height,
				},
				BottomRight: Coordinate{
					X: left + width,
					Y: top + height,
				},
			},
		}

		if *block.BlockType == textract.BlockTypeLine {
			content.Lines = append(content.Lines, data)
		}

		if *block.BlockType == textract.BlockTypeWord {
			content.Words = append(content.Lines, data)
		}
	}

	return content
}
