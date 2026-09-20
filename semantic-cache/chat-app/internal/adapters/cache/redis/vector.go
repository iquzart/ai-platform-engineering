package redis

import (
	"encoding/binary"
	"fmt"
	"math"
)

func VectorBytes(vector []float32) []byte {
	result := make([]byte, len(vector)*4)
	for i, v := range vector {
		binary.LittleEndian.PutUint32(result[i*4:], math.Float32bits(v))
	}
	return result
}
func Similarity(distance float64) float64 { return 1 - distance }
func validateDimension(configured, actual int) error {
	if actual == 0 {
		return fmt.Errorf("embedding has no dimensions")
	}
	if configured > 0 && configured != actual {
		return fmt.Errorf("embedding dimensions = %d, configured dimension = %d", actual, configured)
	}
	return nil
}
