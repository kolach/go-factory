package factory

import (
	"fmt"
	"reflect"
)

// Ctx is the context in which the field value is being generated
type Ctx struct {
	Field    string      // current field name for which the value is generated
	Instance interface{} // the result instance to that the field belongs
	Factory  *Factory    // the reference to the Factory
}

// GeneratorFunc describes field generator signatures
type GeneratorFunc func(ctx Ctx) (interface{}, error)

// FieldGenFunc is the signature of field generator factory.
type FieldGenFunc func(sample reflect.Value) []fieldWithGen

// fieldWithGen is a tuple that keeps together struct field and generator function.
type fieldWithGen struct {
	*reflect.StructField
	gen GeneratorFunc
}

// Factory produces new objects according to specified generators
type Factory struct {
	typ       reflect.Type   // type information about generated instances
	fieldGens []fieldWithGen // field / generator tuples
	callDepth int            // factory call depth
}

// dive clones factory with incremented call depth
func (f *Factory) dive() *Factory {
	return &Factory{
		typ:       f.typ,
		fieldGens: f.fieldGens,
		callDepth: f.callDepth + 1,
	}
}

// CallDepth returns factory call depth
func (f *Factory) CallDepth() int {
	return f.callDepth
}

// Derive produces a new factory overriding field generators
// with the list provided.
func (f *Factory) Derive(fieldGenFuncs ...FieldGenFunc) *Factory {
	// Create new generators and lookup map to fast find generator by firld name
	newGenList := make([]fieldWithGen, 0, len(fieldGenFuncs))
	newGensMap := make(map[string]GeneratorFunc)
	sample := f.new()
	for _, fieldGenFunc := range fieldGenFuncs {
		for _, fg := range fieldGenFunc(sample) {
			newGensMap[fg.Name] = fg.gen
			newGenList = append(newGenList, fg)
		}
	}

	// result generators for a new factory
	fieldGens := make([]fieldWithGen, len(f.fieldGens))

	// 1. copy or override original field generators
	for i, fg := range f.fieldGens {
		if gen, ok := newGensMap[fg.Name]; ok {
			delete(newGensMap, fg.Name)
			fg.gen = gen
		}
		fieldGens[i] = fg
	}

	// 2. append new field generators
	for _, fg := range newGenList {
		if _, ok := newGensMap[fg.Name]; ok {
			fieldGens = append(fieldGens, fg)
		}
	}

	return &Factory{
		callDepth: f.callDepth, // inherit currenet call depth
		fieldGens: fieldGens,   // set new generators
		typ:       f.typ,
	}
}

func (f *Factory) new() reflect.Value {
	return reflect.New(f.typ)
}

// SetFields fills in the struct instance fields
func (f *Factory) SetFields(i interface{}, fieldGenFuncs ...FieldGenFunc) error {
	if len(fieldGenFuncs) > 0 {
		return f.Derive(fieldGenFuncs...).SetFields(i)
	}

	// create execution context
	ctx := Ctx{Instance: i, Factory: f.dive()}

	elem := reflect.ValueOf(i).Elem()

	for _, fg := range f.fieldGens {
		// bind field name o context
		ctx.Field = fg.Name

		// generate field value
		val, err := fg.gen(ctx)
		if err != nil {
			return err
		}

		valueof := reflect.ValueOf(val)

		switch valueof.Kind() {
		case reflect.Ptr:
			// deref pointer if field is not a pointer kind
			if fg.Type.Kind() != reflect.Ptr {
				valueof = valueof.Elem()
			}
		case reflect.Invalid:
			// for example we are here if fg.generator(ctx) returns (nil, nil)
			valueof = reflect.Zero(fg.Type)
		}

		// find field by index
		field := elem.FieldByIndex(fg.Index)
		// and assign value to field
		field.Set(valueof)
	}
	return nil
}

