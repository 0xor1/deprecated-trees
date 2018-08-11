package json

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/0xor1/panic"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"
)

type Json struct {
	data interface{}
}

// New returns a pointer to a new, empty `Json` object
func New() (*Json, error) {
	return FromString("{}")
}

// MustNew is a call to New with a panic on none nil error
func MustNew() *Json {
	js, err := New()
	panic.IfNotNil(err)
	return js
}

// FromInterface returns a pointer to a new `Json` object
// after assigning `i` to its internal data
func FromInterface(i interface{}) *Json {
	return &Json{i}
}

// FromString returns a pointer to a new `Json` object
// after unmarshaling `str`
func FromString(str string) (*Json, error) {
	return FromBytes([]byte(str))
}

// MustFromString is a call to FromString with a panic on none nil error
func MustFromString(str string) *Json {
	js, err := FromString(str)
	panic.IfNotNil(err)
	return js
}

// FromBytes returns a pointer to a new `Json` object
// after unmarshaling `bytes`
func FromBytes(b []byte) (*Json, error) {
	return FromReader(bytes.NewReader(b))
}

// MustFromBytes is a call to FromBytes with a panic on none nil error
func MustFromBytes(b []byte) *Json {
	js, err := FromBytes(b)
	panic.IfNotNil(err)
	return js
}

// FromFile returns a pointer to a new `Json` object
// after unmarshaling the contents from `file` into it
func FromFile(file string) (*Json, error) {
	fullPath, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	return FromBytes(data)
}

// MustFromFile is a call to FromFile with a panic on none nil error
func MustFromFile(file string) *Json {
	js, err := FromFile(file)
	panic.IfNotNil(err)
	return js
}

// FromReader returns a *Json by decoding from an io.Reader
func FromReader(r io.Reader) (*Json, error) {
	if r == nil {
		return FromString("null")
	}
	rc, ok := r.(io.ReadCloser)
	if !ok {
		rc = ioutil.NopCloser(r)
	}
	return FromReadCloser(rc)
}

// MustFromReader is a call to FromReader with a panic on none nil error
func MustFromReader(r io.Reader) *Json {
	js, err := FromReader(r)
	panic.IfNotNil(err)
	return js
}

// FromReadCloser returns a *Json by decoding from an io.ReadCloser and calls the io.ReadCloser Close method
func FromReadCloser(rc io.ReadCloser) (*Json, error) {
	if rc == nil {
		return FromString("null")
	}
	defer rc.Close()
	j := &Json{}
	dec := json.NewDecoder(rc)
	dec.UseNumber()
	err := dec.Decode(&j.data)
	return j, err
}

// MustFromReadCloser is a call to FromReadCloser with a panic on none nil error
func MustFromReadCloser(rc io.ReadCloser) *Json {
	js, err := FromReadCloser(rc)
	panic.IfNotNil(err)
	return js
}

// ToBytes returns its marshaled data as `[]byte`
func (j *Json) ToBytes() ([]byte, error) {
	return j.MarshalJSON()
}

// MustToBytes is a call to ToBytes with a panic on none nil error
func (j *Json) MustToBytes() []byte {
	bs, err := j.ToBytes()
	panic.IfNotNil(err)
	return bs
}

// ToString returns its marshaled data as `string`
func (j *Json) ToString() (string, error) {
	b, err := j.ToBytes()
	return string(b), err
}

// MustToString is a call to ToString with a panic on none nil error
func (j *Json) MustToString() string {
	str, err := j.ToString()
	panic.IfNotNil(err)
	return str
}

// ToPrettyBytes returns its marshaled data as `[]byte` with indentation
func (j *Json) ToPrettyBytes() ([]byte, error) {
	return json.MarshalIndent(&j.data, "", "  ")
}

// MustToPrettyBytes is a call to ToPrettyBytes with a panic on none nil error
func (j *Json) MustToPrettyBytes() []byte {
	bs, err := j.ToPrettyBytes()
	panic.IfNotNil(err)
	return bs
}

// ToPrettyString returns its marshaled data as `string` with indentation
func (j *Json) ToPrettyString() (string, error) {
	b, err := j.ToPrettyBytes()
	return string(b), err
}

