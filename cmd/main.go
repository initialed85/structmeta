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
			// and extra a fresh zero'd instance of the type at that location (in this case, an anonymous struct)
			// that is completely divorced from the original type itself
			log.Printf("a new t.Grinch is %#+v", otherObject.Zero())
		}
	}
}
