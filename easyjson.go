// Package easyjson provides everything that is needed for ease and simple operating with JSON data structure.
// It offers a fluent API for JSON manipulation, path-based access, and builder patterns.
package easyjson

import (
	"encoding/hex"
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// JSON represents a JSON value that can be manipulated using path-based operations.
type JSON struct {
	Value interface{}
}

// NewJSON creates a new JSON instance from any Go value.
func NewJSON(value interface{}) JSON {
	if value == nil {
		return JSON{Value: nil}
	}

	val := reflect.ValueOf(value)
	typ := reflect.TypeOf(value)
	if typ.Kind() == reflect.Slice && typ.Elem().Kind() != reflect.Interface {
		length := val.Len()
		result := make([]interface{}, length)
		for i := 0; i < length; i++ {
			result[i] = val.Index(i).Interface()
		}
		return JSON{Value: result}
	}

	return JSON{Value: value}
}

// NewJSONBytes creates a new JSON instance from byte slice, encoding it as hex string.
func NewJSONBytes(value []byte) JSON {
	codedBytes := hex.EncodeToString(value)
	var j JSON
	j.Value = codedBytes
	return j
}

// NewJSONNull creates a new JSON instance with null value.
func NewJSONNull() JSON {
	return NewJSON(nil)
}

// NewJSONObject creates a new empty JSON object.
func NewJSONObject() JSON {
	return NewJSON(make(map[string]interface{}))
}

// NewJSONObjectWithKeyValue creates a new JSON object with a single key-value pair.
func NewJSONObjectWithKeyValue(key string, value JSON) JSON {
	m := make(map[string]interface{})
	m[key] = value.Value
	return NewJSON(m)
}

// NewJSONArray creates a new empty JSON array.
func NewJSONArray() JSON {
	return NewJSON(make([]interface{}, 0))
}

/*func (j1 JSON) Equals(j2 JSON) bool {
	if j1.IsObject() && j2.IsObject() {
		if len(j1.ObjectKeys()) == len(j2.ObjectKeys()) {
			for _, key := range j1.ObjectKeys() {
				if !j2.PathExists(key) {
					return false
				}
				if !j1.GetByPath(key).Equals(j2.GetByPath(key)) {
					return false
				}
			}
			return true
		}
		return false
	}
	if j1.IsArray() && j2.IsArray() {
		if j1.ArraySize() == j2.ArraySize() {
			for i := 0; i < j1.ArraySize(); i++ {
				element_found := false
				for j := 0; j < j2.ArraySize(); j++ {
					if j1.GetByPath(strconv.Itoa(i)).Equals(j2.GetByPath(strconv.Itoa(j))) {
						element_found = true
					}
				}
				if !element_found {
					return false
				}
			}
			return true
		}
		return false
	}
	if j1.Value == j2.Value {
		return true
	}
	return false
}*/

// Equals compares two JSON values for deep equality.
func (j1 JSON) Equals(j2 JSON) bool {
	return reflect.DeepEqual(j1.Value, j2.Value)
}

type pathIter struct {
	s     string
	i     int
	delim byte
}

func (it *pathIter) next() (tok string, ok bool) {
	if it.i >= len(it.s) {
		return "", false
	}
	j := it.i
	for j < len(it.s) && it.s[j] != it.delim {
		j++
	}
	tok = it.s[it.i:j]
	it.i = j
	if it.i < len(it.s) && it.s[it.i] == it.delim {
		it.i++
	}
	return tok, true
}

func (j JSON) PathExists(p string, delimiter ...string) bool {
	delim := byte('.')
	if len(delimiter) > 0 && len(delimiter[0]) > 0 {
		delim = delimiter[0][0]
	}
	if p == "" {
		return true
	}

	cur := j.Value
	it := pathIter{s: p, i: 0, delim: delim}
	for {
		tok, ok := it.next()
		if !ok || tok == "" {
			break
		}
		switch v := cur.(type) {
		case map[string]interface{}:
			nv, ok := v[tok]
			if !ok {
				return false
			}
			cur = nv
		case []interface{}:
			idx, err := strconv.Atoi(tok)
			if err != nil || idx < 0 || idx >= len(v) {
				return false
			}
			cur = v[idx]
		default:
			return false
		}
	}
	return true
}

func (j JSON) GetByPath(p string, delimiter ...string) JSON {
	delim := byte('.')
	if len(delimiter) > 0 && len(delimiter[0]) > 0 {
		delim = delimiter[0][0]
	}
	if p == "" {
		return j
	}

	cur := j.Value
	it := pathIter{s: p, i: 0, delim: delim}
	for {
		tok, ok := it.next()
		if !ok || tok == "" {
			break
		}
		switch v := cur.(type) {
		case map[string]interface{}:
			nv, ok := v[tok]
			if !ok {
				return NewJSONNull()
			}
			cur = nv
		case []interface{}:
			idx, err := strconv.Atoi(tok)
			if err != nil || idx < 0 || idx >= len(v) {
				return NewJSONNull()
			}
			cur = v[idx]
		default:
			return NewJSONNull()
		}
	}
	return NewJSON(cur)
}

// GetByPathPtr returns a pointer to the JSON value at the specified path.
func (j JSON) GetByPathPtr(p string) *JSON {
	res := j.GetByPath(p)
	return &res
}

// GetPtr returns a pointer to this JSON instance.
func (j JSON) GetPtr() *JSON {
	return &j
}

func nextPathToken(p string, i int, delim byte) (tok string, next int, ok bool) {
	if i >= len(p) {
		return "", i, false
	}
	j := i
	for j < len(p) && p[j] != delim {
		j++
	}
	tok = p[i:j]
	if j < len(p) && p[j] == delim {
		return tok, j + 1, true
	}
	return tok, j, true
}

func jvSetValueByPath(parent *interface{}, parentKeyOrIdForThisValue string, jv *interface{}, p string, v *interface{}, delimiter string) bool {
	delim := byte('.')
	if delimiter != "" {
		delim = delimiter[0]
	}

	// Recursive setter: walks the path and sets value, creating intermediate nodes.
	var set func(cur interface{}, pos int) (interface{}, bool)
	set = func(cur interface{}, pos int) (interface{}, bool) {
		tok, next, ok := nextPathToken(p, pos, delim)
		if !ok || tok == "" {
			return nil, false
		}
		last := next >= len(p)

		switch cv := cur.(type) {

		case map[string]interface{}:
			// Object case: always create missing children as objects (STRICT MODE).
			if last {
				cv[tok] = *v
				return cv, true
			}
			child, exists := cv[tok]
			if !exists || child == nil {
				// <<< FIX: no look-ahead to decide []interface{} by numeric token >>>
				cv[tok] = map[string]interface{}{}
				child = cv[tok]
			}
			newChild, ok := set(child, next)
			if !ok {
				return cur, false
			}
			cv[tok] = newChild
			return cv, true

		case []interface{}:
			// Array case: indices must be numeric; negative index means push (last token only).
			id, err := strconv.Atoi(tok)
			if err != nil {
				return cur, false
			}
			if last {
				if id < 0 {
					cv = append(cv, *v)
				} else {
					jvSetArrayValue(&cv, id, *v)
				}
				return cv, true
			}
			if id < 0 || id >= len(cv) {
				return cur, false
			}
			newChild, ok := set(cv[id], next)
			if !ok {
				return cur, false
			}
			cv[id] = newChild
			return cv, true

		default:
			// Neither object nor array — cannot traverse further.
			return cur, false
		}
	}

	newRoot, ok := set(*jv, 0)
	if !ok {
		return false
	}
	*jv = newRoot
	return true
}

func jvRemoveValueByPath(jv *interface{}, p string, delimiter string) bool {
	delim := byte('.')
	if delimiter != "" {
		delim = delimiter[0]
	}
	var rm func(cur *interface{}, idx int) bool
	rm = func(cur *interface{}, idx int) bool {
		tok, next, ok := nextPathToken(p, idx, delim)
		if !ok || tok == "" {
			*cur = nil
			return true
		}
		last := next >= len(p)
		switch cv := (*cur).(type) {
		case map[string]interface{}:
			if last {
				delete(cv, tok)
				return true
			}
			nxt, ok := cv[tok]
			if !ok {
				return true
			}
			return rm(&nxt, next)
		case []interface{}:
			id, err := strconv.Atoi(tok)
			if err != nil || id < 0 || id >= len(cv) {
				return false
			}
			if last {
				cv[id] = nil
				return true
			}
			return rm(&cv[id], next)
		default:
			return false
		}
	}
	return rm(jv, 0)
}

func jvDeepMerge(jv1 *interface{}, jv2 *interface{}) {
	switch x1 := (*jv1).(type) {
	case map[string]interface{}:
		if x2, ok := (*jv2).(map[string]interface{}); ok {
			for k2, v2 := range x2 {
				if v1, exists := x1[k2]; exists {
					jvDeepMerge(&v1, &v2)
					x1[k2] = v1
				} else {
					x1[k2] = v2
				}
			}
		}
	case []interface{}:
		if x2, ok := (*jv2).([]interface{}); ok {
			a1 := x1

			for _, v2 := range x2 {
				if v2 == nil {
					continue
				}
				dup := false
				for _, v1 := range a1 {
					if v1 == nil {
						continue
					}
					if reflect.DeepEqual(v1, v2) {
						dup = true
						break
					}
				}
				if !dup {
					a1 = append(a1, v2)
				}
			}

			*jv1 = a1
		}
	default:
		*jv1 = *jv2
	}
}

func jvAddValueToArray(jArray *[]interface{}, v interface{}) bool {
	*jArray = append(*jArray, v)
	return true
}

func jvSetArrayValue(jArray *[]interface{}, id int, v interface{}) bool {
	if id >= 0 && len(*jArray) >= 0 {
		for ok := true; ok; ok = (len(*jArray) <= id) {
			jvAddValueToArray(jArray, nil)
		}
		(*jArray)[id] = v
		return true
	}
	return false
}

// SetByPath sets a value at the specified path in the JSON structure.
// Creates intermediate objects/arrays as needed.
func (j *JSON) SetByPath(p string, v JSON, delimiter ...string) bool {
	delim := "."
	if len(delimiter) > 0 {
		delim = delimiter[0]
	}
	return jvSetValueByPath(nil, "", &j.Value, p, &v.Value, delim)
}

// SetByPathCustomDelimiter sets a value using a custom path delimiter.
func (j *JSON) SetByPathCustomDelimiter(p string, v JSON, delimiter string) bool {
	return jvSetValueByPath(nil, "", &j.Value, p, &v.Value, delimiter)
}

// DeepMerge merges another JSON value into this one recursively.
func (j *JSON) DeepMerge(v JSON) {
	if j == nil {
		j = &v
	} else {
		jvDeepMerge(&j.Value, &v.Value)
	}
}

// RemoveByPath removes a value at the specified path.
func (j *JSON) RemoveByPath(p string, delimiter ...string) bool {
	delim := "."
	if len(delimiter) > 0 {
		delim = delimiter[0]
	}
	return jvRemoveValueByPath(&j.Value, p, delim)
}

func jvValueToBytes(jv interface{}) []byte {
	bytes, _ := json.Marshal(jv)
	return bytes
}

func jvValueToString(jv interface{}) string {
	return string(jvValueToBytes(jv))
}

func deepCopy(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		m := make(map[string]interface{}, len(x))
		for k, vv := range x {
			m[k] = deepCopy(vv)
		}
		return m
	case []interface{}:
		s := make([]interface{}, len(x))
		for i, vv := range x {
			s[i] = deepCopy(vv)
		}
		return s
	default:
		return x
	}
}

