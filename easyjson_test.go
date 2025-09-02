package easyjson

import (
	"testing"
)

func TestReproduce_ArrayWithHoles(t *testing.T) {
	j := NewJSONObject()
	ok := j.SetByPath("1:3:6:1:2:1:17:1:4:1:2.3", NewJSON(79))
	if !ok {
		t.Fatalf("SetByPath failed")
	}
	got := j.ToString()
	t.Log(got)
}

func TestNewJSONObjectFromMap(t *testing.T) {
	data := map[string]interface{}{
		"name":        "test",
		"value":       42,
		"nested.path": "works",
	}

	json := NewJSONObjectFromMap(data)

	if json.GetByPath("name").AsStringDefault("") != "test" {
		t.Error("name not set correctly")
	}

	if json.GetByPath("value").AsNumericDefault(0) != 42 {
		t.Error("value not set correctly")
	}

	if json.GetByPath("nested.path").AsStringDefault("") != "works" {
		t.Error("nested path not set correctly")
	}
}

func TestAddToArrayBuilder(t *testing.T) {
	builder := NewJSONBuilder()

	// Test adding to non-existent array (should create it)
	builder.AddToArray("items", "first")
	builder.AddToArray("items", "second")

	json := builder.Build()

	if json.GetByPath("items").ArraySize() != 2 {
		t.Errorf("Expected array size 2, got %d", json.GetByPath("items").ArraySize())
	}

	if json.GetByPath("items").ArrayElement(0).AsStringDefault("") != "first" {
		t.Error("First array element not correct")
	}

	if json.GetByPath("items").ArrayElement(1).AsStringDefault("") != "second" {
		t.Error("Second array element not correct")
	}
}

// Benchmark tests to compare old vs new approaches
func BenchmarkOldApproach(b *testing.B) {
	for i := 0; i < b.N; i++ {
		json := NewJSONObject()
		json.SetByPath("ip_address", NewJSON("192.168.1.1"))
		json.SetByPath("port", NewJSON(8080))
		json.SetByPath("enabled", NewJSON(true))
		json.SetByPath("name", NewJSON("test"))
		_ = json.ToString()
	}
}

func BenchmarkNewApproachBulk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		data := map[string]interface{}{
			"ip_address": "192.168.1.1",
			"port":       8080,
			"enabled":    true,
			"name":       "test",
		}
		json := NewJSONObjectFromMap(data)
		_ = json.ToString()
	}
}

func BenchmarkNewApproachBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		json := NewJSONBuilder().
			Set("ip_address", "192.168.1.1").
			Set("port", 8080).
			Set("enabled", true).
			Set("name", "test").
			Build()
		_ = json.ToString()
	}
}