// MustToPrettyString is a call to ToPrettyString with a panic on none nil error
func (j *Json) MustToPrettyString() string {
	str, err := j.ToPrettyString()
	panic.IfNotNil(err)
	return str
}

// ToFile writes the Json to the `file` with permission `perm`
func (j *Json) ToFile(file string, perm os.FileMode) error {
	b, err := j.ToBytes()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, b, perm)
}

// MustToFile is a call to ToFile with a panic on none nil error
func (j *Json) MustToFile(file string, perm os.FileMode) {
	panic.IfNotNil(j.ToFile(file, perm))
}

// ToReader returns its marshaled data as `io.Reader`
func (j *Json) ToReader() (io.Reader, error) {
	b, err := j.ToBytes()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// MustToReader is a call to ToReader with a panic on none nil error
func (j *Json) MustToReader() io.Reader {
	r, err := j.ToReader()
	panic.IfNotNil(err)
	return r
}

// Implements the json.Marshaler interface.
func (j *Json) MarshalJSON() ([]byte, error) {
	return json.Marshal(&j.data)
}

// Implements the json.Unmarshaler interface.
func (j *Json) UnmarshalJSON(p []byte) error {
	jNew, err := FromReader(bytes.NewReader(p))
	j.data = jNew.data
	return err
}

// Get searches for the item as specified by the path.
// path can contain strings or ints to navigate through json
// objects and slices. If the given path is not present then
// the deepest valid value is returned along with an error.
//
//   js.Get("top_level", "dict", 3, "foo")
func (j *Json) Get(path ...interface{}) (*Json, error) {
	tmp := j
	for i, k := range path {
		if key, ok := k.(string); ok {
			if m, err := tmp.Map(); err == nil {
				if val, ok := m[key]; ok {
					tmp = &Json{val}
				} else {
					return tmp, &jsonPathError{path[:i], path[i:]}
				}
			} else {
				return tmp, &jsonPathError{path[:i], path[i:]}
			}
		} else if index, ok := k.(int); ok {
			if a, err := tmp.Slice(); err == nil {
				if index < 0 || index >= len(a) {
					return tmp, &jsonPathError{path[:i], path[i:]}
				} else {
					tmp = &Json{a[index]}
				}
			} else {
				return tmp, &jsonPathError{path[:i], path[i:]}
			}
		} else {
			return tmp, &jsonPathError{path[:i], path[i:]}
		}
	}
	return tmp, nil
}

// MustGet is a call to Get with a panic on none nil error
func (j *Json) MustGet(path ...interface{}) *Json {
	js, err := j.Get(path...)
	panic.IfNotNil(err)
	return js
}

// Set modifies `Json`, recursively checking/creating map keys and checking
// slice indices for the supplied path, and then finally writing in the value.
// Set will only create maps where the current map[key] does not exist,
// if the key exists, even if the value is nil, a new map will not be created and an
// error wil be returned.
func (j *Json) Set(val interface{}, path ...interface{}) error {
	if len(path) == 0 {
		j.data = val
		return nil
	}

	tmp := j

	for i := 0; i < len(path); i++ {
		if key, ok := path[i].(string); ok {
			if m, err := tmp.Map(); err == nil {
				if i == len(path)-1 {
					m[key] = val
				} else {
					_, ok := path[i+1].(string)
					_, exists := m[key]
					if ok && !exists {
						m[key] = map[string]interface{}{}
					}
					tmp = &Json{m[key]}
				}
			} else {
				return &jsonPathError{path[:i], path[i:]}
			}
		} else if index, ok := path[i].(int); ok {
			if a, err := tmp.Slice(); err == nil && index >= 0 && index < len(a) {
				if i == len(path)-1 {
					a[index] = val
				} else {
					tmp = &Json{a[index]}
				}
			} else {
				return &jsonPathError{path[:i], path[i:]}
			}
		} else {
			return &jsonPathError{path[:i], path[i:]}
		}
	}

	return nil
}

// MustSet is a call to Set with a panic on none nil error
func (j *Json) MustSet(val interface{}, path ...interface{}) {
	panic.IfNotNil(j.Set(val, path...))
}

// Del modifies `Json` maps and slices by deleting/removing the last `path` segment if it is present,
func (j *Json) Del(path ...interface{}) error {
	if len(path) == 0 {
		j.data = nil
		return nil
	}

	i := len(path) - 1
	tmp, err := j.Get(path[:i]...)
	if err != nil {
		err.(*jsonPathError).MissingPath = append(err.(*jsonPathError).MissingPath, path[i])
		return err
	}

	if key, ok := path[i].(string); ok {
		if m, err := tmp.Map(); err != nil {
			return &jsonPathError{path[:i], path[i:]}
		} else {
			delete(m, key)
		}
	} else if index, ok := path[i].(int); ok {
		if a, err := tmp.Slice(); err != nil {
			return &jsonPathError{path[:i], path[i:]}
		} else if index < 0 || index >= len(a) {
			return &jsonPathError{path[:i], path[i:]}
		} else {
			a, a[len(a)-1] = append(a[:index], a[index+1:]...), nil
			if i == 0 {
				j.data = a
			} else {
				tmp, _ = j.Get(path[:i-1]...)
				if key, ok := path[i-1].(string); ok {
					tmp.MapOrDefault(nil)[key] = a //is this safe? should be 100% certainty ;)
				} else if index, ok := path[i-1].(int); ok {
					tmp.SliceOrDefault(nil)[index] = a //is this safe? should be 100% certainty ;)
				}
			}
		}
	} else {
		return &jsonPathError{path[:i], path[i:]}
	}
	return nil
}

// MustDel is a call to Del with a panic on none nil error
func (j *Json) MustDel(path ...interface{}) {
	panic.IfNotNil(j.Del(path...))
}

// Interface returns the underlying data
func (j *Json) Interface(path ...interface{}) (interface{}, error) {
	tmp, err := j.Get(path...)
	return tmp.data, err
}

// MustInterface is a call to Interface with a panic on none nil error
func (j *Json) MustInterface(path ...interface{}) interface{} {
	i, err := j.Interface(path...)
	panic.IfNotNil(err)
	return i
}

// Map type asserts to `map[string]interface{}`
func (j *Json) Map(path ...interface{}) (map[string]interface{}, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return nil, err
	}
	if m, ok := tmp.data.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errors.New("type assertion to map[string]interface{} failed")
}

