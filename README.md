# Deep Match
### A more versatile reflect.DeepEqual
---

Use this package to compare define and compare equality of values of all types.  It works like reflect.DeepEqual, but allows you to set options defined by the matcher.

Use a Matcher to define how equality is determined
```
type Matcher struct {
	// MaxDepth is the number of times Match can be called recursively when dealing with nested data structures.
	// A value of 0 means that there is no maximum depth.
	MaxDepth int
  
	// ExcludeExported is set to true when equality checks should skip exported fields.
	ExcludeExported bool
  
	// ExcludeUnexported is set to true when equality checks should skip unexported fields.
	ExcludeUnexported bool
  
	// ExcludedFieldNames is a list of field names that should not be checked for equality. Dot separators need
	// to be used to separate structure name from field name.
	ExcludedFieldNames []string
}
```
#### Examples:

```
type s struct {
  A int
  b string
}

m := Matcher {
  ExcludeUnexported: true,
}

isEqual := m.Matches(s{A: 6, b: "asdf"}, s{A: 6, b: "fdsa"})
// isEqual = true

m.ExcludeUnexported = false
isEqual = m.Matches(s{A: 6, b: "asdf"}, s{A: 6, b: "fdsa"})
// isEqual = false

m.ExcludedFieldNames = []string{"b"}
isEqual = m.Matches(s{A: 6, b: "asdf"}, s{A: 6, b: "fdsa"})
// isEqual = true
```
