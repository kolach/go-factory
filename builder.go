package factory

// Builder is a struct that implements builder pattern to create a new factory
type Builder struct {
	proto interface{}
	fGens []FieldGenFunc
}

// ForBuilder is an interface with a single method `For` to bind
// generator function to fields.
type ForBuilder interface {
	For(field ...string) *Builder
}

type forBuilder struct {
	g GeneratorFunc
	b *Builder
}

func (f *forBuilder) For(fields ...string) *Builder {
	return f.b.WithGen(f.g, fields...)
}

// NewBuilder allocates a new factory builder
func NewBuilder(proto interface{}) *Builder {
	return &Builder{proto: proto, fGens: []FieldGenFunc{}}
}

// WithGen adds new gennerator to builder
func (b *Builder) WithGen(g GeneratorFunc, fields ...string) *Builder {
	b.fGens = append(b.fGens, WithGen(g, fields...))
	return b
}

// Use accepts generator function arguments and returns a ForBuilder interface
// with a single method `For` to bind generator to struct fiends
func (b *Builder) Use(i interface{}, args ...interface{}) ForBuilder {
	return &forBuilder{
		g: NewGenerator(i, args...),
		b: b,
	}
}

// And is synonim for Use
func (b *Builder) And(i interface{}, args ...interface{}) ForBuilder {
	return b.Use(i, args...)
}

// Build create a new factory
func (b *Builder) Build() *Factory {
	return NewFactory(b.proto, b.fGens...)
}
