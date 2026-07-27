package embeddings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatchTexts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		texts     []string
		batchSize int
		expected  [][]string
	}{
		{
			texts:     []string{},
			batchSize: 1,
			expected:  [][]string{},
		},
		{
			texts:     []string{"foo bar zoo"},
			batchSize: 4,
			expected:  [][]string{{"foo bar zoo"}},
		},
		{
			texts:     []string{"foo bar zoo", "foo"},
			batchSize: 7,
			expected:  [][]string{{"foo bar zoo", "foo"}},
		},
		{
			texts:     []string{"foo", "bar", "zoo"},
			batchSize: 2,
			expected:  [][]string{{"foo", "bar"}, {"zoo"}},
		},
		{
			texts:     []string{"foo", "bar", "zoo", "baz", "qux"},
			batchSize: 2,
			expected:  [][]string{{"foo", "bar"}, {"zoo", "baz"}, {"qux"}},
		},
		{
			texts:     []string{"foo", "bar", "zoo", "baz"},
			batchSize: 2,
			expected:  [][]string{{"foo", "bar"}, {"zoo", "baz"}},
		},
		{
			texts:     []string{"foo", "bar", "zoo", "baz", "qux"},
			batchSize: 3,
			expected:  [][]string{{"foo", "bar", "zoo"}, {"baz", "qux"}},
		},
		{
			texts:     []string{"foo", "bar", "zoo", "baz", "qux"},
			batchSize: 6,
			expected:  [][]string{{"foo", "bar", "zoo", "baz", "qux"}},
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.expected, BatchTexts(tc.texts, tc.batchSize))
	}
}

func TestMinInt(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		nums     []int
		expected int
	}{
		{
			name:     "ascending order",
			nums:     []int{1, 2, 3, 23, 34},
			expected: 1,
		},
		{
			name:     "mixed order",
			nums:     []int{3, 2, 1, 34, 2213},
			expected: 1,
		},
		{
			name:     "nil slice",
			nums:     nil,
			expected: 0,
		},
		{
			name:     "empty slice",
			nums:     []int{},
			expected: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := minInt(tc.nums)
			if got != tc.expected {
				t.Errorf("MinInt(%v) = %v, want %v", tc.nums, got, tc.expected)
			}
		})
	}
}
