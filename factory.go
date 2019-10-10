package factory

import (
	"fmt"
	"reflect"
)

// Ctx is the context in wich the field value is being generated
type Ctx struct {
	Field    string      // current field name for which the value is generated
	Instance interface{} // the result instance to that the field belongs
	Factory  *Factory    // the reference to the Factory
}

// GeneratorFunc describes field generator signatures
type GeneratorFunc func(ctx Ctx) (interface{}, error)

// fieldGen is a tuple that keeps together field name and generator function.
type fieldGen struct {
	fieldName string
	generator GeneratorFunc
}

// Factory is the work horse of the package that produces instances
type Factory struct {
	typ       reflect.Type // type information about generated instances
	fieldGens []fieldGen   // field / generator tuples
}

// Derive creates a new factory overriding fields generators
// with the list provided.
func (f *Factory) Derive(fieldGenFuncs ...FieldGenFunc) *Factory {
	// collect fields generator to override
	overrides := make(map[string]GeneratorFunc)
	sample := f.new()
	for _, fieldGenFunc := range fieldGenFuncs {
		for _, fg := range fieldGenFunc(sample) {
			overrides[fg.fieldName] = fg.generator
		}
	}

	// allocate a new factory
	newF := &Factory{typ: f.typ, fieldGens: make([]fieldGen, len(f.fieldGens))}
	// copy or override original field generators
	for i, fg := range f.fieldGens {
		if gen, ok := overrides[fg.fieldName]; ok {
			fg.generator = gen
		}
		newF.fieldGens[i] = fg
	}
	return newF
}

func (f *Factory) new() reflect.Value {
	return reflect.New(f.typ)
}

// SetFields fills in the struct instance fields
func (f *Factory) SetFields(i interface{}, fieldGenFuncs ...FieldGenFunc) error {
	return f.setFields(reflect.ValueOf(i), fieldGenFuncs...)
}

// MustSetFields calls SetFields and panics on error
func (f *Factory) MustSetFields(i interface{}, fieldGenFuncs ...FieldGenFunc) {
	if err := f.SetFields(i, fieldGenFuncs...); err != nil {
		panic(err)
	}
}

func (f *Factory) setFields(instance reflect.Value, fieldGenFuncs ...FieldGenFunc) error {
	// create execution context
	elem, i := instance.Elem(), instance.Interface()

	ctx := Ctx{Instance: i, Factory: f}

	for _, fg := range f.fieldGens {
		// bind field name o context
		ctx.Field = fg.fieldName
		// find field by name
		field := elem.FieldByName(fg.fieldName)
		// generate field value
		val, err := fg.generator(ctx)
		if err != nil {
			return err
		}

		// assign value to field
		valueof := reflect.ValueOf(val)
		// deref pointer if field value is not a pointer
		if field.Kind() != reflect.Ptr && valueof.Kind() == reflect.Ptr {
			valueof = valueof.Elem()
		}

		field.Set(valueof)
	}
	return nil
}

// Create makes a new instance
func (f *Factory) Create(fieldGenFuncs ...FieldGenFunc) (interface{}, error) {
	if len(fieldGenFuncs) > 0 {
		return f.Derive(fieldGenFuncs...).Create()
	}
	// allocate a new instance
	instance := f.new()
	if err := f.setFields(instance, fieldGenFuncs...); err != nil {
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

// FieldGenFunc is the signature of field generator factory.
type FieldGenFunc func(sample reflect.Value) []fieldGen

// WithGen adds generator function to factory.
// WithGen returns a function that generates an array of field generators,
// each of which has embedded check for field is present in the object being created and can be set.
func WithGen(g GeneratorFunc, fields ...string) FieldGenFunc {
	return func(sample reflect.Value) []fieldGen {
		gens := []fieldGen{}
		for _, fieldName := range fields {
			// check that field exists in generated model
			field := sample.Elem().FieldByName(fieldName)
			if !field.IsValid() {
				panic(fmt.Errorf("field %q not found in %s", fieldName, sample.Type().Name()))
			}
			// and can be set
			if !field.CanSet() {
				panic(fmt.Errorf("field %q can not be set in %s", fieldName, sample.Type().Name()))
			}
			gens = append(gens, fieldGen{fieldName, g})
		}
		return gens
	}
}

// NewFactory is factory constructor
func NewFactory(proto interface{}, fieldGenFuncs ...FieldGenFunc) *Factory {
	typ := reflect.TypeOf(proto)

	f := &Factory{typ: typ}

	// sample is used to validate during the factory construction process that all
	// provided fields exist in a given model and can be set.
	sample := f.new()

	if proto != reflect.Zero(typ).Interface() {
		// if proto object is non-zero type,
		// walk object fields and create field generator for each field with non-zero value
		val := reflect.ValueOf(proto)
		for i := 0; i < typ.NumField(); i++ {
			sField := typ.Field(i)
			fVal := val.Field(i).Interface()
			if fVal != reflect.Zero(sField.Type).Interface() {
				makeFieldGen := Use(fVal).For(sField.Name)
				f.fieldGens = append(f.fieldGens, makeFieldGen(sample)...)
			}
		}

	}

	// create field generators
	for _, makeFieldGen := range fieldGenFuncs {
		f.fieldGens = append(f.fieldGens, makeFieldGen(sample)...)
	}
	return f
}
