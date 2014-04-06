package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"
)

type Reader struct {
	*csv.Reader
}

func (r *Reader) Unmarshal(i interface{}) error {
	// TODO Error checking
	// The destination interface must of type slice
	sliceValue := reflect.ValueOf(i).Elem()

	// New struct
	// if v.Kind() == reflect.Ptr {
	//     v = v.Elem()
	// }

	// Get the type of the slice element
	elem := sliceValue.Type().Elem()

	// Read all
	for {
		record, err := r.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// Create a new slice element and append to the slice
		n := reflect.New(elem)
		newElem := n.Elem()

		err = setValue(record, &newElem)
		if err != nil {
			return err
		}

		sliceValue.Set(reflect.Append(sliceValue, newElem))
	}
	return nil
}

func (r *Reader) UnmarshalOne(i interface{}) error {
	// Get the next record from the reader
	record, err := r.Read()
	if err != nil {
		return err
	}

	// Get the value of the given interface
	value := reflect.ValueOf(i)

	// TODO Will indirect work or must the function be passed a pointer?
	elem := reflect.Indirect(value)

	return setValue(record, &elem)
}

// TODO How to persist the destination schema, with tags, etc...
func setValue(values []string, elem *reflect.Value) error {
	for i := 0; i < elem.NumField(); i += 1 {
		f := elem.Field(i)
		if !f.IsValid() || !f.CanSet() {
			return fmt.Errorf("Field %d cannot be set", i)
		}

		// TODO What about using a type switch instead?
		switch f.Kind() {
		case reflect.String:
			f.SetString(values[i])
		case reflect.Int64:
			// Attempt to convert the value to an int64
			v, err := strconv.ParseInt(values[i], 10, 64)
			if err != nil {
				// TODO wrap with the current field
				return err
			}
			f.SetInt(v)
		case reflect.Float64:
			// Attempt to convert the value to a float64
			v, err := strconv.ParseFloat(values[i], 64)
			if err != nil {
				// TODO wrap with the current field
				return err
			}
			f.SetFloat(v)
		case reflect.Bool:
			// Attempt to convert the value to a boolean
			v, err := strconv.ParseBool(values[i])
			if err != nil {
				// TODO wrap with the current field
				return err
			}
			f.SetBool(v)
		case reflect.Struct:
			switch f.Interface().(type) {
			case time.Time:
				// TODO Allow the layout to be set by the unmarshaler
				parsed, err := time.Parse(time.RFC3339, values[i])
				if err != nil {
					// TODO wrap with the current field
					return err
				}
				f.Set(reflect.ValueOf(parsed))
			default:
				return fmt.Errorf("unknown struct")
			}
		default:
			return fmt.Errorf("unknown type: %s", f.Kind())
		}
	}
	return nil
}

func NewReader(r io.Reader) *Reader {
	return &Reader{csv.NewReader(r)}
}
