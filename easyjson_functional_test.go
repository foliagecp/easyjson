package easyjson

import (
	"encoding/hex"
	"reflect"
	"strconv"
	"testing"
)

func TestNewJSON_Basics(t *testing.T) {
	j := NewJSONObject()
	if !j.IsObject() {
		t.Fatalf("expected object")
	}
	arr := NewJSONArray()
	if !arr.IsArray() {
		t.Fatalf("expected array")
	}
	null := NewJSONNull()
	if !null.IsNull() {
		t.Fatalf("expected null")
	}

	jArr := NewJSON([]int{1, 2, 3})
	if !jArr.IsArray() {
		t.Fatalf("expected array from []int")
	}
	if jArr.ArraySize() != 3 {
		t.Fatalf("expected size 3, got %d", jArr.ArraySize())
	}
}

func TestPathExistsAndGetByPath_Object(t *testing.T) {
	j := NewJSONObject()
	ok := j.SetByPath("a.b.c", NewJSON(123))
	if !ok {
		t.Fatalf("SetByPath failed")
	}

	if !j.PathExists("a.b.c") {
		t.Fatalf("PathExists a.b.c should be true")
	}
	if j.PathExists("a.b.x") {
		t.Fatalf("PathExists a.b.x should be false")
	}

	got := j.GetByPath("a.b.c")
	num, ok := got.AsNumeric()
	if !ok || num != 123 {
		t.Fatalf("expected 123, got %v (ok=%v)", got.Value, ok)
	}
}

func TestPath_CustomDelimiter(t *testing.T) {
	j := NewJSONObject()
	if !j.SetByPathCustomDelimiter("foo/bar/baz", NewJSON("ok"), "/") {
		t.Fatalf("SetByPathCustomDelimiter failed")
	}
	if !j.PathExists("foo/bar/baz", "/") {
		t.Fatalf("PathExists with custom delimiter failed")
	}
	if s, ok := j.GetByPath("foo/bar/baz", "/").AsString(); !ok || s != "ok" {
		t.Fatalf("expected ok at foo/bar/baz, got %q", s)
	}
}

// SetByPath never creates arrays implicitly: a numeric path segment under a
// missing key produces an object with that key, not an array (see
// TestSetByPath_NumericSegmentDoesNotCreateArray). Arrays must be created
// explicitly, after which numeric indexing and push work as expected.
func TestSetByPath_ArrayIndexingAndPush(t *testing.T) {
	j := NewJSONObject()

	if !j.SetByPath("arr", NewJSONArray()) {
		t.Fatalf("SetByPath arr = [] failed")
	}
	if !j.SetByPath("arr.2", NewJSON("bar")) {
		t.Fatalf("SetByPath arr.2 failed")
	}
	if sz := j.GetByPath("arr").ArraySize(); sz != 3 {
		t.Fatalf("expected size 3 after setting index 2, got %d", sz)
	}
	if s, ok := j.GetByPath("arr.2").AsString(); !ok || s != "bar" {
		t.Fatalf("expected bar at arr.2, got %q", s)
	}

	if !j.SetByPath("arr.-1", NewJSON("tail")) {
		t.Fatalf("SetByPath arr.-1 failed")
	}
	if sz := j.GetByPath("arr").ArraySize(); sz != 4 {
		t.Fatalf("expected size 4 after push, got %d", sz)
	}
	if s, ok := j.GetByPath("arr.3").AsString(); !ok || s != "tail" {
		t.Fatalf("expected tail at arr.3, got %q", s)
	}

	if !j.SetByPath("arr.0", NewJSONObject()) {
		t.Fatalf("SetByPath arr.0 = {} failed")
	}
	if !j.SetByPath("arr.0.x", NewJSON("y")) {
		t.Fatalf("SetByPath arr.0.x failed")
	}
	if s, ok := j.GetByPath("arr.0.x").AsString(); !ok || s != "y" {
		t.Fatalf("expected y at arr.0.x, got %q", s)
	}
}