// MustMap is a call to Map with a panic on none nil error
func (j *Json) MustMap(path ...interface{}) map[string]interface{} {
	v, err := j.Map(path...)
	panic.IfNotNil(err)
	return v
}

// MapOrDefault guarantees the return of a `map[string]interface{}` (with specified default)
//
// useful when you want to iterate over map values in a succinct manner:
//		for k, v := range js.MapOrDefault(nil) {
//			fmt.Println(k, v)
//		}
func (j *Json) MapOrDefault(def map[string]interface{}, path ...interface{}) map[string]interface{} {
	if a, err := j.Map(path...); err == nil {
		return a
	}
	return def
}

// Map type asserts to `map[string]string`
func (j *Json) MapString(path ...interface{}) (map[string]string, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return nil, err
	}
	if m, ok := tmp.data.(map[string]interface{}); ok {
		ms := map[string]string{}
		for k, v := range m {
			if kStr, ok := v.(string); ok {
				ms[k] = kStr
			} else {
				return nil, errors.New("type assertion of map value to string failed")
			}
		}
		return ms, nil
	}
	return nil, errors.New("type assertion to map[string]string{} failed")
}

// MustMapString is a call to MapString with a panic on none nil error
func (j *Json) MustMapString(path ...interface{}) map[string]string {
	v, err := j.MapString(path...)
	panic.IfNotNil(err)
	return v
}

// MapStringOrDefault guarantees the return of a `map[string]string{}` (with specified default)
//
// useful when you want to iterate over map values in a succinct manner:
//		for k, v := range js.MapStringOrDefault(nil) {
//			fmt.Println(k, v)
//		}
func (j *Json) MapStringOrDefault(def map[string]string, path ...interface{}) map[string]string {
	if m, err := j.MapString(path...); err == nil {
		return m
	}
	return def
}

// Slice type asserts to a `slice`
func (j *Json) Slice(path ...interface{}) ([]interface{}, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return nil, err
	}
	if a, ok := tmp.data.([]interface{}); ok {
		return a, nil
	}
	return nil, errors.New("type assertion to []interface{} failed")
}

