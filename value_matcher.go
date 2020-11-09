package deepmatch

import "fmt"

// ValueMatcher is a matcher defined with a value. It implements the gomock.Matcher interface so that it
// can be used to wrap any value.
// See https://github.com/golang/mock
type ValueMatcher struct {
	Matcher
	Value interface{}
}

func (vm ValueMatcher) Matches(x interface{}) bool {
	return vm.Matcher.Matches(x, vm.Value)
}

func (vm ValueMatcher) String() string {
	return fmt.Sprintf("is equal to %v", vm.Value)
}
