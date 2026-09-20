package redis

import (
	"encoding/binary"
	"math"
	"testing"
	"time"
)

func TestVectorBytes(t *testing.T) {
	got := VectorBytes([]float32{1.5, -2})
	if len(got) != 8 || math.Float32frombits(binary.LittleEndian.Uint32(got)) != 1.5 || math.Float32frombits(binary.LittleEndian.Uint32(got[4:])) != -2 {
		t.Fatalf("unexpected vector encoding: %v", got)
	}
}

func TestSearchFieldsDecodesRESP3Result(t *testing.T) {
	expires := time.Now().UTC().Format(time.RFC3339Nano)
	fields, found, err := searchFields(map[interface{}]interface{}{
		"total_results": int64(1),
		"results": []interface{}{map[interface{}]interface{}{
			"id": "semantic_cache:1",
			"extra_attributes": map[interface{}]interface{}{
				"response":        []byte("cached response"),
				"expires_at":      expires,
				"distance":        "0.00001",
				"embedding_model": "embed",
			},
		}},
	})
	if err != nil || !found {
		t.Fatalf("found=%v err=%v", found, err)
	}
	if fields["response"] != "cached response" || fields["expires_at"] != expires || fields["distance"] != "0.00001" {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}
func TestSimilarity(t *testing.T) {
	if got := Similarity(.04); math.Abs(got-.96) > 1e-9 {
		t.Fatalf("similarity = %v", got)
	}
}
func TestValidateDimension(t *testing.T) {
	if err := validateDimension(3, 2); err == nil {
		t.Fatal("expected mismatch error")
	}
	if err := validateDimension(0, 2); err != nil {
		t.Fatal(err)
	}
}

func TestNewConfiguresPassword(t *testing.T) {
	for _, password := range []string{"test-password", ""} {
		t.Run(password, func(t *testing.T) {
			cache := New("localhost:6379", password, 3)
			defer cache.Close()

			if got := cache.client.Options().Password; got != password {
				t.Fatalf("Redis password = %q, want %q", got, password)
			}
		})
	}
}
