package introspect

import (
	"fmt"
	"reflect"
)

type Path struct {
	Depth          int
	VisitedObjects map[*Object]struct{}
	Field          string
	Tag            reflect.StructTag
	Embedded       bool
}

var objectByType = make(map[reflect.Type]*Object)

type StructFieldObject struct {
	Field    string
	Tag      reflect.StructTag
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

func (o *Object) DebugFormat(paths ...*Path) string {
	var path *Path
	if len(paths) > 0 {
		path = paths[0]
	} else {
		path = &Path{
			Depth:          0,
			VisitedObjects: make(map[*Object]struct{}),
		}
	}

	getIndentPrefix := func() string {
		prefix := ""
		for i := 0; i < path.Depth; i++ {
			prefix += "  "
		}
		return prefix
	}

	objects := ""
	for i, object := range o.Objects {
		if i != 0 {
			objects += " | "
		}

		objects += object.Name
	}

	field := ""

	if path.Field != "" {
		field = fmt.Sprintf("%s: ", path.Field)
	}

	if field == "" {
		if path.Depth == 0 {
			field = "(root): "
		} else {
			field = "(anon): "
		}
	}

	tag := ""

	if path.Tag != "" {
		tag = fmt.Sprintf("%s ", path.Tag)
	}

	path.Field = ""
	path.Tag = ""
	path.Embedded = false

	_, visited := path.VisitedObjects[o]
	if visited && o.Kind == reflect.Struct {
		return fmt.Sprintf(
			"%s\t%s%s%s %s // recursion\n",
			o.Kind, getIndentPrefix(), field, o.Name, tag,
		)
	}

	output := fmt.Sprintf(
		"%s\t%s%s%s %s\n",
		o.Kind, getIndentPrefix(), field, o.Name, tag,
	)

	path.VisitedObjects[o] = struct{}{}

	if o.PointerValue != nil {
		path.Depth += 1
		path.Field = "(ptr value)"
		output += o.PointerValue.DebugFormat(path)
		path.Depth -= 1
	}

	if o.SliceValue != nil {
		path.Depth += 1
		path.Field = "(slice elem)"
		output += o.SliceValue.DebugFormat(path)
		path.Depth -= 1
	}

	if o.MapKey != nil {
		path.Depth += 1
		path.Field = "(map key)"
		output += o.MapKey.DebugFormat(path)
		path.Depth -= 1
	}

	if o.MapValue != nil {
		path.Depth += 1
		path.Field = "(map elem)"
		output += o.MapValue.DebugFormat(path)
		path.Depth -= 1
	}

	for _, structField := range o.StructFields {
		path.Depth += 1

		path.Field = structField.Field
		path.Tag = structField.Tag
		path.Embedded = structField.Embedded
		output += structField.DebugFormat(path)

		path.Depth -= 1
	}

	return output
}

func (o *Object) Zero() any {
	return o.zero
}

func introspect(t any, parent *Object, objects *[]*Object) (object *Object, err error) {
	if objects == nil {
		_objects := make([]*Object, 0)
		objects = &_objects
	}

	typeOf := reflect.TypeOf(t)

	object = objectByType[typeOf]
	if object != nil {
		return object, nil
	}

	if typeOf == nil {
		object = &Object{
			Name:         "any", // TODO: is this a safe assumption?
			Type:         typeOf,
			PointerValue: nil,
			SliceValue:   nil,
			MapKey:       nil,
			MapValue:     nil,
			StructFields: nil,
			Parent:       parent,
		}

		objectByType[typeOf] = object

		return object, nil
	}

kindSwitch:
	switch typeOf.Kind() {

	case reflect.Pointer:
		object = &Object{
			Kind:         typeOf.Kind(),
			Type:         typeOf,
			PointerValue: nil,
			SliceValue:   nil,
			MapKey:       nil,
			MapValue:     nil,
			StructFields: nil,
			Parent:       parent,
		}

		objectByType[typeOf] = object

		object.PointerValue, err = introspect(
			reflect.New(typeOf.Elem()).Elem().Interface(),
			object,
			objects,
		)
		if err != nil {
			break
		}

		if object.PointerValue != nil {
			object.Name = fmt.Sprintf("*%s", object.PointerValue.Name)
		}

		*objects = append(*objects, object)

	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.String:

		object = &Object{
			Name:         typeOf.Name(),
			Kind:         typeOf.Kind(),
			Type:         typeOf,
			PointerValue: nil,
			SliceValue:   nil,
			MapKey:       nil,
			MapValue:     nil,
			StructFields: nil,
			Parent:       parent,
			zero:         reflect.New(typeOf).Elem().Interface(),
			handled:      true,
		}

		objectByType[typeOf] = object

		*objects = append(*objects, object)

	case reflect.Array, reflect.Slice:

		object = &Object{
			Kind:         typeOf.Kind(),
			Type:         typeOf,
			PointerValue: nil,
			SliceValue:   nil,
			MapKey:       nil,
			MapValue:     nil,
			StructFields: nil,
			Parent:       parent,
		}

		objectByType[typeOf] = object

		object.SliceValue, err = introspect(
			reflect.New(typeOf.Elem()).Elem().Interface(),
			object,
			objects,
		)
		if err != nil {
			break
		}

		if object.SliceValue != nil {
			object.Name = fmt.Sprintf("[]%s", object.SliceValue.Name)
		}

		*objects = append(*objects, object)

	case reflect.Map:

		object = &Object{
			Kind:         typeOf.Kind(),
			Type:         typeOf,
			PointerValue: nil,
			SliceValue:   nil,
			MapKey:       nil,
			MapValue:     nil,
			StructFields: nil,
			Parent:       parent,
		}

		objectByType[typeOf] = object

		object.MapKey, err = introspect(
			reflect.New(typeOf.Key()).Elem().Interface(),
			object,
			objects,
		)
		if err != nil {
			break
		}

		objectByType[typeOf] = object

		object.MapValue, err = introspect(
			reflect.New(typeOf.Elem()).Elem().Interface(),
			object,
			objects,
		)
		if err != nil {
			break
		}

		if object.MapKey != nil && object.MapValue != nil {
			object.Name = fmt.Sprintf("map[%s]%s", object.MapKey.Name, object.MapValue.Name)
		}

		*objects = append(*objects, object)

	case reflect.Struct:

		object = &Object{
			Name:         typeOf.Name(),
			Kind:         typeOf.Kind(),
			Type:         typeOf,
			PointerValue: nil,
			SliceValue:   nil,
			MapKey:       nil,
			MapValue:     nil,
			StructFields: make([]*StructFieldObject, 0),
			Parent:       parent,
		}

		objectByType[typeOf] = object

		for i := 0; i < typeOf.NumField(); i++ {
			field := typeOf.Field(i)

			structFieldObject := &StructFieldObject{
				Field:    field.Name,
				Tag:      field.Tag,
				Embedded: field.Anonymous,
			}

			structFieldObject.Object, err = introspect(
				reflect.New(field.Type).Elem().Interface(),
				object,
				objects,
			)
			if err != nil {
				break kindSwitch
			}

			structFieldObject.Field = field.Name

			objectByType[field.Type] = structFieldObject.Object

			object.StructFields = append(object.StructFields, structFieldObject)
		}

		if object.Name == "" {
			object.Name = "struct { "
			for _, structField := range object.StructFields {
				object.Name += structField.Name + " "
			}
			object.Name += "}"
		}

		*objects = append(*objects, object)

	case reflect.Interface,
		reflect.Chan,
		reflect.UnsafePointer,
		reflect.Func,
		reflect.Invalid:
		// TODO: figure out what to do here
		// err = fmt.Errorf("unsupported kind %v (%#+v) for %#+v%v", typeOf.Kind().String(), typeOf.Kind(), t, parentSummary)

		// TODO: because this probably isn't right
		object = &Object{
			Name:         "any",
			Type:         reflect.TypeOf(new(interface{})),
			PointerValue: nil,
			SliceValue:   nil,
			MapKey:       nil,
			MapValue:     nil,
			StructFields: nil,
			Parent:       parent,
		}

		return object, nil

	default:
		err = fmt.Errorf("unhandled kind %v (%#+v) for %#+v", typeOf.Kind().String(), typeOf.Kind(), t)
	}

	if object == nil && err == nil {
		err = fmt.Errorf("object unexpectly nil after kind switch (kind: %s, type: %v)", typeOf.Kind(), typeOf)
	}

	if err != nil {
		return nil, err
	}

	object.Objects = *objects

	for _, otherObject := range object.Objects {
		switch otherObject.Kind {

		case reflect.Pointer:
			if otherObject.PointerValue != nil && otherObject.PointerValue.zero != nil {
				otherObject.zero = reflect.New(reflect.PointerTo(reflect.TypeOf(otherObject.PointerValue.zero))).Elem().Interface()
				otherObject.handled = true
			}

		case reflect.Bool,
			reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64,
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64,
			reflect.Uintptr,
			reflect.Float32,
			reflect.Float64,
			reflect.Complex64,
			reflect.Complex128,
			reflect.String:

		case reflect.Array, reflect.Slice:
			if otherObject.SliceValue != nil && otherObject.SliceValue.zero != nil {
				otherObject.zero = reflect.New(reflect.SliceOf(reflect.TypeOf(otherObject.SliceValue.zero))).Elem().Interface()
				otherObject.handled = true
			}

		case reflect.Map:
			if otherObject.MapKey != nil && otherObject.MapKey.zero != nil && otherObject.MapValue != nil && otherObject.MapValue.zero != nil {
				otherObject.zero = reflect.New(reflect.MapOf(
					reflect.TypeOf(otherObject.MapKey.zero),
					reflect.TypeOf(otherObject.MapValue.zero),
				)).Elem().Interface()
				otherObject.handled = true
			}

		case reflect.Struct:
			structFields := make([]reflect.StructField, 0)
			for _, structField := range otherObject.StructFields {
				if structField.zero == nil {
					continue
				}

				structFields = append(
					structFields,
					reflect.StructField{
						Name:    structField.Field,
						Type:    reflect.TypeOf(structField.zero),
						Tag:     structField.Tag,
						PkgPath: "introspect",
					},
				)
			}

			// TODO: not only is this problematic, I think its slow- let's hope I don't need it somewhere
			// func() {
			// 	defer func() {
			// 		r := recover()
			// 		if r != nil {
			// 			log.Printf("warning: recovered from a panic that is probably related to the fact that the anonymous struct we're trying to create is too large")
			// 			log.Printf("warning: the impact of this is that the Object.Zero() call for the output of Introspect(%#+v) will unexpectedly return nil", t)
			// 		}
			// 	}()

			// 	a := reflect.StructOf(structFields)
			// 	b := reflect.New(a)
			// 	c := b.Elem()
			// 	d := c.Interface()

			// 	otherObject.zero = d
			// }()

			otherObject.zero = reflect.New(reflect.TypeOf(t)).Elem().Interface()

			otherObject.handled = true
		}
	}

	return object, nil
}

func Introspect(t any) (*Object, error) {
	o, err := introspect(t, nil, nil)
	if err != nil {
		return nil, err
	}

	return o, nil
}