// MustSlice is a call to MustSlice with a panic on none nil error
func (j *Json) MustSlice(path ...interface{}) []interface{} {
	v, err := j.Slice(path...)
	panic.IfNotNil(err)
	return v
}

// SliceOrDefault guarantees the return of a `[]interface{}` (with specified default)
//
// useful when you want to iterate over slice values in a succinct manner:
//		for i, v := range js.SliceOrDefault(nil) {
//			fmt.Println(i, v)
//		}
func (j *Json) SliceOrDefault(def []interface{}, path ...interface{}) []interface{} {
	if a, err := j.Slice(path...); err == nil {
		return a
	}
	return def
}

// Bool type asserts to `bool`
func (j *Json) Bool(path ...interface{}) (bool, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return false, err
	}
	if s, ok := tmp.data.(bool); ok {
		return s, nil
	}
	return false, errors.New("type assertion to bool failed")
}

// MustBool is a call to Bool with a panic on none nil error
func (j *Json) MustBool(path ...interface{}) bool {
	v, err := j.Bool(path...)
	panic.IfNotNil(err)
	return v
}

// BoolOrDefault guarantees the return of a `bool` (with specified default)
//
// useful when you explicitly want a `bool` in a single value return context:
//     myFunc(js.BoolOrDefault(true))
func (j *Json) BoolOrDefault(def bool, path ...interface{}) bool {
	if b, err := j.Bool(path...); err == nil {
		return b
	}
	return def
}

// String type asserts to `string`
func (j *Json) String(path ...interface{}) (string, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return "", err
	}
	if s, ok := tmp.data.(string); ok {
		return s, nil
	}
	return "", errors.New("type assertion to string failed")
}

// MustString is a call to String with a panic on none nil error
func (j *Json) MustString(path ...interface{}) string {
	v, err := j.String(path...)
	panic.IfNotNil(err)
	return v
}

// StringOrDefault guarantees the return of a `string` (with specified default)
//
// useful when you explicitly want a `string` in a single value return context:
//     myFunc(js.StringOrDefault("my_default"))
func (j *Json) StringOrDefault(def string, path ...interface{}) string {
	if s, err := j.String(path...); err == nil {
		return s
	}
	return def
}

// StringSlice type asserts to a `slice` of `string`
func (j *Json) StringSlice(path ...interface{}) ([]string, error) {
	arr, err := j.Slice(path...)
	if err != nil {
		return nil, err
	}
	retArr := make([]string, 0, len(arr))
	for _, a := range arr {
		if s, ok := a.(string); a == nil || !ok {
			return nil, errors.New("none string value encountered")
		} else {
			retArr = append(retArr, s)
		}
	}
	return retArr, nil
}

// MustStringSlice is a call to StringSlice with a panic on none nil error
func (j *Json) MustStringSlice(path ...interface{}) []string {
	v, err := j.StringSlice(path...)
	panic.IfNotNil(err)
	return v
}

// StringSliceOrDefault guarantees the return of a `[]string` (with specified default)
//
// useful when you want to iterate over slice values in a succinct manner:
//		for i, s := range js.StringSliceOrDefault(nil) {
//			fmt.Println(i, s)
//		}
func (j *Json) StringSliceOrDefault(def []string, path ...interface{}) []string {
	if a, err := j.StringSlice(path...); err == nil {
		return a
	}
	return def
}

// Time type asserts to `time.Time`
func (j *Json) Time(path ...interface{}) (time.Time, error) {
	var t time.Time
	tmp, err := j.Get(path...)
	if err != nil {
		return t, err
	}
	if t, ok := tmp.data.(time.Time); ok {
		return t, nil
	} else if tStr, ok := tmp.data.(string); ok {
		if t.UnmarshalText([]byte(tStr)) == nil {
			return t, nil
		}
	}
	return t, errors.New("type assertion/unmarshalling to time.Time failed")
}

// MustTime is a call to Time with a panic on none nil error
func (j *Json) MustTime(path ...interface{}) time.Time {
	v, err := j.Time(path...)
	panic.IfNotNil(err)
	return v
}

