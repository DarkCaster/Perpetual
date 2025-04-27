package op_embed

import (
	"math"
	"testing"
)

func TestCosineSearch(t *testing.T) {
	tests := []struct {
		name     string
		x        []float32
		y        []float32
		want     float32
		wantBool bool
	}{
		{
			name:     "identical vectors",
			x:        []float32{1.0, 2.0, 3.0},
			y:        []float32{1.0, 2.0, 3.0},
			want:     1.0,
			wantBool: true,
		},
		{
			name:     "orthogonal vectors",
			x:        []float32{1.0, 0.0, 0.0},
			y:        []float32{0.0, 1.0, 0.0},
			want:     0.0,
			wantBool: true,
		},
		{
			name:     "opposite vectors",
			x:        []float32{1.0, 2.0, 3.0},
			y:        []float32{-1.0, -2.0, -3.0},
			want:     -1.0,
			wantBool: true,
		},
		{
			name:     "different length vectors",
			x:        []float32{1.0, 2.0, 3.0},
			y:        []float32{1.0, 2.0},
			want:     -math.MaxFloat32,
			wantBool: false,
		},
		{
			name:     "zero vector",
			x:        []float32{0.0, 0.0, 0.0},
			y:        []float32{1.0, 2.0, 3.0},
			want:     0.0,
			wantBool: true,
		},
		{
			name:     "both zero vectors",
			x:        []float32{0.0, 0.0, 0.0},
			y:        []float32{0.0, 0.0, 0.0},
			want:     0.0, // NaN in pure math, but our implementation returns 0
			wantBool: true,
		},
		{
			name:     "similar vectors",
			x:        []float32{1.0, 2.0, 3.0},
			y:        []float32{2.0, 3.0, 4.0},
			want:     0.99258333,
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotBool := cosineSearch(tt.x, tt.y)
			if gotBool != tt.wantBool {
				t.Errorf("cosineSearch() bool = %v, want %v", gotBool, tt.wantBool)
			}
			if tt.wantBool {
				// Use approximate comparison for floating point values
				if math.Abs(float64(got-tt.want)) > 1e-6 {
					t.Errorf("cosineSearch() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestSimilaritySearch(t *testing.T) {
	// Test case for SimilaritySearch function
	searchVectors := [][]float32{
		{1.0, 0.0, 0.0},
		{0.0, 1.0, 0.0},
	}

	filesSourceVectors := map[string][][]float32{
		"file1.txt": {
			{1.0, 0.0, 0.0},
			{0.5, 0.5, 0.0},
		},
		"file2.txt": {
			{0.0, 1.0, 0.0},
			{0.0, 0.5, 0.5},
		},
		"file3.txt": {
			{0.0, 0.0, 1.0},
			{0.3, 0.3, 0.3},
		},
	}

	expected := map[string]float32{
		"file1.txt": 1.0,     // Perfect match with first search vector
		"file2.txt": 1.0,     // Perfect match with second search vector
		"file3.txt": 0.57735, // Best match is with the second vector
	}

	result := SimilaritySearch(searchVectors, filesSourceVectors)

	if len(result) != len(expected) {
		t.Errorf("SimilaritySearch() returned %d results, want %d", len(result), len(expected))
	}

	for file, score := range expected {
		if resultScore, ok := result[file]; !ok {
			t.Errorf("SimilaritySearch() missing result for %s", file)
		} else if math.Abs(float64(resultScore-score)) > 1e-5 {
			t.Errorf("SimilaritySearch() for %s = %v, want %v", file, resultScore, score)
		}
	}
}

// test using this example, but with cosine distance instead of dot:
// https://github.com/qdrant/qdrant/blob/master/docs/QUICK_START.md
func TestSimilaritySearchFromQdrantExample(t *testing.T) {
	// Test case for SimilaritySearch function
	searchVectors := [][]float32{
		{0.2, 0.1, 0.9, 0.7},
	}

	filesSourceVectors := map[string][][]float32{
		"id1": {
			{0.05, 0.61, 0.76, 0.74},
		},
		"id2": {
			{0.19, 0.81, 0.75, 0.11},
		},
		"id3": {
			{0.36, 0.55, 0.47, 0.94},
		},
		"id4": {
			{0.18, 0.01, 0.85, 0.80},
		},
		"id5": {
			{0.24, 0.18, 0.22, 0.44},
		},
		"id6": {
			{0.35, 0.08, 0.11, 0.44},
		},
	}

	expected := map[string]float32{
		"id4": 0.99248314,
		"id1": 0.89463294,
		"id5": 0.8543979,
		"id3": 0.83872515,
		"id6": 0.7216256,
		"id2": 0.66603535,
	}

	result := SimilaritySearch(searchVectors, filesSourceVectors)

	if len(result) != len(expected) {
		t.Errorf("SimilaritySearch() returned %d results, want %d", len(result), len(expected))
	}

	for file, score := range expected {
		if resultScore, ok := result[file]; !ok {
			t.Errorf("SimilaritySearch() missing result for %s", file)
		} else if math.Abs(float64(resultScore-score)) > 1e-5 {
			t.Errorf("SimilaritySearch() for %s = %v, want %v", file, resultScore, score)
		}
	}
}