// Normalize converts JSON to a canonical form:
// - all numbers -> float64
// - arrays are sorted deterministically by the canonical string of elements
// - objects are normalized recursively (for the string canon, keys are sorted)
func (j *JSON) Normalize() {
	j.Value = normalizeValue(j.Value)
}

// (optional) Returns a clone in canonical form.
func (j JSON) NormalizedClone() JSON {
	c := j.Clone()
	c.Normalize()
	return c
}

// normalizeValue — recursive normalization of a node.
func normalizeValue(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		m := make(map[string]interface{}, len(x))
		for k, vv := range x {
			m[k] = normalizeValue(vv)
		}
		return m

	case []interface{}:
		arr := make([]interface{}, 0, len(x))
		for _, e := range x {
			arr = append(arr, normalizeValue(e))
		}
		// deterministic array sort by each element's canonical string
		sort.SliceStable(arr, func(i, j int) bool {
			return canonicalString(arr[i]) < canonicalString(arr[j])
		})
		return arr

	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int8:
		return float64(x)
	case int16:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	case uint:
		return float64(x)
	case uint8:
		return float64(x)
	case uint16:
		return float64(x)
	case uint32:
		return float64(x)
	case uint64:
		return float64(x)
	case json.Number:
		if f, err := x.Float64(); err == nil {
			return f
		}
		// if parsing fails, keep it as a string
		return x.String()
	default:
		return x
	}
}

