package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    " hello world ",
			expected: []string{"hello", "world"},
		},
		{
			input:    " testing testing testing",
			expected: []string{"testing", "testing", "testing"},
		},
		{
			input:    "             many spaces              ",
			expected: []string{"many", "spaces"},
		},
		{
			input:    "nospaces",
			expected: []string{"nospaces"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)

		if len(actual) == 0 {
			t.Errorf("cleanInput didn't return anything")
		}

		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]

			if word != expectedWord {
				t.Errorf("Actual did not match expected\nActual: %s\nExpected: %s", word, expectedWord)
			}
		}
	}
}