// TimeOrDefault guarantees the return of a `time.Time` (with specified default)
//
// useful when you explicitly want a `time.Time` in a single value return context:
//     myFunc(js.TimeOrDefault(defaultTime))
func (j *Json) TimeOrDefault(def time.Time, path ...interface{}) time.Time {
	if t, err := j.Time(path...); err == nil {
		return t
	}
	return def
}

// TimeSlice type asserts to a `slice` of `time.Time`
func (j *Json) TimeSlice(path ...interface{}) ([]time.Time, error) {
	arr, err := j.Slice(path...)
	if err != nil {
		return nil, err
	}
	retArr := make([]time.Time, 0, len(arr))
	for _, a := range arr {
		if s, ok := a.(time.Time); a == nil || !ok {
			return nil, errors.New("none time.Time value encountered")
		} else {
			retArr = append(retArr, s)
		}
	}
	return retArr, nil
}

// MustTimeSlice is a call to TimeSlice with a panic on none nil error
func (j *Json) MustTimeSlice(path ...interface{}) []time.Time {
	v, err := j.TimeSlice(path...)
	panic.IfNotNil(err)
	return v
}

// TimeSliceOrDefault guarantees the return of a `[]time.Time` (with specified default)
//
// useful when you want to iterate over slice values in a succinct manner:
//		for i, t := range js.TimeSliceOrDefault(nil) {
//			fmt.Println(i, t)
//		}
func (j *Json) TimeSliceOrDefault(def []time.Time, path ...interface{}) []time.Time {
	if a, err := j.TimeSlice(path...); err == nil {
		return a
	}
	return def
}

// Duration type asserts to `time.Duration`
func (j *Json) Duration(path ...interface{}) (time.Duration, error) {
	var d time.Duration
	tmp, err := j.String(path...)
	if err != nil {
		return d, err
	}
	return time.ParseDuration(tmp)
}

// MustDuration is a call to Duration with a panic on none nil error
func (j *Json) MustDuration(path ...interface{}) time.Duration {
	v, err := j.Duration(path...)
	panic.IfNotNil(err)
	return v
}

// DurationOrDefault guarantees the return of a `time.Duration` (with specified default)
//
// useful when you explicitly want a `time.Duration` in a single value return context:
//     myFunc(js.DurationOrDefault(defaultDuration))
func (j *Json) DurationOrDefault(def time.Duration, path ...interface{}) time.Duration {
	if d, err := j.Duration(path...); err == nil {
		return d
	}
	return def
}

// DurationSlice type asserts to a `slice` of `time.Duration`
func (j *Json) DurationSlice(path ...interface{}) ([]time.Duration, error) {
	arr, err := j.StringSlice(path...)
	if err != nil {
		return nil, err
	}
	retArr := make([]time.Duration, 0, len(arr))
	for _, a := range arr {
		if d, err := time.ParseDuration(a); err != nil {
			return nil, err
		} else {
			retArr = append(retArr, d)
		}
	}
	return retArr, nil
}

// MustDurationSlice is a call to DurationSlice with a panic on none nil error
func (j *Json) MustDurationSlice(path ...interface{}) []time.Duration {
	v, err := j.DurationSlice(path...)
	panic.IfNotNil(err)
	return v
}

// DurationSliceOrDefault guarantees the return of a `[]time.Duration` (with specified default)
//
// useful when you want to iterate over slice values in a succinct manner:
//		for i, t := range js.DurationSliceOrDefault(nil) {
//			fmt.Println(i, t)
//		}
func (j *Json) DurationSliceOrDefault(def []time.Duration, path ...interface{}) []time.Duration {
	if a, err := j.DurationSlice(path...); err == nil {
		return a
	}
	return def
}

// Int coerces into an int
func (j *Json) Int(path ...interface{}) (int, error) {
	f, err := j.Float64(path...)
	return int(f), err
}

// MustInt is a call to Int with a panic on none nil error
func (j *Json) MustInt(path ...interface{}) int {
	v, err := j.Int(path...)
	panic.IfNotNil(err)
	return v
}

// IntOrDefault guarantees the return of an `int` (with specified default)
//
// useful when you explicitly want an `int` in a single value return context:
//     myFunc(js.IntOrDefault(5150))
func (j *Json) IntOrDefault(def int, path ...interface{}) int {
	if i, err := j.Int(path...); err == nil {
		return i
	}
	return def
}