// canonicalString — deterministic JSON string used for comparison/sorting.
// Objects are serialized with sorted keys; arrays are assumed already normalized.
func canonicalString(v interface{}) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64, string:
		b, _ := json.Marshal(x)
		return string(b)
	case []interface{}:
		var b strings.Builder
		b.WriteByte('[')
		for i, e := range x {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(canonicalString(e))
		}
		b.WriteByte(']')
		return b.String()
	case map[string]interface{}:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		b.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			kq, _ := json.Marshal(k)
			b.Write(kq)
			b.WriteByte(':')
			b.WriteString(canonicalString(x[k]))
		}
		b.WriteByte('}')
		return b.String()
	default:
		// fallback: for uncommon types, rely on json.Marshal
		b, _ := json.Marshal(x)
		return string(b)
	}
}

// Clone creates a deep copy of the JSON value.
func (j JSON) Clone() JSON {
	return NewJSON(deepCopy(j.Value))
}

// ToBytes converts the JSON value to byte slice.
func (j JSON) ToBytes() []byte {
	return jvValueToBytes(j.Value)
}

// ToString converts the JSON value to string representation.
func (j JSON) ToString() string {
	return jvValueToString(j.Value)
}

// IsNull checks if the JSON value is null.
func (j JSON) IsNull() bool {
	return j.Value == nil
}

