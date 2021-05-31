package csql

var _ CSQLer = &Client{}

// Client implements the csql.CSQLer methods.
type Client struct {
	parseJSON func(input interface{}) (interface{}, error)
}

// New generates a Client pointer instance.
func New() *Client {
	return &Client{
		parseJSON: parseJSON,
	}
}

// CSQLToES implements the csql.CSQLer.CSQLToES interface method.
func (c *Client) CSQLToES(csqlJSON map[string]interface{}) (map[string]interface{}, error) {
	esJSON, err := c.parseJSON(csqlJSON)
	if err != nil {
		return nil, &ErrorParseCSQLJSON{err: err}
	}

	return esJSON.(map[string]interface{}), nil
}
