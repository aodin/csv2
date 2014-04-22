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
	layouts map[int]string
}

func (r *Reader) setLayouts(v reflect.Type) {
	r.layouts = make(map[int]string)

	// TODO Only needs to be done for time.Time fields
	for i := 0; i < v.NumField(); i += 1 {
		f := v.Field(i)
		tag := f.Tag.Get("csv")
		if tag != "" {
			r.layouts[i] = tag
		}
	}
}

func (r *Reader) Unmarshal(i interface{}) error {
	// TODO Error checking
	// The destination interface must of type slice
	sliceValue := reflect.ValueOf(i).Elem()

	// Get the type of the slice element
	elem := sliceValue.Type().Elem()

	// Check the struct tags for any layouts
	r.setLayouts(elem)

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

		err = r.setValue(record, &newElem)
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

	t := reflect.TypeOf(i)
	if t.Kind() != reflect.Ptr {
		t = reflect.PtrTo(t)
	}
	r.setLayouts(t.Elem())

	return r.setValue(record, &elem)
}

// TODO How to persist the destination schema, with tags, etc...
func (r *Reader) setValue(values []string, elem *reflect.Value) error {
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
				// Check if an alternative layout should be used
				layout := r.layouts[i]
				if layout == "" {
					layout = time.RFC3339
				}
				parsed, err := time.Parse(layout, values[i])

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
	return &Reader{Reader: csv.NewReader(r)}
}
