package reflection

import (
	"reflect"
	"strings"

	"user-service/internal/strcase"
)

// Options is the configuration structure optionally mutated via the [Variadic] constructor used throughout the package.
type Options struct {
	Lowercase bool // Lowercase forces lowercase on structure's fields. Defaults to true.
	Kebab     bool // Kebab will cast fields to kebab-style-casing. Defaults to true.
}

// Variadic represents a functional constructor for the [Options] type. Typical callers of Variadic won't need to perform
// nil checks as all implementations first construct an [Options] reference using packaged default(s).
type Variadic func(o *Options)

// options represents a default constructor.
func options() *Options {
	return &Options{
		Lowercase: true,
		Kebab:     true,
	}
}

func Map(obj interface{}, settings ...Variadic) map[string]interface{} {
	o := options()
	for _, option := range settings {
		option(o)
	}

	// 1. Create an empty map named result to store the fields and their values.
	result := make(map[string]interface{})

	// 2. Get the reflect.Value and reflect.Type of the input object.
	val := reflect.ValueOf(obj)

	// 3. If the input object is a pointer, dereference it to get the underlying value.
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	// 4. Iterate through the fields of the struct using a for loop.
	for i := 0; i < val.NumField(); i++ {
		// 5. For each field, get its name and kind (e.g., string, int, struct).
		field := typ.Field(i).Name
		if o.Kebab {
			field = strcase.ToKebab(field)
		}

		if !o.Lowercase {
			field = strings.ToLower(field)
		}

		// field value's kind
		kind := val.Field(i).Kind()

		var value interface{}

		// 6. If the field is a struct, recursively call structToMap to get the map representation of the nested struct.
		// Otherwise, get the field value directly.
		if kind == reflect.Struct {
			value = Map(val.Field(i).Interface())
		} else {
			value = val.Field(i).Interface()
		}

		// 7. Add the field name and value to the result map.
		result[field] = value
	}

	return result
}
