package factory

import (
	"fmt"
	"reflect"

	randomdata "github.com/Pallinder/go-randomdata"
)

// Value converts passed value into static generator
func Value(i interface{}) GeneratorFunc {
	return func(ExecCtx) (interface{}, error) {
		return i, nil
	}
}

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

// Func tries to adapt arbitrary function to be used as generator
func Func(f interface{}, args ...interface{}) GeneratorFunc {
	val := reflect.ValueOf(f)

	// panic if `f' is not a function
	if val.Kind() != reflect.Func {
		panic(fmt.Errorf("expect function type but was: %T", f))
	}

	typ := reflect.TypeOf(f)

	// check input argumrnts
	if !typ.IsVariadic() && typ.NumIn() != len(args) {
		panic(fmt.Errorf("not enough input arguments to make a function call. Expected: %d, was: %d",
			typ.NumIn(), len(args)))
	}

	// check function signature. Perimted number is 1 or 2
	if typ.NumOut() == 0 || typ.NumOut() > 2 {
		panic(fmt.Errorf("expect function to return 1 or 2 values but was: %d", typ.NumOut()))
	}

	// check second output parameter implements error interface
	if typ.NumOut() == 2 && !typ.Out(1).Implements(errorInterface) {
		panic(fmt.Errorf("expect second returned type implements error: %+v", typ.Out(1)))
	}

	// prepare input arguments
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	return func(ExecCtx) (interface{}, error) {
		r := val.Call(in)
		if len(r) == 1 || r[1].IsNil() {
			return r[0].Interface(), nil
		}
		return r[0].Interface(), r[1].Interface().(error)
	}
}

// Seq returns function that sequentially generates integers in interval [0, max)
func Seq(max int) func() int {
	n := 0
	return func() int {
		defer func() { n++ }()
		return n % max
	}
}

// Rnd returns function that randomly enerates integers in interval [0, max)
func Rnd(max int) func() int {
	return func() int {
		return randomdata.Number(max)
	}
}

// Select picks a value from options
func Select(f func(int) func() int, options ...interface{}) GeneratorFunc {
	g := f(len(options))
	return func(ExecCtx) (interface{}, error) {
		return options[g()], nil
	}
}

// SeqSelect = Select(Seq, options...)
func SeqSelect(options ...interface{}) GeneratorFunc {
	return Select(Seq, options...)
}

// RndSelect = Select(Rnd, options...)
func RndSelect(options ...interface{}) GeneratorFunc {
	return Select(Rnd, options...)
}

// NewGenerator makes a field generator function
func NewGenerator(i interface{}, args ...interface{}) GeneratorFunc {
	// for usecases like:
	// func myGenFunc() GeneratorFunc {
	//   return func(ExecCtx) (interface{}, error) { ...  }
	// }
	if genFunc, ok := i.(GeneratorFunc); ok {
		return genFunc
	}

	// for usecases like:
	// func myGenFunc() func(ExecCtx) (interface{}, error) {
	//   return func(ExecCtx) (interface{}, error) { ...  }
	// }
	//
	// or in-place declarations:
	// f := NewFactory(
	//   User{},
	//   Use(func(ctx ExecCtx) (interface{}, error) { ... }),
	// )
	if genFunc, ok := i.(func(ExecCtx) (interface{}, error)); ok {
		return genFunc
	}

	// if i is a factory use Create method
	if fact, ok := i.(*Factory); ok {
		return func(ExecCtx) (interface{}, error) {
			return fact.Create()
		}
	}

	// if i is a function, use function to generator converter
	if v := reflect.ValueOf(i); v.Kind() == reflect.Func {
		// use Func adapter in case i is of Kind Func
		return Func(i, args...)
	}

	// if it's just some static value, use value to generator converter
	if len(args) == 0 {
		// use static value generator if no other arguments provided
		return Value(i)
	}

	// otherwise make generator function to randomly select from given options
	return RndSelect(append([]interface{}{i}, args...)...)
}
