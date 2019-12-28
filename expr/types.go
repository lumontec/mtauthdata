// argument types. to let functions describe their inputs and outputs
package expr

//import "regexp"

// Arg is an argument to a GraphiteFunc
// note how every implementation has a val property.
// this property should point to value accessible to the function.
// the value will be set up by the planner; it assures that
// by the time Func.Exec() is called, the function has access to all
// needed inputs, whether simple values, or in the case of ArgSeries*
// inputs other functions to call which will feed it data.
type Arg interface {
	Key() string
	Optional() bool
}