// Numeric path segments are plain object keys, not array indices. Auto-creating
// arrays from them used to blow up on paths that merely look numeric (e.g. SNMP
// OIDs like "1:3:6:1:2:1:17:1:4:1:2.3"), producing huge sparse arrays.
func TestSetByPath_NumericSegmentDoesNotCreateArray(t *testing.T) {
	j := NewJSONObject()
	if !j.SetByPath("arr.2", NewJSON("bar")) {
		t.Fatalf("SetByPath arr.2 failed")
	}

	arr := j.GetByPath("arr")
	if arr.IsArray() {
		t.Fatalf("arr must not be auto-created as array")
	}
	if !arr.IsObject() {
		t.Fatalf("arr should be object, got %T", arr.Value)
	}
	if s, ok := j.GetByPath("arr.2").AsString(); !ok || s != "bar" {
		t.Fatalf("expected bar at arr.2, got %q", s)
	}
	if keys := arr.ObjectKeys(); len(keys) != 1 || keys[0] != "2" {
		t.Fatalf("expected single key \"2\", got %v", keys)
	}
}

func TestRemoveByPath_Object(t *testing.T) {
	j := NewJSONObject()
	j.SetByPath("a.b.c", NewJSON(1))
	j.SetByPath("a.b.d", NewJSON(2))
	// Numeric segments are object keys, so this is an object, not an array.
	j.SetByPath("numkeys.0", NewJSON("x"))
	j.SetByPath("numkeys.1", NewJSON("y"))

	if !j.RemoveByPath("a.b.c") {
		t.Fatalf("RemoveByPath a.b.c failed")
	}
	if j.PathExists("a.b.c") {
		t.Fatalf("a.b.c should be removed")
	}
	if j.GetByPath("a.b.d").AsNumericDefault(-1) != 2 {
		t.Fatalf("sibling a.b.d should survive removal")
	}

	if !j.RemoveByPath("numkeys.0") {
		t.Fatalf("RemoveByPath numkeys.0 failed")
	}
	if j.PathExists("numkeys.0") {
		t.Fatalf("numkeys.0 should be removed")
	}
	if s, ok := j.GetByPath("numkeys.1").AsString(); !ok || s != "y" {
		t.Fatalf("expected y at numkeys.1, got %q", s)
	}
}

// Removing an array element clears the slot in place — the array keeps its
// length instead of shifting the tail, so indices of the remaining elements
// stay stable.
func TestRemoveByPath_Array(t *testing.T) {
	j := NewJSONObject()
	j.SetByPath("arr", NewJSON([]string{"x", "y", "z"}))

	if !j.RemoveByPath("arr.0") {
		t.Fatalf("RemoveByPath arr.0 failed")
	}
	if sz := j.GetByPath("arr").ArraySize(); sz != 3 {
		t.Fatalf("array must keep its length after remove, got %d", sz)
	}
	if v := j.GetByPath("arr.0"); !v.IsNull() {
		t.Fatalf("expected nil at arr.0 after remove, got %v", v.Value)
	}
	if s, ok := j.GetByPath("arr.1").AsString(); !ok || s != "y" {
		t.Fatalf("expected y at arr.1, got %q", s)
	}
	if s, ok := j.GetByPath("arr.2").AsString(); !ok || s != "z" {
		t.Fatalf("expected z at arr.2, got %q", s)
	}

	// Removal inside an element keeps the element itself.
	j.SetByPath("arr.2", NewJSONObject())
	j.SetByPath("arr.2.k", NewJSON("v"))
	if !j.RemoveByPath("arr.2.k") {
		t.Fatalf("RemoveByPath arr.2.k failed")
	}
	if j.PathExists("arr.2.k") {
		t.Fatalf("arr.2.k should be removed")
	}
	if !j.GetByPath("arr.2").IsObject() {
		t.Fatalf("arr.2 should still be an object")
	}

	// Out-of-range and non-numeric indices are rejected.
	if j.RemoveByPath("arr.9") {
		t.Fatalf("RemoveByPath on out-of-range index should fail")
	}
	if j.RemoveByPath("arr.foo") {
		t.Fatalf("RemoveByPath on non-numeric index should fail")
	}
	if sz := j.GetByPath("arr").ArraySize(); sz != 3 {
		t.Fatalf("failed removals must not modify the array, got size %d", sz)
	}
}

