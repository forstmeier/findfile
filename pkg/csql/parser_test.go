package csql

import (
	"reflect"
	"testing"
)

func Test_parseJSONObject(t *testing.T) {
	tests := []struct {
		description string
		input       interface{}
		esQuery     interface{}
		error       error
	}{
		{
			description: "more than one attribute in root csql object",
			input: map[string]interface{}{
				"first":  "value",
				"second": "value",
			},
			esQuery: nil,
			error:   errorTooManyAttributes,
		},
		{
			description: "validation error in csql object",
			input: map[string]interface{}{
				"search": map[string]interface{}{
					"text": "",
				},
			},
			esQuery: nil,
			error:   errorMissingText,
		},
		{
			description: "type received in csql not supported",
			input:       "type_not_supported",
			esQuery:     nil,
			error:       errorTypeNotSupported,
		},
		{
			description: "successful invocation single csql search object",
			input: map[string]interface{}{
				"search": map[string]interface{}{
					"text": "lookup",
					"page": 1,
					"coordinates": [2][2]float64{
						{0.1, 0.2},
						{0.3, 0.4},
					},
				},
			},
			esQuery: map[string]interface{}{
				"bool": map[string]interface{}{
					"filter": map[string]interface{}{
						"geo_shape": map[string]interface{}{
							"coordinates": map[string]interface{}{
								"shape": map[string]interface{}{
									"type": "envelope",
									"coordinates": [2][2]float64{
										{0.3, 0.4},
										{0.1, 0.2},
									},
								},
							},
						},
					},
					"must": map[string]interface{}{
						"match": map[string]interface{}{
							"text": "lookup",
						},
					},
				},
			},
			error: nil,
		},
		{
			description: "single csql \"and\" array with child error",
			input: map[string]interface{}{
				"and": []interface{}{
					map[string]interface{}{
						"search": map[string]interface{}{
							"text": "",
						},
					},
				},
			},
			esQuery: nil,
			error:   errorMissingText,
		},
		{
			description: "successful invocation single csql \"and\" array with two search objects",
			input: map[string]interface{}{
				"and": []interface{}{
					map[string]interface{}{
						"search": map[string]interface{}{
							"text": "lookup",
							"page": 1,
							"coordinates": [2][2]float64{
								{0.1, 0.2},
								{0.3, 0.4},
							},
						},
					},
					map[string]interface{}{
						"search": map[string]interface{}{
							"text": "another",
							"page": 1,
							"coordinates": [2][2]float64{
								{0.3, 0.4},
								{0.5, 0.6},
							},
						},
					},
				},
			},
			esQuery: map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []interface{}{
						map[string]interface{}{
							"bool": map[string]interface{}{
								"filter": map[string]interface{}{
									"geo_shape": map[string]interface{}{
										"coordinates": map[string]interface{}{
											"shape": map[string]interface{}{
												"type": "envelope",
												"coordinates": [2][2]float64{
													{0.3, 0.4},
													{0.1, 0.2},
												},
											},
										},
									},
								},
								"must": map[string]interface{}{
									"match": map[string]interface{}{
										"text": "lookup",
									},
								},
							},
						},
						map[string]interface{}{
							"bool": map[string]interface{}{
								"filter": map[string]interface{}{
									"geo_shape": map[string]interface{}{
										"coordinates": map[string]interface{}{
											"shape": map[string]interface{}{
												"type": "envelope",
												"coordinates": [2][2]float64{
													{0.5, 0.6},
													{0.3, 0.4},
												},
											},
										},
									},
								},
								"must": map[string]interface{}{
									"match": map[string]interface{}{
										"text": "another",
									},
								},
							},
						},
					},
				},
			},
			error: nil,
		},
		{
			description: "single csql \"or\" array with child error",
			input: map[string]interface{}{
				"or": []interface{}{
					map[string]interface{}{
						"search": map[string]interface{}{
							"text": "",
						},
					},
				},
			},
			esQuery: nil,
			error:   errorMissingText,
		},
		{
			description: "successful invocation single csql \"or\" array with two search objects",
			input: map[string]interface{}{
				"or": []interface{}{
					map[string]interface{}{
						"search": map[string]interface{}{
							"text": "lookup",
							"page": 1,
							"coordinates": [2][2]float64{
								{0.1, 0.2},
								{0.3, 0.4},
							},
						},
					},
					map[string]interface{}{
						"search": map[string]interface{}{
							"text": "another",
							"page": 1,
							"coordinates": [2][2]float64{
								{0.3, 0.4},
								{0.5, 0.6},
							},
						},
					},
				},
			},
			esQuery: map[string]interface{}{
				"bool": map[string]interface{}{
					"should": []interface{}{
						map[string]interface{}{
							"bool": map[string]interface{}{
								"filter": map[string]interface{}{
									"geo_shape": map[string]interface{}{
										"coordinates": map[string]interface{}{
											"shape": map[string]interface{}{
												"type": "envelope",
												"coordinates": [2][2]float64{
													{0.3, 0.4},
													{0.1, 0.2},
												},
											},
										},
									},
								},
								"must": map[string]interface{}{
									"match": map[string]interface{}{
										"text": "lookup",
									},
								},
							},
						},
						map[string]interface{}{
							"bool": map[string]interface{}{
								"filter": map[string]interface{}{
									"geo_shape": map[string]interface{}{
										"coordinates": map[string]interface{}{
											"shape": map[string]interface{}{
												"type": "envelope",
												"coordinates": [2][2]float64{
													{0.5, 0.6},
													{0.3, 0.4},
												},
											},
										},
									},
								},
								"must": map[string]interface{}{
									"match": map[string]interface{}{
										"text": "another",
									},
								},
							},
						},
					},
				},
			},
			error: nil,
		},
		{
			description: "single csql \"not\" array with child error",
			input: map[string]interface{}{
				"not": []interface{}{
					map[string]interface{}{
						"search": map[string]interface{}{
							"text": "",
						},
					},
				},
			},
			esQuery: nil,
			error:   errorMissingText,
		},
		{
			description: "successful invocation single csql \"not\" array with two search objects",
			input: map[string]interface{}{
				"not": []interface{}{
					map[string]interface{}{
						"search": map[string]interface{}{
							"text": "lookup",
							"page": 1,
							"coordinates": [2][2]float64{
								{0.1, 0.2},
								{0.3, 0.4},
							},
						},
					},
					map[string]interface{}{
						"search": map[string]interface{}{
							"text": "another",
							"page": 1,
							"coordinates": [2][2]float64{
								{0.3, 0.4},
								{0.5, 0.6},
							},
						},
					},
				},
			},
			esQuery: map[string]interface{}{
				"bool": map[string]interface{}{
					"must_not": []interface{}{
						map[string]interface{}{
							"bool": map[string]interface{}{
								"filter": map[string]interface{}{
									"geo_shape": map[string]interface{}{
										"coordinates": map[string]interface{}{
											"shape": map[string]interface{}{
												"type": "envelope",
												"coordinates": [2][2]float64{
													{0.3, 0.4},
													{0.1, 0.2},
												},
											},
										},
									},
								},
								"must": map[string]interface{}{
									"match": map[string]interface{}{
										"text": "lookup",
									},
								},
							},
						},
						map[string]interface{}{
							"bool": map[string]interface{}{
								"filter": map[string]interface{}{
									"geo_shape": map[string]interface{}{
										"coordinates": map[string]interface{}{
											"shape": map[string]interface{}{
												"type": "envelope",
												"coordinates": [2][2]float64{
													{0.5, 0.6},
													{0.3, 0.4},
												},
											},
										},
									},
								},
								"must": map[string]interface{}{
									"match": map[string]interface{}{
										"text": "another",
									},
								},
							},
						},
					},
				},
			},
			error: nil,
		},
		{
			description: "successful invocation single csql \"and\" array with child search object and \"or\" array with two child search objects",
			input: map[string]interface{}{
				"and": []interface{}{
					map[string]interface{}{
						"search": map[string]interface{}{
							"text": "lookup",
							"page": 1,
							"coordinates": [2][2]float64{
								{0.1, 0.2},
								{0.3, 0.4},
							},
						},
					},
					map[string]interface{}{
						"or": []interface{}{
							map[string]interface{}{
								"search": map[string]interface{}{
									"text": "another",
									"page": 1,
									"coordinates": [2][2]float64{
										{0.3, 0.4},
										{0.5, 0.6},
									},
								},
							},
							map[string]interface{}{
								"search": map[string]interface{}{
									"text": "alternative",
									"page": 1,
									"coordinates": [2][2]float64{
										{0.3, 0.4},
										{0.5, 0.6},
									},
								},
							},
						},
					},
				},
			},
			esQuery: map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []interface{}{
						map[string]interface{}{
							"bool": map[string]interface{}{
								"filter": map[string]interface{}{
									"geo_shape": map[string]interface{}{
										"coordinates": map[string]interface{}{
											"shape": map[string]interface{}{
												"type": "envelope",
												"coordinates": [2][2]float64{
													{0.3, 0.4},
													{0.1, 0.2},
												},
											},
										},
									},
								},
								"must": map[string]interface{}{
									"match": map[string]interface{}{
										"text": "lookup",
									},
								},
							},
						},
						map[string]interface{}{
							"bool": map[string]interface{}{
								"should": []interface{}{
									map[string]interface{}{
										"bool": map[string]interface{}{
											"filter": map[string]interface{}{
												"geo_shape": map[string]interface{}{
													"coordinates": map[string]interface{}{
														"shape": map[string]interface{}{
															"type": "envelope",
															"coordinates": [2][2]float64{
																{0.5, 0.6},
																{0.3, 0.4},
															},
														},
													},
												},
											},
											"must": map[string]interface{}{
												"match": map[string]interface{}{
													"text": "another",
												},
											},
										},
									},
									map[string]interface{}{
										"bool": map[string]interface{}{
											"filter": map[string]interface{}{
												"geo_shape": map[string]interface{}{
													"coordinates": map[string]interface{}{
														"shape": map[string]interface{}{
															"type": "envelope",
															"coordinates": [2][2]float64{
																{0.5, 0.6},
																{0.3, 0.4},
															},
														},
													},
												},
											},
											"must": map[string]interface{}{
												"match": map[string]interface{}{
													"text": "alternative",
												},
											},
										},
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
			esQuery, err := parseJSON(test.input)

			if err != test.error {
				t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
			}

			if !reflect.DeepEqual(esQuery, test.esQuery) {
				t.Errorf("incorrect es query, received: %v, expected: %v", esQuery, test.esQuery)
			}
		})
	}
}

func Test_validateSearchJSON(t *testing.T) {
	tests := []struct {
		description string
		input       searchObject
		error       error
	}{
		{
			description: "empty text field",
			input: searchObject{
				Text: "",
			},
			error: errorMissingText,
		},
		{
			description: "empty page number field",
			input: searchObject{
				Text: "search value",
				Page: 0,
			},
			error: errorPageNumberZero,
		},
		{
			description: "bottom right coordinates equal zero",
			input: searchObject{
				Text: "search value",
				Page: 1,
				Coordinates: [2][2]float64{
					{
						float64(0),
						float64(0),
					},
					{
						float64(0),
						float64(0),
					},
				},
			},
			error: errorCoordinatesZero,
		},
		{
			description: "top left values equal bottom right values",
			input: searchObject{
				Text: "search value",
				Page: 1,
				Coordinates: [2][2]float64{
					{
						float64(0.3),
						float64(0.4),
					},
					{
						float64(0.3),
						float64(0.4),
					},
				},
			},
			error: errorCoordinatesMisplaced,
		},
		{
			description: "top left values greater than bottom right values",
			input: searchObject{
				Text: "search value",
				Page: 1,
				Coordinates: [2][2]float64{
					{
						float64(0.3),
						float64(0.4),
					},
					{
						float64(0.1),
						float64(0.2),
					},
				},
			},
			error: errorCoordinatesMisplaced,
		},
		{
			description: "successful invocation with correct csql query",
			input: searchObject{
				Text: "search value",
				Page: 1,
				Coordinates: [2][2]float64{
					{
						float64(0.1),
						float64(0.2),
					},
					{
						float64(0.3),
						float64(0.4),
					},
				},
			},
			error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if err := validateSearchJSON(test.input); err != test.error {
				t.Errorf("incorrect error, received: %v, expected: %v", err, test.error)
			}
		})
	}
}
