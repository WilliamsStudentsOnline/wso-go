package search_test

import (
	"testing"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	. "github.com/WilliamsStudentsOnline/wso-go/lib/search"
	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	testCases := []struct {
		name     string
		query    string
		expected *Query
	}{
		// The following are tests pulled from examples on the old WSO website to assert it functions the same
		{
			name:  "value",
			query: "Ephraim",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("Ephraim"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "implicit and values",
			query: "Ephraim Williams",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("Ephraim"),
								},
							},
							{
								Value: &Value{
									Word: lib.StrToPtr("Williams"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "explicit and values",
			query: "Ephraim AND Williams",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("Ephraim"),
								},
							},
							{
								Value: &Value{
									Word: lib.StrToPtr("Williams"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "complex and-or values",
			query: "Ephraim Williams OR Mark Hopkins",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("Ephraim"),
								},
							},
							{
								Value: &Value{
									Word: lib.StrToPtr("Williams"),
								},
							},
						},
					},
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("Mark"),
								},
							},
							{
								Value: &Value{
									Word: lib.StrToPtr("Hopkins"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "complex or-and values",
			query: "Ephraim OR Hopkins AND Thompson OR Griffin",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("Ephraim"),
								},
							},
						},
					},
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("Hopkins"),
								},
							},
							{
								Value: &Value{
									Word: lib.StrToPtr("Thompson"),
								},
							},
						},
					},
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("Griffin"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "complex group-or-and values",
			query: "(Ephraim OR Hopkins) AND (Thompson OR Griffin)",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Or: []*Expression{
									{
										And: []*Condition{
											{
												Value: &Value{
													Word: lib.StrToPtr("Ephraim"),
												},
											},
										},
									},
									{
										And: []*Condition{
											{
												Value: &Value{
													Word: lib.StrToPtr("Hopkins"),
												},
											},
										},
									},
								},
							},
							{
								Or: []*Expression{
									{
										And: []*Condition{
											{
												Value: &Value{
													Word: lib.StrToPtr("Thompson"),
												},
											},
										},
									},
									{
										And: []*Condition{
											{
												Value: &Value{
													Word: lib.StrToPtr("Griffin"),
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
		},
		{
			name:  "field",
			query: "name: Ephraim",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Field: &Field{
									Key: "name",
									Value: &Value{
										Word: lib.StrToPtr("Ephraim"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "string",
			query: "\"Ephraim Williams\"",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									String: lib.StrToPtr("Ephraim Williams"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "field string",
			query: "name: \"Ephraim Williams\"",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Field: &Field{
									Key: "name",
									Value: &Value{
										String: lib.StrToPtr("Ephraim Williams"),
									},
								},
							},
						},
					},
				},
			},
		},
		// Run abstract boolean expression tests
		{
			name:  "boolean expression a&b|c&d",
			query: "a AND b OR c AND d",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("a"),
								},
							},
							{
								Value: &Value{
									Word: lib.StrToPtr("b"),
								},
							},
						},
					},
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("c"),
								},
							},
							{
								Value: &Value{
									Word: lib.StrToPtr("d"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "boolean expression a|b&c|d",
			query: "a OR b AND c OR d",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("a"),
								},
							},
						},
					},
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("b"),
								},
							},
							{
								Value: &Value{
									Word: lib.StrToPtr("c"),
								},
							},
						},
					},
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("d"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "boolean expression a&b|c",
			query: "a AND b OR c",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("a"),
								},
							},
							{
								Value: &Value{
									Word: lib.StrToPtr("b"),
								},
							},
						},
					},
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("c"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "boolean expression a|b&c",
			query: "a OR b AND c",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("a"),
								},
							},
						},
					},
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("b"),
								},
							},
							{
								Value: &Value{
									Word: lib.StrToPtr("c"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "boolean expression a&b&c",
			query: "a AND b AND c",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("a"),
								},
							},
							{
								Value: &Value{
									Word: lib.StrToPtr("b"),
								},
							},
							{
								Value: &Value{
									Word: lib.StrToPtr("c"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "boolean expression a|b|c",
			query: "a OR b OR c",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("a"),
								},
							},
						},
					},
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("b"),
								},
							},
						},
					},
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("c"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "boolean expression a&b",
			query: "a AND b",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("a"),
								},
							},
							{
								Value: &Value{
									Word: lib.StrToPtr("b"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "boolean expression a|b",
			query: "a OR b",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("a"),
								},
							},
						},
					},
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("b"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "boolean expression a",
			query: "a",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Value: &Value{
									Word: lib.StrToPtr("a"),
								},
							},
						},
					},
				},
			},
		},
		// Other custom expressions that I have come across
		{
			name:  "two fields",
			query: "room: 103 dorm: East",
			expected: &Query{
				Or: []*Expression{
					{
						And: []*Condition{
							{
								Field: &Field{
									Key: "room",
									Value: &Value{
										Int: lib.IntToPtr(103),
									},
								},
							},
							{
								Field: &Field{
									Key: "dorm",
									Value: &Value{
										Word: lib.StrToPtr("East"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ast, err := ParseSearchQuery(tc.query)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, ast)
		})
	}

	// Ensure this doesnt break anything
	_, err := ParseSearchQuery("but (what does this do) this too \"plus also this??\" foo:\"bar:baz\" hello:wo-rld hi there \"plus this?\" AND (this:that) OR \"fpp\"")
	assert.NoError(t, err)
}

func BenchmarkParseSearchQuery(b *testing.B) {
	benchmarks := []struct {
		name  string
		query string
	}{
		// Test case benchmarks
		{
			"value",
			"Ephraim",
		},
		{
			"implicit and values",
			"Ephraim Williams",
		},
		{
			"explicit and values",
			"Ephraim AND Williams",
		},
		{
			"complex and-or values",
			"Ephraim Williams OR Mark Hopkins",
		},
		{
			"complex or-and values",
			"Ephraim OR Hopkins AND Thompson OR Griffin",
		},
		{
			"complex group-or-and values",
			"(Ephraim OR Hopkins) AND (Thompson OR Griffin)",
		},
		{
			"field",
			"name: Ephraim",
		},
		{
			"string",
			"\"Ephraim Williams\"",
		},
		{
			"field string",
			"name: \"Ephraim Williams\"",
		},
		{
			"boolean expression a&b|c&d",
			"a AND b OR c AND d",
		},
		{
			"boolean expression a|b&c|d",
			"a OR b AND c OR d",
		},
		{
			"boolean expression a&b|c",
			"a AND b OR c",
		},
		{
			"boolean expression a|b&c",
			"a OR b AND c",
		},
		{
			"boolean expression a&b&c",
			"a AND b AND c",
		},
		{
			"boolean expression a|b|c",
			"a OR b OR c",
		},
		{
			"boolean expression a&b",
			"a AND b",
		},
		{
			"boolean expression a|b",
			"a OR b",
		},
		{
			"boolean expression a",
			"a",
		},
		// Actual other queries
		{
			"Complex Query",
			"but (what does this do) this too \"plus also this??\" foo:\"bar:baz\" hello:world hi there \"plus this?\" AND (this:that) OR \"fpp\"",
		},
		{
			"Regular Advanced",
			"Aidan Lloyd-Tucker home:\"Palo Alto\"",
		},
		{
			"Basic",
			"Aidan Lloyd-Tucker",
		},
		{
			"Unix",
			"al15",
		},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_, _ = ParseSearchQuery(bm.query)
			}
		})
	}
}
