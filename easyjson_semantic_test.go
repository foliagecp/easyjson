package easyjson

import (
	"encoding/json"
	"math"
	"testing"
)

type semanticTestStruct struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestSemanticallyEquals(t *testing.T) {
	pointerNumber := 1
	var nilSlice []interface{}

	tests := []struct {
		name  string
		left  JSON
		right JSON
		want  bool
	}{
		{
			name: "object key order is ignored",
			left: NewJSON(map[string]interface{}{
				"name": "server",
				"resources": map[string]interface{}{
					"cpu": int64(8),
					"ram": float64(16),
				},
			}),
			right: NewJSON(map[string]interface{}{
				"resources": map[string]interface{}{
					"ram": int(16),
					"cpu": float64(8),
				},
				"name": "server",
			}),
			want: true,
		},
		{
			name:  "array order is significant",
			left:  NewJSON([]interface{}{"first", "second"}),
			right: NewJSON([]interface{}{"second", "first"}),
			want:  false,
		},
		{
			name:  "pointer and value are equivalent",
			left:  NewJSON(&pointerNumber),
			right: NewJSON(1),
			want:  true,
		},
		{
			name:  "equivalent number representations",
			left:  NewJSON([]interface{}{int(1), float64(1), json.Number("1e0")}),
			right: NewJSON([]interface{}{json.Number("1.0"), uint64(1), float32(1)}),
			want:  true,
		},
		{
			name:  "large integers retain precision",
			left:  NewJSON(uint64(9_007_199_254_740_993)),
			right: NewJSON(float64(9_007_199_254_740_993)),
			want:  false,
		},
		{
			name:  "large equivalent integer and float",
			left:  NewJSON(int64(1_000_000_000_000_000_000)),
			right: NewJSON(float64(1_000_000_000_000_000_000)),
			want:  true,
		},
		{
			name:  "null differs from object",
			left:  NewJSONNull(),
			right: NewJSONObject(),
			want:  false,
		},
		{
			name:  "null differs from array",
			left:  NewJSONNull(),
			right: NewJSONArray(),
			want:  false,
		},
		{
			name:  "object differs from array",
			left:  NewJSONObject(),
			right: NewJSONArray(),
			want:  false,
		},
		{
			name:  "nil slice represents null",
			left:  NewJSON(nilSlice),
			right: NewJSONNull(),
			want:  true,
		},
		{
			name:  "nil slice differs from empty array",
			left:  NewJSON(nilSlice),
			right: NewJSONArray(),
			want:  false,
		},
		{
			name:  "typed nil containers represent null",
			left:  NewJSON(map[string]interface{}(nil)),
			right: NewJSONNull(),
			want:  true,
		},
		{
			name:  "invalid JSON numbers are not equal",
			left:  NewJSON(json.Number("not-a-number")),
			right: NewJSON(json.Number("not-a-number")),
			want:  false,
		},
		{
			name:  "JSON null is not a number",
			left:  NewJSON(json.Number("null")),
			right: NewJSON(json.Number("null")),
			want:  false,
		},
		{
			name:  "JSON string is not a number",
			left:  NewJSON(json.Number(`"abc"`)),
			right: NewJSON(json.Number(`"abc"`)),
			want:  false,
		},
		{
			name:  "JSON object is not a number",
			left:  NewJSON(json.Number(`{"a":1}`)),
			right: NewJSON(json.Number(`{"a":1}`)),
			want:  false,
		},
		{
			name:  "number with surrounding whitespace is invalid",
			left:  NewJSON(json.Number(" 1 ")),
			right: NewJSON(json.Number("1")),
			want:  false,
		},
		{
			name:  "non-finite numbers are not JSON numbers",
			left:  NewJSON(math.NaN()),
			right: NewJSON(math.NaN()),
			want:  false,
		},
		{
			name:  "number differs from string",
			left:  NewJSON(1),
			right: NewJSON("1"),
			want:  false,
		},
		{
			name:  "boolean differs from number",
			left:  NewJSON(true),
			right: NewJSON(1),
			want:  false,
		},
		{
			name:  "same size objects with different keys differ",
			left:  NewJSON(map[string]interface{}{"left": 1}),
			right: NewJSON(map[string]interface{}{"right": 1}),
			want:  false,
		},
		{
			name: "deeply nested values",
			left: NewJSON(map[string]interface{}{
				"a": []interface{}{map[string]interface{}{
					"b": []interface{}{json.Number("1e0")},
				}},
			}),
			right: NewJSON(map[string]interface{}{
				"a": []interface{}{map[string]interface{}{
					"b": []interface{}{float64(1)},
				}},
			}),
			want: true,
		},
		{
			name:  "struct values use their JSON representation",
			left:  NewJSON(semanticTestStruct{Name: "server", Count: 1}),
			right: NewJSON(map[string]interface{}{"count": float64(1), "name": "server"}),
			want:  true,
		},
		{
			name:  "byte slices compare as JSON arrays",
			left:  NewJSON([]byte{1, 2, 3}),
			right: NewJSON([]interface{}{float64(1), float64(2), float64(3)}),
			want:  true,
		},
		{
			name:  "huge equivalent exponents",
			left:  NewJSON(json.Number("1e100000")),
			right: NewJSON(json.Number("10e99999")),
			want:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.left.SemanticallyEquals(test.right); got != test.want {
				t.Fatalf(
					"SemanticallyEquals(%s, %s) = %v, want %v",
					test.left.ToString(),
					test.right.ToString(),
					got,
					test.want,
				)
			}
			if got := test.right.SemanticallyEquals(test.left); got != test.want {
				t.Fatalf("SemanticallyEquals must be symmetric: got %v, want %v", got, test.want)
			}
		})
	}
}

func TestSemanticallyEqualsAfterJSONRoundTrip(t *testing.T) {
	original := NewJSON(map[string]interface{}{
		"count": int64(1),
		"items": []interface{}{uint64(2), int32(3)},
	})
	roundTripped, ok := JSONFromBytes(original.ToBytes())
	if !ok {
		t.Fatal("JSONFromBytes failed")
	}
	if original.Equals(roundTripped) {
		t.Fatal("exact Equals should observe changed concrete numeric types")
	}
	if !original.SemanticallyEquals(roundTripped) {
		t.Fatal("SemanticallyEquals should preserve JSON meaning across a round trip")
	}
}

func TestSemanticallyEqualsHandlesCycles(t *testing.T) {
	left := map[string]interface{}{}
	left["self"] = left
	right := map[string]interface{}{}
	right["self"] = right

	if !NewJSON(left).SemanticallyEquals(NewJSON(right)) {
		t.Fatal("equivalent cyclic maps should terminate and compare equal")
	}

	var pointerCycle interface{}
	pointerCycle = &pointerCycle
	value := NewJSON(pointerCycle)
	if value.SemanticallyEquals(value) {
		t.Fatal("a cyclic pointer chain is not a representable JSON value")
	}
}
