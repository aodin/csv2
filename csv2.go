package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"
)

var (
	ErrNotPointer = errors.New("csv2: destination must be a pointer")
	ErrNotSlice   = errors.New("csv2: destination must be a slice")
	ErrNotStruct  = errors.New("csv2: destination must be a struct")
)

type Reader struct {
	*csv.Reader
	layouts map[int]string
}

// Check the given struct type for any "csv" tags.
// These layouts are used for alternative parse formats.
func (r *Reader) setLayouts(v reflect.Type) {
	r.layouts = make(map[int]string)
	for i := 0; i < v.NumField(); i += 1 {
		f := v.Field(i)
		tag := f.Tag.Get("csv")
		if tag != "" {
			r.layouts[i] = tag
		}
	}
}

// Unmarshal the entire Reader into the given destination.
// The destination interface must of pointer of type slice.
func (r *Reader) Unmarshal(i interface{}) error {
	sv := reflect.ValueOf(i)
	if sv.Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	sliceValue := sv.Elem()
	if sliceValue.Kind() != reflect.Slice {
		return ErrNotSlice
	}

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

// Unmarshal a single row of the Reader into the given struct.
// The destination interface must of pointer of type struct.
func (r *Reader) UnmarshalOne(i interface{}) error {
	// Get the value of the given interface
	value := reflect.ValueOf(i)
	if value.Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	elem := value.Elem()
	if elem.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	// Get the next record from the reader
	record, err := r.Read()
	if err != nil {
		return err
	}

	// Get the type of the interface to check for layouts
	t := reflect.TypeOf(i)
	r.setLayouts(t.Elem())
	return r.setValue(record, &elem)
}

// Set the values of the given struct with the reflect package.
// Fields are processed in
func (r *Reader) setValue(values []string, elem *reflect.Value) error {
	// TODO wrap the errors with the current field
	for i := 0; i < elem.NumField(); i += 1 {
		f := elem.Field(i)
		if !f.IsValid() || !f.CanSet() {
			return fmt.Errorf("csv2: field %d cannot be set", i)
		}

		// TODO What about using a type switch instead? benchmark it.
		switch f.Kind() {
		case reflect.String:
			f.SetString(values[i])
		case reflect.Int64:
			// Attempt to convert the value to an int64
			v, err := strconv.ParseInt(values[i], 10, 64)
			if err != nil {
				return err
			}
			f.SetInt(v)
		case reflect.Float64:
			// Attempt to convert the value to a float64
			v, err := strconv.ParseFloat(values[i], 64)
			if err != nil {
				return err
			}
			f.SetFloat(v)
		case reflect.Bool:
			// Attempt to convert the value to a boolean
			v, err := strconv.ParseBool(values[i])
			if err != nil {
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
					return err
				}
				f.Set(reflect.ValueOf(parsed))
			default:
				return fmt.Errorf(
					"csv2: unknown destination struct for field %d",
					i,
				)
			}
		default:
			return fmt.Errorf(
				"csv2: unsupported type %s for field %d",
				f.Kind(),
				i,
			)
		}
	}
	return nil
}

func NewReader(r io.Reader) *Reader {
	return &Reader{Reader: csv.NewReader(r)}
}
