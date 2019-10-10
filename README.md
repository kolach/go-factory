# Factory

Library to create various test objects.

## Usage

Suppose we have a user model like:

```go
type User struct {
  Username  string
  FirstName string
  LastName  string
  Email     string
  Age       uint8
  Married   bool
}
```

Then the minimum code to create a factory is:

```go
import 	(
  . "github.com/kolach/go-factory"
)


userFact := NewFactory(User{}}
```

And then you can use it like:

```go
// allocate the object yourself and set the fields using factory
var user User
if err := userFact.SetFields(&user); err != nil {
  panic(err)
}

// or use MustSetFields that panics on error
userFact.MustSetFields(&user)

// or use factory Create method and cast the result to *User
if i, err := userFact.Create(); err == nil {
  user := i.(*User)
}

// or, if it's OK to panic on factory failure
user := userFact.MustCreate().(*User)
```

Factory's `Create` and `MustCreate` methods return `interface{}` so you need to cast it to `*User` to use.

The factory above creates a user with empty fields which is pretty useless.
To assign some values to the fields the field generators must be registered in the factory.

### Field generators

The syntax to register a field generator is straight forward: `Use(<field-generator>).For(<field-name>)`.

```go
userFactory := NewFactory(
  User{},
  Use(name).For("FirstName"),
  Use(name).For("LastName"),
  Use(name).For("Username"),
)
```

Where `name` is a reference to generator function. In most canonical form it has a signature:

```go
type GeneratorFunc func(ctx Ctx) (interface{}, error)
```

Where `ctx` is the context in which the field value is being generated:

```go
type Ctx struct {
	Field    string      // current field name for which the value is generated
	Instance interface{} // the result instance to that the field belongs
	Factory  *Factory    // the reference to the Factory
}

```

So here is how our `name` generator may look like:

```go
func name(ctx Ctx) (interface{}, error) {
  switch ctx.Field {
  case "FirstName":
    return "John", nil
  case "Username":
    u := ctx.Instance.(*User)
    return strings.ToLower(u.FirstName), nil
  default:
    // otherwise
    return "Smith", nil
  }
}
```

In many cases you do not need to write generator function. The `Use` function is smart enough to generate it for you.
Let's now review alternative options:

#### List of values as generator function

The list of values produce a generator function that randomly selects an option from the list each time the generator is invoked.

```go
Use("John", "Jack", "Joe").For("FirstName")
Use(true, false).For("Married")
// single value is also an option:
Use(true).For("Married")
```

#### Functions as field generators

Any function that returns some value or value and error are good to use as generators. If the function need the input
arguments, they can be enlisted next to the function name.

Let's use some generator from [go-randomdata](https://github.com/Pallinder/go-randomdata) package. For example
there is a `randomdata.Number` function:

```go
package randomdata

// Number returns a random number, if only one integer (n1) is supplied it returns a number in [0,n1)
// if a second argument is supplied it returns a number in [n1,n2)
func Number(numberRange ...int) int {
  ...
}
```

And here is how we can use it in a factory:

```go
userFactory := NewFactory(
  User{},
  Use(randomdata.Number, 25, 50).For("Age"),
)
```

Here is another sample:

```go
userFactory := NewFactory(
  User{},
  Use(randomdata.FirstName, randomdata.Male).For("FirstName"),
  Use(randomdata.LastName, randomdata.Male).For("LastName"),
)

```

#### Another factory as a field generator

Suppose our `User` model has an `Address` field with is a struct with fields:

```go
type Address struct {
  Country string
  State   string
  Street  string
}


type User struct {
  ...
  Address Address
  BillingAddress *Address
}
```

We now can create address factory and use it in user factory to fill in Address and BillingAddress fields:

```go
addressFactory := NewFactory(
  Address{},
  Use(randomdata.Country, randomdata.FullCountry).For("Country"),
  Use(randomdata.State, randomdata.Large).For("State"),
  Use(randomdata.City).For("City"),
  Use(randomdata.Street).For("Street"),
)

userFactory := NewFactory(
  User{},
  Use(addressFactory).For("Address")
  Use(addressFactory).For("BillingAddress")
)
```

### Overriding field generators

Suppose we have a user factory:

```go
userFactory := NewFactory(
  User{},
  Use(randomdata.FirstName, randomdata.Male).For("FirstName"),
  Use(randomdata.LastName, randomdata.Male).For("LastName"),
  Use(randomdata.Number, 25, 50).For("Age"),
  Use(true, false).For("Married"),
  Use(randomdata.Email).For("Email"),
  Use(addressFactory).For("Address")
  Use(addressFactory).For("BillingAddress")
)
```

And we need to generate a user with female first and last names. We can easily do it overriding the
field generators on `Create` or `MustCreate` calls:

```go
user := userFactory.MustCreate(
  Use(randomdata.FirstName, randomdata.Female).For("FirstName"),
  Use(randomdata.LastName, randomdata.Female).For("LastName"),
).(*User)
```

### Creating a new factory deriving from existing one

A new factory may be created deriving from existing one. Suppose we have a factory:

```go
addressFactory := NewFactory(
  Address{},
  Use(randomdata.Country, randomdata.FullCountry).For("Country"),
  Use(randomdata.State, randomdata.Large).For("State"),
  Use(randomdata.City).For("City"),
  Use(randomdata.Street).For("Street"),
)
```

And we want to have such a factory that generates US, New York addresses. It can be done with factory's `Derive` method:

```go
nyAddressFactory := addressFactory.Derive(
  Use("US").For("Country"),
  Use("New York").For("State"),
  Use("New York").For("City"),
)

nyAddress := nyAddressFactory.MustCreate().(*Address)
```

## Prototype object

The first parameter to `NewFactory` function is actually the prototype for the object to produce. It's not necessary must
be an emprt object like in all the examples above. Here is an example then it has some values in fields:

```go
userFactory := NewFactory(
  User{Age: 32, Married: true},
)
```

Like you can guess all the users produced by the factory will have `Age = 32` and `Married = true`.
The factory above equals to:

```go
userFactory := NewFactory(
  User{},
  Use(32).For("Age"),
  Use(true).For("Married"),
)
```

It's worth to mention that using the prototype object makes the factory work slower.
The version that uses field generators is currently a prefferable way to set field values.
