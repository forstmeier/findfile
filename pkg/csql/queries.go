package csql

func newSearchQuery(input searchObject) map[string]interface{} {
	// flipping coordinate values to map between Textract and
	// ElasticSearch which operate with different x-y axes
	topLeft := input.Coordinates[1]
	bottomRight := input.Coordinates[0]

	return map[string]interface{}{
		"bool": map[string]interface{}{
			"must": map[string]interface{}{
				"match": map[string]interface{}{
					"text": input.Text,
				},
			},
			"filter": map[string]interface{}{
				"geo_shape": map[string]interface{}{
					"coordinates": map[string]interface{}{
						"shape": map[string]interface{}{
							"type": "envelope",
							"coordinates": [2][2]float64{
								{
									topLeft[0],
									topLeft[1],
								},
								{
									bottomRight[0],
									bottomRight[1],
								},
							},
						},
					},
				},
			},
		},
	}
}

func newOperatorQuery(input interface{}, name string) interface{} {
	return map[string]interface{}{
		"bool": map[string]interface{}{
			name: input,
		},
	}
}
