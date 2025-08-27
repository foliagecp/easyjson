package easyjson

import (
	"reflect"
	"runtime"
	"strconv"
	"testing"
)

// Sinks to prevent compiler from eliminating results.
var (
	sinkJSON JSON
	sinkAny  interface{}
	sinkB    []byte
)

// ------------------------------------
// Helper path sets (compatible with current SetByPath semantics)
// IMPORTANT: short and long sets use DISTINCT roots to avoid prefix conflicts.
// ------------------------------------

// Object-only paths (no arrays), DISTINCT roots
func objPathsShort() []string {
	return []string{
		"o1.a.b.c",
		"o2.m.n.k",
		"o3.user.name.first",
		"o4.cfg.env.dev",
		"o5.x.y.z",
	}
}

func objPathsLong() []string {
	return []string{
		"L1.a.b.c.d.e",
		"L2.root.config.service.api.v1",
		"L3.u.profile.contacts.phone.mobile",
		"L4.metrics.prom.exporter.rules.static",
		"L5.deep.n1.n2.n3.n4.n5",
	}
}

// Array-LEAF paths only (array index is the last token).
// DISTINCT roots from object paths to avoid any overlap.
func arrLeafPathsShort() []string {
	return []string{
		"arrS.2",
		"listS.0",
		"itemsS.3",
		"aS.5",
		"numsS.1",
	}
}

func arrLeafPathsLong() []string {
	return []string{
		"arrL.10",
		"chainL.12",
		"aL.7",
		"mL.15",
		"complexL.20",
	}
}

// ------------------------------------
// Shared fixture helpers
// ------------------------------------

// Builds a JSON and sets each path to a numeric value (1..N).
// Fails the provided testing TB (works for both *testing.T and *testing.B) on any error.
func buildJSONWithPathsMustTB(tb testing.TB, paths []string) JSON {
	tb.Helper()
	j := NewJSONObject()
	for i, p := range paths {
		if ok := j.SetByPath(p, NewJSON(i+1)); !ok {
			tb.Fatalf("failed to SetByPath for fixture, path=%q (index=%d)", p, i)
		}
	}
	// Optional sanity of the fixture
	for i, p := range paths {
		if !j.PathExists(p) {
			tb.Fatalf("fixture path missing: %s", p)
		}
		if got := j.GetByPath(p).AsNumericDefault(-1); int(got) != i+1 {
			tb.Fatalf("fixture value mismatch at %s: want=%d got=%v (type=%T)",
				p, i+1, j.GetByPath(p).Value, j.GetByPath(p).Value)
		}
	}
	return j
}

// ------------------------------------
// Benchmarks: PathExists / GetByPath
// ------------------------------------

func BenchmarkPathExists(b *testing.B) {
	// Build fixture with object-only + array-leaf paths (no conflicting prefixes)
	all := append(append([]string{}, objPathsShort()...), objPathsLong()...)
	all = append(all, arrLeafPathsShort()...)
	all = append(all, arrLeafPathsLong()...)

	base := buildJSONWithPathsMustTB(b, all)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := all[i%len(all)]
		if !base.PathExists(p) {
			b.Fatalf("path not found: %s", p)
		}
	}
}

func BenchmarkGetByPath_ShortObj(b *testing.B) {
	paths := objPathsShort()
	base := buildJSONWithPathsMustTB(b, paths)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := paths[i%len(paths)]
		sinkJSON = base.GetByPath(p)
	}
}

func BenchmarkGetByPath_LongArrLeaf(b *testing.B) {
	paths := arrLeafPathsLong()
	base := buildJSONWithPathsMustTB(b, paths)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := paths[i%len(paths)]
		sinkJSON = base.GetByPath(p)
	}
}

// ------------------------------------
// Benchmarks: SetByPath (update existing)
// ------------------------------------

func BenchmarkSetByPath_UpdateExisting(b *testing.B) {
	j := NewJSONObject()
	path := "oUp.a.b.c.d.e"
	_ = j.SetByPath(path, NewJSON(0))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = j.SetByPath(path, NewJSON(i))
	}
}

// ------------------------------------
// Benchmarks: SetByPath (create new)
// Use only object-only paths and array-LEAF paths.
// Each iteration uses a fresh JSON to avoid unbounded growth.
// ------------------------------------

func BenchmarkSetByPath_CreateNew_Mixed(b *testing.B) {
	paths := []string{
		"oC.a.b.c",       // object
		"arrC.2",         // array (leaf)
		"deepC.v1.v2.v3", // object
		"cfgC.env.dev",   // object
		"listC.5",        // array (leaf)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		j := NewJSONObject()
		for k, p := range paths {
			if ok := j.SetByPath(p, NewJSON(k+i)); !ok {
				b.Fatalf("SetByPath failed for %q", p)
			}
		}
		sinkJSON = j
	}
}

// ------------------------------------
// Benchmarks: RemoveByPath
// Remove existing path then restore it so each iteration is identical.
// ------------------------------------

func BenchmarkRemoveByPath_ObjectKey(b *testing.B) {
	j := NewJSONObject()
	path := "oR.a.b.c"
	_ = j.SetByPath(path, NewJSON(123))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if !j.RemoveByPath(path) {
			b.Fatal("remove failed")
		}
		_ = j.SetByPath(path, NewJSON(123)) // restore
	}
}

