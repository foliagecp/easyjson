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

func TestSetByPath_ArrayIndexingAndPush(t *testing.T) {
	j := NewJSONObject()

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

func TestRemoveByPath_ObjectAndArray(t *testing.T) {
	j := NewJSONObject()
	j.SetByPath("a.b.c", NewJSON(1))
	j.SetByPath("a.b.d", NewJSON(2))
	j.SetByPath("arr.0", NewJSON("x"))
	j.SetByPath("arr.1", NewJSON("y"))

	if !j.RemoveByPath("a.b.c") {
		t.Fatalf("RemoveByPath a.b.c failed")
	}
	if j.PathExists("a.b.c") {
		t.Fatalf("a.b.c should be removed")
	}

	if !j.RemoveByPath("arr.0") {
		t.Fatalf("RemoveByPath arr.0 failed")
	}
	if v := j.GetByPath("arr.0"); !v.IsNull() {
		t.Fatalf("expected nil at arr.0 after remove, got %v", v.Value)
	}
	if s, ok := j.GetByPath("arr.1").AsString(); !ok || s != "y" {
		t.Fatalf("expected y at arr.1, got %q", s)
	}
}

func TestDeepMerge(t *testing.T) {
	a := NewJSONObject()
	a.SetByPath("obj.x", NewJSON(1))
	a.SetByPath("arr.0", NewJSON(1))
	a.SetByPath("arr.1", NewJSON(2))

	b := NewJSONObject()
	b.SetByPath("obj.y", NewJSON(2))
	b.SetByPath("arr.0", NewJSON(2))
	b.SetByPath("arr.2", NewJSON(3))

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
