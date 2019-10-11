package factory_test

import (
	"testing"

	. "github.com/kolach/go-factory"
)

// Factory with non-zero proto object
func BenchmarkProto(b *testing.B) {
	f := NewFactory(User{
		FirstName: "John",
		LastName:  "Smith",
		Username:  "john",
		Email:     "john@hotmail.com",
		Age:       30,
		Married:   false,
	})
	for i := 0; i < b.N; i++ {
		f.MustCreate()
	}
}

// Factory with zero proto object
func BenchmarkProtoEmpty(b *testing.B) {
	f := NewFactory(
		User{},
		Use("John").For("FirstName"),
		Use("Smith").For("LastName"),
		Use("john").For("Username"),
		Use("john@hotmail.com").For("Email"),
		Use(30).For("Age"),
		Use(false).For("Married"),
	)
	for i := 0; i < b.N; i++ {
		f.MustCreate()
	}
}