func BenchmarkRemoveByPath_ArrayIndex(b *testing.B) {
	j := NewJSONObject()
	path := "arrR.3" // array (leaf)
	_ = j.SetByPath(path, NewJSON("x"))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if !j.RemoveByPath(path) {
			b.Fatal("remove failed")
		}
		_ = j.SetByPath(path, NewJSON("x")) // restore
	}
}

// ------------------------------------
// Benchmarks: DeepMerge (mid-size objects)
// Cloning is excluded from measured time.
// ------------------------------------

func BenchmarkDeepMerge_MidSize(b *testing.B) {
	left := NewJSONObject()
	right := NewJSONObject()

	// Fill objects
	for i := 0; i < 50; i++ {
		_ = left.SetByPath("objL.l"+strconv.Itoa(i), NewJSON(i))
		_ = right.SetByPath("objR.r"+strconv.Itoa(i), NewJSON(i))
	}

	// Fill arrays (leaf indices only) with duplicates and some new elements
	for i := 0; i < 50; i++ {
		_ = left.SetByPath("arrLM."+strconv.Itoa(i), NewJSON(i%10))
	}
	for i := 0; i < 50; i++ {
		_ = right.SetByPath("arrLM."+strconv.Itoa(i), NewJSON(i%10))
	}
	_ = right.SetByPath("arrLM.55", NewJSON(999)) // new element

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		x := left.Clone() // exclude clone from the measured section
		b.StartTimer()

		x.DeepMerge(right)
		sinkJSON = x
	}
}

// ------------------------------------
// Benchmarks: Serialization / Deserialization
// ------------------------------------

func BenchmarkToBytes(b *testing.B) {
	all := append(append([]string{}, objPathsLong()...), arrLeafPathsLong()...)
	base := buildJSONWithPathsMustTB(b, all)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sinkB = base.ToBytes()
	}
}

func BenchmarkJSONFromBytes(b *testing.B) {
	all := append(append([]string{}, objPathsLong()...), arrLeafPathsLong()...)
	base := buildJSONWithPathsMustTB(b, all)
	data := base.ToBytes()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		js, ok := JSONFromBytes(data)
		if !ok {
			b.Fatal("unmarshal failed")
		}
		sinkJSON = js
	}
}

// ------------------------------------
// Benchmarks: Clone (deep copy)
// ------------------------------------

func BenchmarkClone(b *testing.B) {
	all := append(append([]string{}, objPathsLong()...), arrLeafPathsLong()...)
	base := buildJSONWithPathsMustTB(b, all)

	// Add some extra nested structures (object-only)
	_ = base.SetByPath("deepC2.v1.v2.v3", NewJSON(42))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sinkJSON = base.Clone()
	}
}

// ------------------------------------
// Sanity for fixtures used by benches
// ------------------------------------

func Test_buildJSONWithPaths_sanity(t *testing.T) {
	all := append(append([]string{}, objPathsLong()...), arrLeafPathsLong()...)
	j := buildJSONWithPathsMustTB(t, all)
	for i, p := range all {
		if !j.PathExists(p) {
			t.Fatalf("path missing: %s", p)
		}
		if got := j.GetByPath(p).AsNumericDefault(-1); int(got) != i+1 {
			t.Fatalf("value mismatch at %s: want=%d got=%v (type=%T)",
				p, i+1, j.GetByPath(p).Value, j.GetByPath(p).Value)
		}
	}
}

// --- helpers: semantic JSON equality (numbers normalized) ---

func toFloat64(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case int32:
		return float64(x), true
	case int16:
		return float64(x), true
	case int8:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint64:
		return float64(x), true
	case uint32:
		return float64(x), true
	case uint16:
		return float64(x), true
	case uint8:
		return float64(x), true
	default:
		return 0, false
	}
}

func jsonSemanticallyEqual(a, b interface{}) bool {
	// numbers: compare by value regardless of concrete numeric type
	if av, ok := toFloat64(a); ok {
		if bv, ok2 := toFloat64(b); ok2 {
			return av == bv
		}
		// number vs non-number
		return false
	}

	switch ax := a.(type) {
	case map[string]interface{}:
		bx, ok := b.(map[string]interface{})
		if !ok || len(ax) != len(bx) {
			return false
		}
		for k, va := range ax {
			vb, ok := bx[k]
			if !ok || !jsonSemanticallyEqual(va, vb) {
				return false
			}
		}
		return true
	case []interface{}:
		bx, ok := b.([]interface{})
		if !ok || len(ax) != len(bx) {
			return false
		}
		for i := range ax {
			if !jsonSemanticallyEqual(ax[i], bx[i]) {
				return false
			}
		}
		return true
	default:
		// strings, bools, nil, etc.
		return reflect.DeepEqual(a, b)
	}
}

// Non-benchmark sanity: verify ToBytes/FromBytes preserves container shapes.
func TestRoundTripContainers(t *testing.T) {
	j := NewJSONObject()
	_ = j.SetByPath("oRT.a.b.c", NewJSON(1)) // int
	_ = j.SetByPath("arrRT.2", NewJSON("v")) // array with holes
	data := j.ToBytes()
	j2, ok := JSONFromBytes(data)
	if !ok {
		t.Fatal("roundtrip unmarshal failed")
	}
	if !jsonSemanticallyEqual(j.Value, j2.Value) {
		t.Fatalf("roundtrip mismatch (semantic)\norig: %#v\nback: %#v", j.Value, j2.Value)
	}
}

// Keep sinks alive for the linker/GC.
func init() { runtime.KeepAlive(sinkAny) }
