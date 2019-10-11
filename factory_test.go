package factory_test

import (
	"errors"
	"strings"

	randomdata "github.com/Pallinder/go-randomdata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"

	. "github.com/kolach/go-factory"
	. "github.com/kolach/gomega-matchers"
)

type Address struct {
	City   string
	Street string

	// to check unexported field do not break or influences on anything
	i int
	s string
}

type User struct {
	ID        uuid.UUID
	Username  string
	FirstName string
	LastName  string
	Email     string
	Age       int
	Married   bool
	Address   Address
	Comment   string

	// to check unexported field do not break or influences on anything
	i int
	s string
}

var _ = Describe("Factory", func() {
	var (
		userFact *Factory
		addrFact *Factory
	)

	BeforeEach(func() {
		// capitalize username
		firstName := func(ctx Ctx) (interface{}, error) {
			return strings.Title(ctx.Instance.(*User).Username), nil
		}

		// username + '@' + domain
		email := func(domain string) GeneratorFunc {
			return func(ctx Ctx) (interface{}, error) {
				user := ctx.Instance.(*User)
				return user.Username + "@" + domain, nil
			}
		}

		addrFact = NewFactory(
			Address{},
			Use(SeqSelect("CDMX", "Playa del Carmen")).For("City"),
			Use(func(ctx Ctx) (interface{}, error) {
				addr := ctx.Instance.(*Address)
				switch addr.City {
				case "CDMX":
					return "Mexicali", nil
				case "Playa del Carmen":
					return "Paseo Xaman Ha", nil
				default:
					return "Benito Juares", nil
				}
			}).For("Street"),
		)

		userFact = NewFactory(
			User{},
			Use(uuid.NewV4).For("ID"),
			Use("john", "james", "bob", "paul").For("Username"),
			Use(firstName).For("FirstName"),
			Use("Doe", "Smith", "Roy").For("LastName"),
			Use(email("6river.com")).For("Email"),
			Use(true).For("Married"),
			Use(randomdata.Number, 20, 25).For("Age"),
			Use(addrFact).For("Address"),
		)
	})

	It("should set instance fields", func() {
		var u User
		err := userFact.SetFields(&u)
		Ω(err).Should(BeNil())
		Ω(u.ID).ShouldNot(BeNil())
		Ω(u.Username).Should(BelongTo("john", "james", "bob", "paul"))
		Ω(u.FirstName).Should(Equal(strings.Title(u.Username)))
		Ω(u.LastName).Should(BelongTo("Doe", "Smith", "Roy"))
		Ω(u.Email).Should(Equal(u.Username + "@6river.com"))
		Ω(u.Married).Should(BeTrue())
		Ω(u.Age).Should(And(BeNumerically(">=", 20), BeNumerically("<", 25)))
		Ω(u.Address.City).Should(Equal("CDMX"))
		Ω(u.Address.Street).Should(Equal("Mexicali"))
	})

	It("should copy prototype properties", func() {
		proto := User{Married: true, Age: 45, FirstName: "Nick", i: 5}
		userFact := NewFactory(proto, Use("Smith").For("LastName"))

		var user User
		err := userFact.SetFields(&user)

		Ω(err).Should(BeNil())
		Ω(user.Married).Should(Equal(proto.Married))
		Ω(user.Age).Should(Equal(proto.Age))
		Ω(user.FirstName).Should(Equal(proto.FirstName))
		Ω(user.LastName).Should(Equal("Smith"))
	})

	It("should create instances of given type", func() {
		u, ok := userFact.MustCreate().(*User)
		Ω(ok).Should(BeTrue())
		Ω(u.ID).ShouldNot(BeNil())
		Ω(u.Username).Should(BelongTo("john", "james", "bob", "paul"))
		Ω(u.FirstName).Should(Equal(strings.Title(u.Username)))
		Ω(u.LastName).Should(BelongTo("Doe", "Smith", "Roy"))
		Ω(u.Email).Should(Equal(u.Username + "@6river.com"))
		Ω(u.Married).Should(BeTrue())
		Ω(u.Age).Should(And(BeNumerically(">=", 20), BeNumerically("<", 25)))
		Ω(u.Address.City).Should(Equal("CDMX"))
		Ω(u.Address.Street).Should(Equal("Mexicali"))
	})

	It("should allow override existing generators on create", func() {
		u, ok := userFact.MustCreate(
			Use("jane").For("Username"), // override username
			Use(false).For("Married"),   // override married status
			Use(10).For("Age"),          // override age
		).(*User)
		Ω(ok).Should(BeTrue())
		// now check it all
		Ω(u.Username).Should(Equal("jane"))
		Ω(u.FirstName).Should(Equal("Jane")) // check dependend field
		Ω(u.Married).Should(BeFalse())
		Ω(u.Age).Should(Equal(10))
	})

	It("should allow to define a new generator on create", func() {
		u, ok := userFact.MustCreate(
			Use("jane").For("Username"),        // override username
			Use("Blahblahblah").For("Comment"), // new generator
		).(*User)
		Ω(ok).Should(BeTrue())
		// now check it all
		Ω(u.Username).Should(Equal("jane"))
		Ω(u.Comment).Should(Equal("Blahblahblah")) // check new generator
	})

	It("should support generator funcs that return error as second value", func() {
		_, err := userFact.Create(
			Use(func() (string, error) {
				return "John", errors.New("boom")
			}).For("FirstName"),
		)
		Ω(err).ShouldNot(BeNil())
	})

	It("should panic if second return value is not of error type", func() {
		call := func() {
			userFact.Create(
				Use(func() (string, int) {
					return "John", 1
				}).For("FirstName"),
			)
		}
		Ω(call).Should(Panic())
	})

	It("should panic if generator function return more than 2 values", func() {
		call := func() {
			userFact.Create(
				Use(func() (string, int, bool) {
					return "John", 1, true
				}).For("FirstName"),
			)
		}
		Ω(call).Should(Panic())
	})

	It("should panic if generator function arity is not satisfied", func() {
		call := func() {
			userFact.Create(
				Use(func(i int, b bool) (string, int, bool) {
					return "John", 1, true
				}, 4).For("FirstName"),
			)
		}
		Ω(call).Should(Panic())
	})

	It("should panic on attempt to set unexported field", func() {
		Ω(func() { userFact.Create(Use(1).For("i")) }).Should(Panic())
	})

	Describe("MustCreate and MustSetFields", func() {
		It("should panic on error", func() {
			Ω(func() {
				userFact.MustCreate(
					Use(func(ctx Ctx) (interface{}, error) {
						return nil, errors.New("boom")
					}).For("FirstName"),
				)
			}).Should(Panic())

			Ω(func() {
				var u User
				userFact.MustSetFields(
					&u,
					Use(func(ctx Ctx) (interface{}, error) {
						return nil, errors.New("boom")
					}).For("FirstName"),
				)
			}).Should(Panic())

		})
	})
})
