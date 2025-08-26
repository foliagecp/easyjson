package easyjson

import (
	"testing"
)

func TestSetByPaths(t *testing.T) {
	json := NewJSONObject()

	data := map[string]interface{}{
		"ip_address": "192.168.1.1",
		"port":       8080,
		"enabled":    true,
	}

	json.SetByPaths(data)

	if json.GetByPath("ip_address").AsStringDefault("") != "192.168.1.1" {
		t.Error("ip_address not set correctly")
	}

	if json.GetByPath("port").AsNumericDefault(0) != 8080 {
		t.Error("port not set correctly")
	}

	if !json.GetByPath("enabled").AsBoolDefault(false) {
		t.Error("enabled not set correctly")
	}
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

func TestJSONBuilder(t *testing.T) {
	json := NewJSONBuilder().
		Set("ip_address", "192.168.1.1").
		Set("port", 8080).
		SetIfNotEmpty("empty", "").
		SetIfNotEmpty("nonempty", "value").
		Build()

	if json.GetByPath("ip_address").AsStringDefault("") != "192.168.1.1" {
		t.Error("ip_address not set correctly")
	}

	if json.GetByPath("port").AsNumericDefault(0) != 8080 {
		t.Error("port not set correctly")
	}

	if json.PathExists("empty") {
		t.Error("empty value should not be set")
	}

	if json.GetByPath("nonempty").AsStringDefault("") != "value" {
		t.Error("nonempty value not set correctly")
	}
}

func TestBuildArrayFromSlice(t *testing.T) {
	type TestItem struct {
		Name  string
		Value int
	}

	items := []TestItem{
		{Name: "first", Value: 1},
		{Name: "second", Value: 2},
	}

	jsonArray := BuildArrayFromSlice(items, func(item TestItem) map[string]interface{} {
		return map[string]interface{}{
			"name":  item.Name,
			"value": item.Value,
		}
	})

	if jsonArray.ArraySize() != 2 {
		t.Errorf("Expected array size 2, got %d", jsonArray.ArraySize())
	}

	first := jsonArray.ArrayElement(0)
	if first.GetByPath("name").AsStringDefault("") != "first" {
		t.Error("First item name not correct")
	}

	if first.GetByPath("value").AsNumericDefault(0) != 1 {
		t.Error("First item value not correct")
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

func BenchmarkCloneFast(b *testing.B) {
	testData := NewJSONObject()

	testData.SetByPath("name", NewJSON("test-service"))
	testData.SetByPath("version", NewJSON("1.2.3"))
	testData.SetByPath("enabled", NewJSON(true))

	configMap := map[string]interface{}{
		"host":    "localhost",
		"port":    8080,
		"timeout": 30,
		"retries": 3,
		"ssl":     true,
	}
	testData.SetByPath("config", NewJSON(configMap))

	usersMap := map[string]interface{}{
		"admin": map[string]interface{}{
			"password":    "secret123",
			"permissions": []interface{}{"read", "write", "admin"},
			"active":      true,
		},
		"user1": map[string]interface{}{
			"password":    "pass456",
			"permissions": []interface{}{"read"},
			"active":      false,
		},
		"user2": map[string]interface{}{
			"password":    "pass789",
			"permissions": []interface{}{"read", "write"},
			"active":      true,
		},
	}
	testData.SetByPath("users", NewJSON(usersMap))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cloned := testData.Clone()
		_ = cloned
	}
}

/*
➜  easyjson git:(bench-clone) go test -cpu=1,2,4,8 -bench=BenchmarkClone -benchtime=1000000x -benchmem
goos: linux
goarch: amd64
pkg: github.com/foliagecp/easyjson
cpu: AMD Ryzen 5 5600X 6-Core Processor
BenchmarkCloneFast       1000000             11224 ns/op            5768 B/op        130 allocs/op
BenchmarkCloneFast-2     1000000              9449 ns/op            5768 B/op        130 allocs/op
BenchmarkCloneFast-4     1000000              9908 ns/op            5770 B/op        130 allocs/op
BenchmarkCloneFast-8     1000000              9903 ns/op            5771 B/op        130 allocs/op
PASS
ok      github.com/foliagecp/easyjson   40.490s
*/
