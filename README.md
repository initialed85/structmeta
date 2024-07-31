# structmeta

# status: seems to work

This was an attempt to do some reflect-y stuff to save myself a bunch of typing to do with unloading one generated representation of an object into another (different) generated
representation of an object- I never really got it all the way to that point, but there's a function that returns a decent recursive introspection summary of any struct you pass it.

Here are the type definitions that represents the introspection summary:

```go
type StructFieldObject struct {
	Field    string
	Tag      string
	Embedded bool
	*Object
}

type Object struct {
	Name         string
	Kind         reflect.Kind
	Type         reflect.Type
	PointerValue *Object
	SliceValue   *Object
	MapKey       *Object
	MapValue     *Object
	StructFields []*StructFieldObject
	Parent       *Object
	Objects      []*Object
	zero         any
	handled      bool
}
```

So basically the hard work around reflection is done for you and you're left with an `Object` that represents the shape of the original struct.

You can also call `.Zero()` on any `Object` to get an zero / empty instance of the struct that the `Object` represents.

## Usage

Something like this:

```go
package main

import (
	"log"
	"time"

	"github.com/initialed85/structmeta/pkg/introspect"
)

// Thing is a big old complicated type definition
type Thing struct {
	*Thing
	NestedThing        *Thing
	AnotherNestedThing *Thing
	Cheese             float64           `json:"cheese"`
	Milk               map[string]string `json:"milk"`
	Pork               int64             `json:"pork"`
	Grinch             struct {
		Hi int64 `json:"hi"`
	} `json:"grinch"`
	Anon struct {
		Bilk []*struct {
			Dilk map[string]struct {
				Drilk time.Time     `json:"drilk"`
				Boo   time.Duration `json:"boo"`
			} `json:"dilk"`
		} `json:"bilk"`
	} `json:"anon"`
}

func main() {
	// so we create a concrete Thing
	thing := Thing{}

	// and then introspect a pointer to the Thing
	object, err := introspect.Introspect(&thing)
	if err != nil {
		log.Fatal(err)
	}

	// we can easily drill down to the struct fields of the struct behind the pointer
	for _, otherObject := range object.PointerValue.StructFields {
		// and find the field we're interested in
		if otherObject.Field == "Grinch" {
			// and extract a fresh zero'd instance of the type at that location (in this case, an anonymous struct)
			// that is completely divorced from the original type itself
			log.Printf("a new t.Grinch is %#+v", otherObject.Zero())
		}
	}
}
```

The eventual goal for the `.Zero()` function was some way to create new structs with certain mutations from an original struct, e.g.
maybe you want to automatically deref-or-zero any pointers in a structure so that it's all pointerless- maybe you have a special DB-style
object that tracks IsSet and a non-pointer Value or something and you want to wrap each object in that.
