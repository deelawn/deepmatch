package deepmatch

import (
	"reflect"
	"unsafe"
)

// Matcher defines properties that define how two interfaces should be compared.
type Matcher struct {
	// MaxDepth is the number of times Match can be called recursively when dealing with nested data structures.
	// A value of 0 means that there is no maximum depth.
	MaxDepth uint
	// ExcludeExported is set to true when equality checks should skip exported fields.
	ExcludeExported bool
	// ExcludeUnexported is set to true when equality checks should skip unexported fields.
	ExcludeUnexported bool
	// ExcludedFieldNames is a list of field names that should not be checked for equality. Dot separators need
	// to be used to separate structure name from field name.
	ExcludedFieldNames []string
}

// NewValueMatcher returns a type, ValueMatcher, that stores both m and a value
func (m Matcher) NewValueMatcher(value interface{}) ValueMatcher {
	return ValueMatcher{
		Matcher: m,
		Value:   value,
	}
}

// Matches uses a Matcher's properties and equality checks to indicate whether the values of x and y match.
// This is heavily inspired by reflect.DeepEqual from go's standard library.
func (m Matcher) Matches(x, y interface{}) bool {

	if x == nil || y == nil {
		return x == y
	}
	v1 := reflect.ValueOf(x)
	v2 := reflect.ValueOf(y)
	if v1.Type() != v2.Type() {
		return false
	}
	return matchesValue(v1, v2, make(map[visit]bool), 0)
}

// During deepValueEqual, must keep track of checks that are
// in progress. The comparison algorithm assumes that all
// checks in progress are true when it reencounters them.
// Visited comparisons are stored in a map indexed by visit.
type visit struct {
	a1  unsafe.Pointer
	a2  unsafe.Pointer
	typ reflect.Type
}

//go:linkname reflect_valueInterface reflect.valueInterface
func reflect_valueInterface(v reflect.Value, safe bool) interface{}

// Tests for deep equality using reflected types. The map argument tracks
// comparisons that have already been seen, which allows short circuiting on
// recursive types.
func matchesValue(v1, v2 reflect.Value, visited map[visit]bool, depth int) bool {
	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid()
	}
	if v1.Type() != v2.Type() {
		return false
	}

	// if depth > 10 { panic("deepValueEqual") }	// for debugging

	// We want to avoid putting more in the visited map than we need to.
	// For any possible reference cycle that might be encountered,
	// hard(v1, v2) needs to return true for at least one of the types in the cycle,
	// and it's safe and valid to get Value's internal pointer.
	hard := func(v1, v2 reflect.Value) bool {
		switch v1.Kind() {
		case reflect.Map, reflect.Slice, reflect.Ptr, reflect.Interface:
			// Nil pointers cannot be cyclic. Avoid putting them in the visited map.
			return !v1.IsNil() && !v2.IsNil()
		}
		return false
	}

	if hard(v1, v2) {
		// For a Ptr or Map value, we need to check flagIndir,
		// which we do by calling the pointer method.
		// For Slice or Interface, flagIndir is always set,
		// and using v.ptr suffices.
		ptrval := func(v reflect.Value) unsafe.Pointer {
			switch v.Kind() {
			case reflect.Ptr, reflect.Map:
				return unsafe.Pointer(v.Pointer())
			default:
				return unsafe.Pointer(reflect.ValueOf(reflect.ValueOf(v).FieldByName("ptr").Pointer()).Interface().(uintptr))
			}
		}
		addr1 := ptrval(v1)
		addr2 := ptrval(v2)
		if uintptr(addr1) > uintptr(addr2) {
			// Canonicalize order to reduce number of entries in visited.
			// Assumes non-moving garbage collector.
			addr1, addr2 = addr2, addr1
		}

		// Short circuit if references are already seen.
		typ := v1.Type()
		v := visit{addr1, addr2, typ}
		if visited[v] {
			return true
		}

		// Remember for later.
		visited[v] = true
	}

	switch v1.Kind() {
	case reflect.Array:
		for i := 0; i < v1.Len(); i++ {
			if !matchesValue(v1.Index(i), v2.Index(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Slice:
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		for i := 0; i < v1.Len(); i++ {
			if !matchesValue(v1.Index(i), v2.Index(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Interface:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		return matchesValue(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Ptr:
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		return matchesValue(v1.Elem(), v2.Elem(), visited, depth+1)
	case reflect.Struct:
		for i, n := 0, v1.NumField(); i < n; i++ {
			if !matchesValue(v1.Field(i), v2.Field(i), visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Map:
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		if v1.Len() != v2.Len() {
			return false
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		for _, k := range v1.MapKeys() {
			val1 := v1.MapIndex(k)
			val2 := v2.MapIndex(k)
			if !val1.IsValid() || !val2.IsValid() || !matchesValue(val1, val2, visited, depth+1) {
				return false
			}
		}
		return true
	case reflect.Func:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		// Can't do better than this:
		return false
	default:
		// Normal equality suffices
		return reflect_valueInterface(v1, false) == reflect_valueInterface(v2, false)
	}
}