// IsObject checks if the JSON value is an object.
func (j JSON) IsObject() bool {
	_, ok := j.AsObject()
	return ok
}

// IsArray checks if the JSON value is an array.
func (j JSON) IsArray() bool {
	_, ok := j.AsArray()
	return ok
}

// IsNumeric checks if the JSON value is a number.
func (j JSON) IsNumeric() bool {
	_, ok := j.AsNumeric()
	return ok
}

// IsString checks if the JSON value is a string.
func (j JSON) IsString() bool {
	_, ok := j.AsString()
	return ok
}

// IsBool checks if the JSON value is a boolean.
func (j JSON) IsBool() bool {
	_, ok := j.AsBool()
	return ok
}

// AsObject returns the JSON value as a map if it's an object.
func (j JSON) AsObject() (map[string]interface{}, bool) {
	val, ok := j.Value.(map[string]interface{})
	return val, ok
}

// AsArray returns the JSON value as a slice if it's an array.
func (j JSON) AsArray() ([]interface{}, bool) {
	val, ok := j.Value.([]interface{})
	return val, ok
}

// AsArrayString returns the JSON array as a string slice if all elements are strings.
func (j JSON) AsArrayString() ([]string, bool) {
	array, ok := j.AsArray()
	if !ok {
		return nil, false
	}
	for _, v := range array {
		if _, ok := v.(string); !ok {
			return nil, false
		}
	}
	out := make([]string, len(array))
	for i, v := range array {
		out[i] = v.(string)
	}
	return out, true
}

// AddToArray adds an element to the JSON array.
func (j *JSON) AddToArray(jElem JSON) {
	if jv, ok := j.Value.([]interface{}); ok {
		jvAddValueToArray(&jv, jElem.Value)
		j.Value = jv
	}
}

// ArraySize returns the size of the JSON array, or -1 if not an array.
func (j JSON) ArraySize() int {
	if array, ok := j.AsArray(); ok {
		return len(array)
	}
	return -1
}

// ArrayElement returns the element at the specified index in the JSON array.
func (j JSON) ArrayElement(num int) JSON {
	size := j.ArraySize()
	if size >= 0 && num < size {
		array, _ := j.AsArray()
		return NewJSON(array[num])
	}
	return NewJSONNull()
}

// ObjectKeys returns all keys of the JSON object.
func (j JSON) ObjectKeys() []string {
	if object, ok := j.AsObject(); ok {
		var keys []string
		for key := range object {
			keys = append(keys, key)
		}
		return keys
	}
	return []string{}
}

// KeysCount returns the number of keys in the JSON object.
func (j JSON) KeysCount() int {
	return len(j.ObjectKeys())
}

// IsNonEmptyObject checks if the JSON value is a non-empty object.
func (j JSON) IsNonEmptyObject() bool {
	return len(j.ObjectKeys()) > 0
}

// IsNonEmptyArray checks if the JSON value is a non-empty array.
func (j JSON) IsNonEmptyArray() bool {
	return j.ArraySize() > 0
}

// AsNumeric returns the JSON value as a float64 if it's numeric.
func (j JSON) AsNumeric() (float64, bool) {
	switch j.Value.(type) {
	case float64:
		return float64(j.Value.(float64)), true
	case float32:
		return float64(j.Value.(float32)), true
	case int:
		return float64(j.Value.(int)), true
	case int64:
		return float64(j.Value.(int64)), true
	case uint:
		return float64(j.Value.(uint)), true
	case uint64:
		return float64(j.Value.(uint64)), true
	case uint32:
		return float64(j.Value.(uint32)), true
	case byte:
		return float64(j.Value.(byte)), true
	default:
		return 0.0, false
	}
}