func TestDeepMerge(t *testing.T) {
	a := NewJSONObject()
	a.SetByPath("obj.x", NewJSON(1))
	a.SetByPath("arr", NewJSON([]int{1, 2}))

	b := NewJSONObject()
	b.SetByPath("obj.y", NewJSON(2))
	b.SetByPath("arr", NewJSON([]int{2, 3}))

	a.DeepMerge(b)

	if n := a.GetByPath("obj.x").AsNumericDefault(-1); n != 1 {
		t.Fatalf("obj.x expected 1, got %v", n)
	}
	if n := a.GetByPath("obj.y").AsNumericDefault(-1); n != 2 {
		t.Fatalf("obj.y expected 2, got %v", n)
	}

	arr := a.GetByPath("arr")
	if !arr.IsArray() {
		t.Fatalf("arr should be array, got %T", arr.Value)
	}
	if arr.ArraySize() != 3 {
		t.Fatalf("merged array length mismatch: want 3, got %d", arr.ArraySize())
	}
	got := []float64{
		arr.GetByPath("0").AsNumericDefault(-1),
		arr.GetByPath("1").AsNumericDefault(-1),
		arr.GetByPath("2").AsNumericDefault(-1),
	}
	want := []float64{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merged array mismatch\nwant: %#v\ngot : %#v", want, got)
	}
}

func TestEqualsAndClone(t *testing.T) {
	j := NewJSONObject()
	j.SetByPath("a.b", NewJSON("c"))
	j.SetByPath("n", NewJSON(10))

	cl := j.Clone()
	if !j.Equals(cl) {
		t.Fatalf("clone should be equal initially")
	}

	cl.SetByPath("a.b", NewJSON("changed"))
	if j.Equals(cl) {
		t.Fatalf("after change, clone must differ from original")
	}
	if s, _ := j.GetByPath("a.b").AsString(); s != "c" {
		t.Fatalf("original should keep 'c', got %q", s)
	}
}

func TestJSONFromStringAndBytes(t *testing.T) {
	src := `{"a": {"b": 1}, "arr":[1,2,3]}`
	j, ok := JSONFromString(src)
	if !ok {
		t.Fatalf("JSONFromString failed")
	}
	if j.GetByPath("a.b").AsNumericDefault(-1) != 1 {
		t.Fatalf("expected a.b == 1")
	}
	if j.ArraySize() != -1 {
		t.Fatalf("root not an array")
	}

	b := j.ToBytes()
	j2, ok := JSONFromBytes(b)
	if !ok || !j.Equals(j2) {
		t.Fatalf("JSONFromBytes/ToBytes roundtrip failed")
	}
}

func TestNewJSONBytes_HexRoundtrip(t *testing.T) {
	raw := []byte{0xde, 0xad, 0xbe, 0xef}
	j := NewJSONBytes(raw)
	got, ok := j.AsBytes()
	if !ok {
		t.Fatalf("AsBytes failed")
	}
	if !reflect.DeepEqual(got, raw) {
		t.Fatalf("hex roundtrip mismatch: %s vs %s", hex.EncodeToString(got), hex.EncodeToString(raw))
	}
}

func TestAsConversions(t *testing.T) {
	jNum := NewJSON(3.14)
	if v, ok := jNum.AsNumeric(); !ok || v != 3.14 {
		t.Fatalf("AsNumeric failed: %v, %v", v, ok)
	}
	if jNum.AsStringDefault("x") != "x" {
		t.Fatalf("AsStringDefault should return default for non-string")
	}

	jStr := NewJSON("hi")
	if s, ok := jStr.AsString(); !ok || s != "hi" {
		t.Fatalf("AsString failed")
	}
	if !NewJSON(true).AsBoolDefault(false) {
		t.Fatalf("AsBoolDefault failed")
	}
}

func TestObjectKeysAndCounts(t *testing.T) {
	j := NewJSONObject()
	j.SetByPath("x", NewJSON(1))
	j.SetByPath("y", NewJSON(2))
	keys := j.ObjectKeys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d: %v", len(keys), keys)
	}
	if j.KeysCount() != 2 || !j.IsNonEmptyObject() {
		t.Fatalf("KeysCount / IsNonEmptyObject mismatch")
	}
}

func TestArrayHelpers(t *testing.T) {
	arr := NewJSONArray()
	arr.AddToArray(NewJSON("a"))
	arr.AddToArray(NewJSON("b"))
	if arr.ArraySize() != 2 {
		t.Fatalf("ArraySize expected 2, got %d", arr.ArraySize())
	}
	if arr.ArrayElement(1).AsStringDefault("") != "b" {
		t.Fatalf("ArrayElement(1) expected 'b'")
	}
}

