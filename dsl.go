package factory

// FieldGeneratorBuilder is DSL build chain pattern
type FieldGeneratorBuilder struct {
	generator GeneratorFunc
}

// Use this value/function/factory For that field(s)
func Use(i interface{}, args ...interface{}) (g FieldGeneratorBuilder) {
	return FieldGeneratorBuilder{NewGenerator(i, args...)}
}

// For creates FieldGenFunc for each provided field
func (g FieldGeneratorBuilder) For(field ...string) FieldGenFunc {
	return WithGen(g.generator, field...)
}