// MustSetFields calls SetFields and panics on error
func (f *Factory) MustSetFields(i interface{}, fieldGenFuncs ...FieldGenFunc) {
	if err := f.SetFields(i, fieldGenFuncs...); err != nil {
		panic(err)
	}
}

// Create makes a new instance
func (f *Factory) Create(fieldGenFuncs ...FieldGenFunc) (interface{}, error) {
	// allocate a new instance
	instance := f.new()
	if err := f.SetFields(instance.Interface(), fieldGenFuncs...); err != nil {
		return nil, err
	}
	return instance.Interface(), nil
}

// MustCreate creates or panics
func (f *Factory) MustCreate(fieldGenFuncs ...FieldGenFunc) interface{} {
	i, err := f.Create(fieldGenFuncs...)
	if err != nil {
		panic(err)
	}
	return i
}

// WithGen returns a function that generates an array of field generators,
// each of which has embedded check for field is present in the object being created and can be set.
func WithGen(g GeneratorFunc, fields ...string) FieldGenFunc {
	return func(sample reflect.Value) []fieldWithGen {
		gens := []fieldWithGen{}
		typ := sample.Elem().Type()
		for _, fieldName := range fields {
			sField, ok := typ.FieldByName(fieldName)
			if !ok {
				panic(fmt.Errorf("field %q not found in %s", fieldName, sample.Type().Name()))
			}

			// check that field exists in generated model
			field := sample.Elem().FieldByIndex(sField.Index)

			if !field.IsValid() {
				panic(fmt.Errorf("field %q is not valid in %s", fieldName, sample.Type().Name()))
			}

			// and can be set
			if !field.CanSet() {
				panic(fmt.Errorf("field %q can not be set in %s", fieldName, sample.Type().Name()))
			}

			gens = append(gens, fieldWithGen{&sField, g})
		}
		return gens
	}
}

// check if value is zero value
func isZero(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Slice, reflect.Ptr, reflect.Map:
		return val.IsNil()
	default:
		// otherwise allocate zero value
		zero := reflect.Zero(val.Type())
		// and compare
		return val.Interface() == zero.Interface()
	}
}

// ProtoGens takes a proto object and decomposes it into slice of field generators
// for each proto object field that has non-zero value.
func ProtoGens(proto interface{}) (fieldGenFuncs []FieldGenFunc) {
	typ := reflect.TypeOf(proto)

	// if proto object is non-zero type,
	// walk object fields and create field generator for each field with non-zero value
	val := reflect.ValueOf(proto)
	for i := 0; i < typ.NumField(); i++ {
		sField := typ.Field(i)
		// skip unexported fields. from godoc:
		// PkgPath is the package path that qualifies a lower case (unexported)
		// field name. It is empty for upper case (exported) field names.
		if sField.PkgPath != "" {
			continue
		}

		if fVal := val.Field(i); !isZero(fVal) {
			iVal := fVal.Interface()
			fGen := Use(iVal).For(sField.Name)
			if fieldGenFuncs != nil {
				fieldGenFuncs = append(fieldGenFuncs, fGen)
			} else {
				fieldGenFuncs = []FieldGenFunc{fGen}
			}
		}
	}
	return
}

// NewFactory is factory constructor
func NewFactory(proto interface{}, fieldGenFuncs ...FieldGenFunc) *Factory {
	typ := reflect.TypeOf(proto)

	if protogens := ProtoGens(proto); len(protogens) > 0 {
		// prepend field generators with proto generators if there are some
		fieldGenFuncs = append(protogens, fieldGenFuncs...)
	}

	// sample is used to validate during the factory construction process that all
	// provided fields exist in a given interface and can be set.
	sample := reflect.New(typ)
	fieldGens := make([]fieldWithGen, 0, len(fieldGenFuncs))

	// create field generators
	for _, makeFieldGen := range fieldGenFuncs {
		fieldGens = append(fieldGens, makeFieldGen(sample)...)
	}

	return &Factory{typ: typ, fieldGens: fieldGens}
}
