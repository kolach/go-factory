# Factory

[![CircleCI](https://circleci.com/gh/kolach/go-factory.svg?style=svg)](https://circleci.com/gh/kolach/go-factory)

`go-factory` is a fixtures replacement based on user defined generator functions and includes features:

* Use your own or 3rd party libraries as value generators.
* Use your already defined factory as a generator for more complex scenarios.
* Derive your custom factory from existing one (aka factory inheritance).
* Call your factory recursively.
* Factory objects are thread safe.

Reed the chapters bellow to know how to do it.

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
// allocate the object yourself and set the fields using factory without loosing type information.
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

The syntax to register a field generator is either:

```
WithGen(<field-generator>, <field-name>[,firld-name2,field-name3...])
```

or DSL:

```
Use(<field-generator>).For(<field-name>[,field-name2,field-name3...])
```

For example:

```go
userFactory := NewFactory(
  User{},
  Use(name).For("FirstName"),
  Use(name).For("LastName"),
  Use(name).For("Username"),
)
```

Or in shorter form:

```go
userFactory := NewFactory(
  User{},
  Use(name).For("FirstName", "LastName", "Username"),
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

Order matters! The field generators are triggered in a order of registration. In the example above it is:
  1. FirstName
  2. LastName
  3. Username


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

Of cause it heavily uses reflection to work so use canonical field generator functions if
performance is critical:

```go
userFactory := NewFactory(
  User{},
  Use(func(Ctx) (interface{}, error) {
    return randomdata.Number(25, 50), nil
  }).For("Age"),
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

There is a shorter form if the same generator is used for multiple fields:

```go
userFactory := NewFactory(
  User{},
  Use(addressFactory).For("Address", "BillingAddress")
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
field generators on `Create` or `MustCreate` calls (SetFields and MustSetFields work the same way):

```go
user := userFactory.MustCreate(
  Use(randomdata.FirstName, randomdata.Female).For("FirstName"),
  Use(randomdata.LastName, randomdata.Female).For("LastName"),
).(*User)
```

```go
var user User
userFactory.MustSetFields(
  &user,
  Use(randomdata.FirstName, randomdata.Female).For("FirstName"),
  Use(randomdata.LastName, randomdata.Female).For("LastName"),
)
```

### Creating a new factory deriving from existing one

Overriding field generators on `(Must)SetFields`, `(Must)Create` invocation is not optimal for creating a big number of objects.
As each call with the list of overrides creates a new factory behind the scene and then calls the new factory's corresponding methods.
To make it clear, suppose we have a loop that fills a slice of user objects:

```go
users := make([]*User, 0, 1000)
for i := 0; i < 1000; i++ {
  var user User
  userFactory.MustSetFields(
    &user,
    Use(randomdata.FirstName, randomdata.Female).For("FirstName"),
    Use(randomdata.LastName, randomdata.Female).For("LastName"),
  )
  users = append(users, &user)
}
```

MustSetFields with list of generators is transformed into:

```go
users := make([]*User, 0, 1000)
for i := 0; i < 1000; i++ {
  var user User
  f := userFactory.Derive(
    Use(randomdata.FirstName, randomdata.Female).For("FirstName"),
    Use(randomdata.LastName, randomdata.Female).For("LastName"),
  )
  f.MustSetFields(&user)
  users = append(users, &user)
}
```

Where `Derive` is the method to create a new factory that inherits all its generator functions
from original factory overriding only a few of them.

So for performance reasons (to not create a new factory on each loop iteration) it's wise to rewrite
original code into:

```go
users := make([]*User, 0, 1000)
f := userFactory.Derive(
  Use(randomdata.FirstName, randomdata.Female).For("FirstName"),
  Use(randomdata.LastName, randomdata.Female).For("LastName"),
)
for i := 0; i < 1000; i++ {
  var user User
  f.MustSetFields(&user)
  users = append(users, &user)
}
```

## Prototype object

The first parameter to `NewFactory` function is actually the prototype for the object to produce. It's not necessary must
be an empty object like in all the examples above. Here is an example then it has some values in fields:

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

It's not only equals but represents what really happens inside `NewFactory` function call. The proto object fields are
walked and for each field with non-zero value a field generator is created.

## Recursion

You are totally free to use the factory recursively inside your custom generator functions. And here is how:

```go
type Node strict {
  Child *Node
}


factory := NewFactory(
  Node{},
  Use(func(ctx Ctx) (interface{}, error) {
    return ctx.Factory.Create()
  }).For("Child")
)
```

The context object has a self-reference to the factory object that can be used any time. But of cause you need to define
when to stop and exit the recursion. The factory above if used will lead to program exit with stack overflow.

It's up to you how and when you exit the recursive call. You can roll the dice `randomdata.Number(1, 6)` and exit if
the result is less than `3`:

```go
factory := NewFactory(
  Node{},
  Use(func(ctx Ctx) (interface{}, error) {
    self := ctx.Factory
    if randomdata.Number(1, 6) < 3 {
      return nil, nil
    }
    return self.Create()
  }).For("Child")
)
```

Or you can rely on `Factory.CallDepth()`. The method returns current call depth. It starts with 1 and increases on each
recursive call.

```go
factory := NewFactory(
  Node{},
  Use(func(ctx Ctx) (interface{}, error) {
    self := ctx.Factory
    if self.CallDepth() > randomdata.Number(3, 6) {
      return nil, nil
    }
    return self.Create()
  }).For("Child")
)
```

Let's go through more complex example. Suppose we have hierarchical tree model like:

```go
type Node struct {
  Parent   *Node   `json:"-"` // parent is excluded to avoid recursive calls
  Children []*Node `json:"children"`
  Name     string  `json:"name"`
}

```

And here is a factory that can generate it:

```go
factory = NewFactory(
  Node{},
  Use(randomdata.FirstName, randomdata.RandomGender).For("Name"),
  Use(func(ctx Ctx) (interface{}, error) {
    self := ctx.Factory

    if self.CallDepth() > randomdata.Number(2, 4) {
      // exit recursion if factory call depth is greater than [2, 4)
      return nil, nil
    }

    node := ctx.Instance.(*Node)    // current node that's being created
    size := randomdata.Number(1, 5) // number of children to make
    kids := make([]*Node, size)     // slice to store children nodes

    for i := 0; i < size; i++ {
      kids[i] = &Node{Parent: node}
      if err := self.SetFields(kids[i]); err != nil {
        return nil, err
      }
    }
    return kids, nil
  }).For("Children"),
)
```

## Thread safety

None of the methods of factory object modify the internal state so once created it's totally fine to use the factory
in multiple gorutines IF AND ONLY IF your generator functions are ALSO thread safe.

## Builder pattern to create a factory

The package also supports a builder pattern to create the factory but looks too verbose to use in comparison to DSL syntax used above.
Anyway here is an example:

```go
f := factory.NewBuilder(
  User{},
).Use("John").For(
  "FirstName",
).And("Smith", "Doe", "Milner").For(
  "LastName",
).And("mail@hotmail.com").For(
  "Email",
).And(randomdata.Number, 20, 50).For(
  "Age",
).And(true, false).For(
  "Married",
).Build()

```

Where `And` = `Use`.
