package easyjson

import (
	"encoding/json"
	"testing"
)

// helper: must parse JSON string
func mustJSONFromString(t *testing.T, s string) JSON {
	t.Helper()
	j, ok := JSONFromString(s)
	if !ok {
		t.Fatalf("failed to parse JSON: %s", s)
	}
	return j
}

// helper: normalize both and assert equality
func assertNormalizedEqual(t *testing.T, a, b JSON) {
	t.Helper()
	a.Normalize()
	b.Normalize()
	if !a.Equals(b) {
		t.Fatalf("normalized JSONs are not equal\nA: %s\nB: %s", a.ToString(), b.ToString())
	}
}

// helper: normalize both and assert inequality
func assertNormalizedNotEqual(t *testing.T, a, b JSON) {
	t.Helper()
	a.Normalize()
	b.Normalize()
	if a.Equals(b) {
		t.Fatalf("expected inequality after normalization, but got equal\nA: %s\nB: %s", a.ToString(), b.ToString())
	}
}

func TestNormalize_ArrayOrder(t *testing.T) {
	a := mustJSONFromString(t, `[1,2,3,{"a":1,"b":2}]`)
	b := mustJSONFromString(t, `[{"b":2,"a":1},3,2,1]`)
	assertNormalizedEqual(t, a, b)
}

func TestNormalize_NestedStructures(t *testing.T) {
	a := mustJSONFromString(t, `{
		"a": [
			{"b": 2, "a": 1},
			{"x": [3,1,2]},
			[ {"k":3}, {"k":1}, {"k":2} ]
		],
		"z": {"m": [ {"u":2}, {"u":1} ]}
	}`)
	b := mustJSONFromString(t, `{
		"a": [
			{"x": [2,3,1]},
			{"a": 1, "b": 2},
			[ {"k":2}, {"k":1}, {"k":3} ]
		],
		"z": {"m": [ {"u":1}, {"u":2} ]}
	}`)
	assertNormalizedEqual(t, a, b)
}

func TestNormalize_DuplicatesPreserved(t *testing.T) {
	a := mustJSONFromString(t, `[1,1,2,3,3,3]`)
	b := mustJSONFromString(t, `[3,1,3,2,1,3]`)
	assertNormalizedEqual(t, a, b)
}

func TestNormalize_DifferentMultiplicityNotEqual(t *testing.T) {
	a := mustJSONFromString(t, `[1,2,2]`)
	b := mustJSONFromString(t, `[1,2]`)
	assertNormalizedNotEqual(t, a, b)
}

func TestNormalize_NumberUnification_Primitives(t *testing.T) {
	// Build using native Go ints
	a := NewJSON(map[string]interface{}{
		"n1": int(1),
		"n2": int64(1),
		"n3": uint(1),
	})

	// Build using JSON float literal
	b := mustJSONFromString(t, `{"n1":1.0,"n2":1.0,"n3":1.0}`)

	assertNormalizedEqual(t, a, b)
}

func TestNormalize_NumberUnification_JsonNumber(t *testing.T) {
	// Force json.Number in a nested structure
	m := map[string]interface{}{
		"obj": map[string]interface{}{
			"n": json.Number("42"),
		},
	}
	a := NewJSON(m)
	b := mustJSONFromString(t, `{"obj":{"n":42}}`)
	assertNormalizedEqual(t, a, b)
}

func TestNormalize_MixedArrayTypesOrderAgnostic(t *testing.T) {
	a := mustJSONFromString(t, `[{"a":1},"x",true,3,null]`)
	b := mustJSONFromString(t, `[null,3,true,"x",{"a":1}]`)
	assertNormalizedEqual(t, a, b)
}

func TestNormalize_Idempotent(t *testing.T) {
	j := mustJSONFromString(t, `{"a":[{"b":2,"a":1},{"x":[2,3,1]}]}`)
	j2 := j.NormalizedClone()
	j2_again := j2.NormalizedClone()

	// Both should be equal after repeated normalization
	if !j2.Equals(j2_again) {
		t.Fatalf("normalization should be idempotent\nFirst: %s\nSecond: %s",
			j2.ToString(), j2_again.ToString())
	}
}

func TestNormalize_NotEqualDifferentContent(t *testing.T) {
	a := mustJSONFromString(t, `[{"a":1},{"b":2}]`)
	b := mustJSONFromString(t, `[{"a":1},{"b":3}]`)
	assertNormalizedNotEqual(t, a, b)
}