// IntSlice type asserts to a `slice` of `int`
func (j *Json) IntSlice(path ...interface{}) ([]int, error) {
	arr, err := j.Slice(path...)
	if err != nil {
		return nil, err
	}
	retArr := make([]int, 0, len(arr))
	for _, a := range arr {
		tmp := &Json{a}
		if i, err := tmp.Int(); err != nil {
			return nil, err
		} else {
			retArr = append(retArr, i)
		}
	}
	return retArr, nil
}

// MustIntSlice is a call to IntSlice with a panic on none nil error
func (j *Json) MustIntSlice(path ...interface{}) []int {
	v, err := j.IntSlice(path...)
	panic.IfNotNil(err)
	return v
}

// IntSliceOrDefault guarantees the return of a `[]int` (with specified default)
//
// useful when you want to iterate over slice values in a succinct manner:
//		for i, s := range js.IntSliceOrDefault(nil) {
//			fmt.Println(i, s)
//		}
func (j *Json) IntSliceOrDefault(def []int, path ...interface{}) []int {
	if a, err := j.IntSlice(path...); err == nil {
		return a
	}
	return def
}

// Float64 coerces into a float64
func (j *Json) Float64(path ...interface{}) (float64, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return 0, err
	}
	switch tmp.data.(type) {
	case string:
		return json.Number(tmp.data.(string)).Float64()
	case json.Number:
		return tmp.data.(json.Number).Float64()
	case float32, float64:
		return reflect.ValueOf(tmp.data).Float(), nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(tmp.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(tmp.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// MustFloat64 is a call to Float64 with a panic on none nil error
func (j *Json) MustFloat64(path ...interface{}) float64 {
	v, err := j.Float64(path...)
	panic.IfNotNil(err)
	return v
}

// Float64OrDefault guarantees the return of a `float64` (with specified default)
//
// useful when you explicitly want a `float64` in a single value return context:
//     myFunc(js.Float64OrDefault(5.150))
func (j *Json) Float64OrDefault(def float64, path ...interface{}) float64 {
	if f, err := j.Float64(path...); err == nil {
		return f
	}
	return def
}

// Float64Slice type asserts to a `slice` of `float64`
func (j *Json) Float64Slice(path ...interface{}) ([]float64, error) {
	arr, err := j.Slice(path...)
	if err != nil {
		return nil, err
	}
	retArr := make([]float64, 0, len(arr))
	for _, a := range arr {
		tmp := &Json{a}
		if f, err := tmp.Float64(); err != nil {
			return nil, err
		} else {
			retArr = append(retArr, f)
		}
	}
	return retArr, nil
}

// MustFloat64Slice is a call to Float64Slice with a panic on none nil error
func (j *Json) MustFloat64Slice(path ...interface{}) []float64 {
	v, err := j.Float64Slice(path...)
	panic.IfNotNil(err)
	return v
}

// Float64SliceOrDefault guarantees the return of a `[]float64` (with specified default)
//
// useful when you want to iterate over slice values in a succinct manner:
//		for i, s := range js.Float64SliceOrDefault(nil) {
//			fmt.Println(i, s)
//		}
func (j *Json) Float64SliceOrDefault(def []float64, path ...interface{}) []float64 {
	if a, err := j.Float64Slice(path...); err == nil {
		return a
	}
	return def
}

// Int64 coerces into an int64
func (j *Json) Int64(path ...interface{}) (int64, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return 0, err
	}
	switch tmp.data.(type) {
	case string:
		return json.Number(tmp.data.(string)).Int64()
	case json.Number:
		return tmp.data.(json.Number).Int64()
	case float32, float64:
		return int64(reflect.ValueOf(tmp.data).Float()), nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(tmp.data).Int(), nil
	case uint, uint8, uint16, uint32, uint64:
		return int64(reflect.ValueOf(tmp.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// MustInt64 is a call to Int64 with a panic on none nil error
func (j *Json) MustInt64(path ...interface{}) int64 {
	v, err := j.Int64(path...)
	panic.IfNotNil(err)
	return v
}

// Int64OrDefault guarantees the return of an `int64` (with specified default)
//
// useful when you explicitly want an `int64` in a single value return context:
//     myFunc(js.Int64OrDefault(5150))
func (j *Json) Int64OrDefault(def int64, path ...interface{}) int64 {
	if i, err := j.Int64(path...); err == nil {
		return i
	}
	return def
}

// Int64Slice type asserts to a `slice` of `int64`
func (j *Json) Int64Slice(path ...interface{}) ([]int64, error) {
	arr, err := j.Slice(path...)
	if err != nil {
		return nil, err
	}
	retArr := make([]int64, 0, len(arr))
	for _, a := range arr {
		tmp := &Json{a}
		if i, err := tmp.Int64(); err != nil {
			return nil, err
		} else {
			retArr = append(retArr, i)
		}
	}
	return retArr, nil
}

// MustInt64 is a call to Int64Slice with a panic on none nil error
func (j *Json) MustInt64Slice(path ...interface{}) []int64 {
	v, err := j.Int64Slice(path...)
	panic.IfNotNil(err)
	return v
}

// Int64SliceDefault guarantees the return of a `[]int64` (with specified default)
//
// useful when you want to iterate over slice values in a succinct manner:
//		for i, s := range js.Int64SliceDefault(nil) {
//			fmt.Println(i, s)
//		}
func (j *Json) Int64SliceDefault(def []int64, path ...interface{}) []int64 {
	if a, err := j.Int64Slice(path...); err == nil {
		return a
	}
	return def
}

// Uint64 coerces into an uint64
func (j *Json) Uint64(path ...interface{}) (uint64, error) {
	tmp, err := j.Get(path...)
	if err != nil {
		return 0, err
	}
	switch tmp.data.(type) {
	case string:
		return strconv.ParseUint(tmp.data.(string), 10, 64)
	case json.Number:
		return strconv.ParseUint(tmp.data.(json.Number).String(), 10, 64)
	case float32, float64:
		return uint64(reflect.ValueOf(tmp.data).Float()), nil
	case int, int8, int16, int32, int64:
		return uint64(reflect.ValueOf(tmp.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(tmp.data).Uint(), nil
	}
	return 0, errors.New("invalid value type")
}

// MustUint64 is a call to Uint64 with a panic on none nil error
func (j *Json) MustUint64(path ...interface{}) uint64 {
	v, err := j.Uint64(path...)
	panic.IfNotNil(err)
	return v
}

// MustUInt64 guarantees the return of an `uint64` (with specified default)
//
// useful when you explicitly want an `uint64` in a single value return context:
//     myFunc(js.Uint64OrDefault(5150))
func (j *Json) Uint64OrDefault(def uint64, path ...interface{}) uint64 {
	if i, err := j.Uint64(path...); err == nil {
		return i
	}
	return def
}

// Uint64Slice type asserts to a `slice` of `uint64`
func (j *Json) Uint64Slice(path ...interface{}) ([]uint64, error) {
	arr, err := j.Slice(path...)
	if err != nil {
		return nil, err
	}
	retArr := make([]uint64, 0, len(arr))
	for _, a := range arr {
		tmp := &Json{a}
		if u, err := tmp.Uint64(); err != nil {
			return nil, err
		} else {
			retArr = append(retArr, u)
		}
	}
	return retArr, nil
}

// MustUint64Slice is a call to Uint64Slice with a panic on none nil error
func (j *Json) MustUint64Slice(path ...interface{}) []uint64 {
	v, err := j.Uint64Slice(path...)
	panic.IfNotNil(err)
	return v
}

// Uint64SliceOrDefault guarantees the return of a `[]uint64` (with specified default)
//
// useful when you want to iterate over slice values in a succinct manner:
//		for i, s := range js.Uint64SliceOrDefault(nil) {
//			fmt.Println(i, s)
//		}
func (j *Json) Uint64SliceOrDefault(def []uint64, path ...interface{}) []uint64 {
	if a, err := j.Uint64Slice(path...); err == nil {
		return a
	}
	return def
}

type jsonPathError struct {
	FoundPath   []interface{}
	MissingPath []interface{}
}

func (e *jsonPathError) Error() string {
	return fmt.Sprintf("found: %v missing: %v", e.FoundPath, e.MissingPath)
}
