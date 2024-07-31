package introspect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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

func TestIntrospect(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		v := "Hello world"
		o, err := Introspect(v)
		require.NoError(t, err)
		t.Logf("\n\n%s\n", o.DebugFormat())
		t.Logf("\n\n%#+v\n", o.zero)
	})

	t.Run("PointerToString", func(t *testing.T) {
		v := "Hello world"
		o, err := Introspect(&v)
		require.NoError(t, err)
		t.Logf("\n\n%s\n", o.DebugFormat())
		t.Logf("\n\n%#+v\n", o.zero)
	})

	t.Run("SliceOfString", func(t *testing.T) {
		v := []string{"Hello", "world"}
		o, err := Introspect(v)
		require.NoError(t, err)
		t.Logf("\n\n%s\n", o.DebugFormat())
		t.Logf("\n\n%#+v\n", o.zero)
	})

	t.Run("SliceOfPointerToString", func(t *testing.T) {
		hello := "hello"
		world := "world"
		v := []*string{&hello, &world}
		o, err := Introspect(v)
		require.NoError(t, err)
		t.Logf("\n\n%s\n", o.DebugFormat())
		t.Logf("\n\n%#+v\n", o.zero)
	})

	t.Run("PointerToSliceOfPointerToString", func(t *testing.T) {
		hello := "hello"
		world := "world"
		v := &[]*string{&hello, &world}
		o, err := Introspect(v)
		require.NoError(t, err)
		t.Logf("\n\n%s\n", o.DebugFormat())
		t.Logf("\n\n%#+v\n", o.zero)
	})

	t.Run("PointerToMapOfPointerToStringPointerToString", func(t *testing.T) {
		hello := "hello"
		world := "world"
		v := &map[*string]*string{&hello: &world}
		o, err := Introspect(v)
		require.NoError(t, err)
		t.Logf("\n\n%s\n", o.DebugFormat())
		t.Logf("\n\n%#+v\n", o.zero)
	})

	t.Run("BigOldMess", func(t *testing.T) {
		v := Thing{}
		o, err := Introspect(&v)
		require.NoError(t, err)
		t.Logf("\n\n%s\n", o.DebugFormat())
		t.Logf("\n\n%#+v\n", o.zero)
	})

	t.Run("Simple", func(t *testing.T) {
		v := struct {
			Field string
		}{}

		o1, err := Introspect(v)
		require.NoError(t, err)
		t.Logf("\n\n%s\n", o1.DebugFormat())

		o2, err := Introspect(v)
		require.NoError(t, err)
		t.Logf("\n\n%s\n", o2.DebugFormat())
	})
}