func TestSetByPaths(t *testing.T) {
	j := NewJSONObject()
	j.SetByPaths(map[string]interface{}{
		"a.b": 1,
		"x.y": "z",
	})
	if j.GetByPath("a.b").AsNumericDefault(-1) != 1 {
		t.Fatalf("a.b expected 1")
	}
	if j.GetByPath("x.y").AsStringDefault("") != "z" {
		t.Fatalf("x.y expected z")
	}
}

func TestNewJSONObjectFromMapAndBuildFromTemplate(t *testing.T) {
	data := map[string]interface{}{
		"a.b": 1,
		"x.y": "z",
	}
	j := NewJSONObjectFromMap(data)
	if j.GetByPath("a.b").AsNumericDefault(-1) != 1 ||
		j.GetByPath("x.y").AsStringDefault("") != "z" {
		t.Fatalf("NewJSONObjectFromMap failed")
	}
	tpl := map[string]interface{}{
		"k.v": 42,
	}
	j2 := BuildFromTemplate(tpl)
	if j2.GetByPath("k.v").AsNumericDefault(0) != 42 {
		t.Fatalf("BuildFromTemplate failed")
	}
}

func TestJSONBuilder(t *testing.T) {
	b := NewJSONBuilder().
		Set("a.b", 1).
		SetIfNotEmpty("skip.empty", "").
		AddToArray("arr", "x").
		AddToArray("arr", "y")

	j := b.Build()
	if j.PathExists("skip.empty") {
		t.Fatalf("SetIfNotEmpty should skip empty values")
	}
	if j.GetByPath("a.b").AsNumericDefault(-1) != 1 {
		t.Fatalf("a.b expected 1")
	}
	if j.GetByPath("arr").ArraySize() != 2 {
		t.Fatalf("arr size expected 2")
	}
}

func TestBuildArrayFromSlice(t *testing.T) {
	type item struct{ ID int }
	js := BuildArrayFromSlice([]item{{1}, {2}, {3}}, func(it item) map[string]interface{} {
		return map[string]interface{}{"id": it.ID}
	})
	if js.ArraySize() != 3 {
		t.Fatalf("expected array of 3")
	}
	for i := 0; i < 3; i++ {
		got := js.GetByPath(strconv.Itoa(i) + ".id").AsNumericDefault(-1)
		if int(got) != i+1 {
			t.Fatalf("item %d id expected %d, got %v", i, i+1, got)
		}
	}
}

func TestAsInt64(t *testing.T) {
	// int64 value that loses precision through float64
	big := int64(1<<53 + 1) // 9007199254740993
	j := NewJSON(big)

	got, ok := j.AsInt64()
	if !ok {
		t.Fatalf("AsInt64 returned ok=false for int64 value")
	}
	if got != big {
		t.Fatalf("AsInt64 precision loss: expected %d, got %d", big, got)
	}

	// AsNumeric loses precision for this value
	f, _ := j.AsNumeric()
	if int64(f) == big {
		t.Fatalf("expected float64 to lose precision for 2^53+1, but it didn't")
	}

	// Test with other numeric types
	tests := []struct {
		name string
		val  interface{}
		want int64
	}{
		{"int", int(42), 42},
		{"uint", uint(100), 100},
		{"uint32", uint32(200), 200},
		{"uint64", uint64(300), 300},
		{"byte", byte(7), 7},
		{"float64", float64(99), 99},
		{"float32", float32(50), 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := NewJSON(tt.val)
			got, ok := j.AsInt64()
			if !ok {
				t.Fatalf("AsInt64 returned ok=false for %T", tt.val)
			}
			if got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}

	// Non-numeric should return false
	if _, ok := NewJSON("hello").AsInt64(); ok {
		t.Fatalf("AsInt64 should return false for string")
	}
}

func TestAsInt64Default(t *testing.T) {
	j := NewJSON(int64(42))
	if got := j.AsInt64Default(0); got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
	if got := NewJSON("x").AsInt64Default(-1); got != -1 {
		t.Fatalf("expected default -1, got %d", got)
	}
}

func TestAsInt64_SetByPath(t *testing.T) {
	big := int64(1<<53 + 1)
	j := NewJSONObject()
	j.SetByPath("op.time", NewJSON(big))

	got, ok := j.GetByPath("op.time").AsInt64()
	if !ok {
		t.Fatalf("AsInt64 returned false after SetByPath")
	}
	if got != big {
		t.Fatalf("precision loss after SetByPath: expected %d, got %d", big, got)
	}
}