// AsString returns the JSON value as a string if it's a string.
func (j JSON) AsString() (string, bool) {
	switch j.Value.(type) {
	case string:
		return j.Value.(string), true
	default:
		return "", false
	}
}

// AsBytes returns the JSON value as bytes if it's a hex-encoded string.
func (j JSON) AsBytes() ([]byte, bool) {
	switch j.Value.(type) {
	case string:
		if b, err := hex.DecodeString(j.Value.(string)); err == nil {
			return b, true
		}
		return nil, false
	default:
		return nil, false
	}
}

// AsBool returns the JSON value as a boolean if it's a boolean.
func (j JSON) AsBool() (bool, bool) {
	switch j.Value.(type) {
	case bool:
		return j.Value.(bool), true
	default:
		return false, false
	}
}

// AsNumericDefault returns the JSON value as float64 or defaultValue if not numeric.
func (j JSON) AsNumericDefault(defaultValue float64) float64 {
	if v, ok := j.AsNumeric(); ok {
		return v
	}
	return defaultValue
}

// AsStringDefault returns the JSON value as string or defaultValue if not string.
func (j JSON) AsStringDefault(defaultValue string) string {
	if v, ok := j.AsString(); ok {
		return v
	}
	return defaultValue
}

// AsBoolDefault returns the JSON value as boolean or defaultValue if not boolean.
func (j JSON) AsBoolDefault(defaultValue bool) bool {
	if v, ok := j.AsBool(); ok {
		return v
	}
	return defaultValue
}

// JSONFromBytes creates a JSON instance from byte slice by parsing it as JSON.
func JSONFromBytes(b []byte) (JSON, bool) {
	var j JSON
	if err := json.Unmarshal(b, &j.Value); err != nil {
		return NewJSONNull(), false
	}
	return j, true
}

// JSONFromString creates a JSON instance from string by parsing it as JSON.
func JSONFromString(s string) (JSON, bool) {
	return JSONFromBytes([]byte(s))
}

// Deprecated: JSONFromArray is no longer needed — use NewJSON instead, it now handles slices properly.
// JSONFromArray creates a JSON array from any comparable slice.
func JSONFromArray[T comparable](array []T) JSON {
	new_array := make([]interface{}, len(array))
	for i, v := range array {
		new_array[i] = v
	}
	return NewJSON(new_array)
}

// SetByPaths sets multiple paths at once for better DX
func (j *JSON) SetByPaths(pathValues map[string]interface{}) {
	for path, value := range pathValues {
		j.SetByPath(path, NewJSON(value))
	}
}

// NewJSONObjectFromMap creates JSON object directly from map
func NewJSONObjectFromMap(data map[string]interface{}) JSON {
	result := NewJSONObject()
	for k, v := range data {
		result.SetByPath(k, NewJSON(v))
	}
	return result
}

// BuildFromTemplate builds JSON from template with value substitution
func BuildFromTemplate(template map[string]interface{}) JSON {
	return NewJSONObjectFromMap(template)
}

// JSONBuilder provides fluent interface for JSON construction.
// It allows method chaining for building complex JSON structures efficiently.
type JSONBuilder struct {
	json JSON
}

// NewJSONBuilder creates a new JSON builder
func NewJSONBuilder() *JSONBuilder {
	return &JSONBuilder{json: NewJSONObject()}
}

// Set sets a path value and returns builder for chaining
func (b *JSONBuilder) Set(path string, value interface{}) *JSONBuilder {
	b.json.SetByPath(path, NewJSON(value))
	return b
}

// SetIfNotEmpty sets value only if it's not empty/zero
func (b *JSONBuilder) SetIfNotEmpty(path string, value interface{}) *JSONBuilder {
	if value != nil && value != "" && value != 0 {
		b.json.SetByPath(path, NewJSON(value))
	}
	return b
}

// AddToArray adds value to array at path
func (b *JSONBuilder) AddToArray(path string, value interface{}) *JSONBuilder {
	if !b.json.PathExists(path) {
		b.json.SetByPath(path, NewJSONArray())
	}
	arr := b.json.GetByPath(path)
	arr.AddToArray(NewJSON(value))
	b.json.SetByPath(path, arr)
	return b
}

// Build returns the final JSON object
func (b *JSONBuilder) Build() JSON {
	return b.json
}

// BuildArrayFromSlice creates JSON array from Go slice with transformation
func BuildArrayFromSlice[T any](slice []T, transform func(T) map[string]interface{}) JSON {
	arr := NewJSONArray()
	for _, item := range slice {
		arr.AddToArray(NewJSONObjectFromMap(transform(item)))
	}
	return arr
}
