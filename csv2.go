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
	ErrNotPointer = errors.New("csv2: non-pointer receiver")
	ErrNotSlice   = errors.New("csv2: receiver must be a slice of structs")
	ErrNotStruct  = errors.New("csv2: receiver must be a struct")
	ErrUnwritable = errors.New("csv2: writers accept struct or struct slices")
)

// GetFieldNames returns a string array of the given interface's field names
// if the given interface is a struct or slice of structs
func GetFieldNames(i interface{}) ([]string, error) {
	// Given interface must be a struct or a slice of structs
	// TODO Pointers!?
	value := reflect.Indirect(reflect.ValueOf(i))
	var elem reflect.Type
	switch value.Kind() {
	case reflect.Struct:
		elem = value.Type()
	case reflect.Slice:
		elem = value.Type().Elem()
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Struct {
			return nil, ErrUnwritable
		}
	default:
		return nil, ErrUnwritable
	}
	// Get the names of the struct fields
	fields := make([]string, elem.NumField())
	for index := 0; index < elem.NumField(); index += 1 {
		fields[index] = elem.Field(index).Name
	}
	return fields, nil
}

// setLayout checks the given struct type for any "csv" tags.
// This layout is used for alternative parse formats.
func setLayout(v reflect.Type) map[int]string {
	layout := make(map[int]string)
	for i := 0; i < v.NumField(); i += 1 {
		f := v.Field(i)
		tag := f.Tag.Get("csv")
		if tag != "" {
			layout[i] = tag
		}
	}
	return layout
}

// Reader wraps the csv.Reader and adds a map of csv struct tags.
type Reader struct {
	*csv.Reader
	layout map[int]string
}

// Unmarshal reads the entire Reader into the given destination.
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
	if elem.Kind() != reflect.Struct {
		return ErrNotSlice
	}

	// Check the struct tags for any custom csv layout tags
	// TODO Check if already set?
	r.layout = setLayout(elem)

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

// UnmarshalOne reads a single row of the Reader into the given struct.
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
	r.layout = setLayout(t.Elem())
	return r.setValue(record, &elem)
}

// Set the values of the given struct with the reflect package.
// Fields are processed in sequential order.
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
				layout := r.layout[i]
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

// NewReader returns a new csv2 Reader by wrapping a csv Reader.
func NewReader(r io.Reader) *Reader {
	return &Reader{Reader: csv.NewReader(r)}
}

// Writer wraps the csv.Writer and adds a map of csv struct tags.
type Writer struct {
	*csv.Writer
	layout map[int]string
}

// WriteHeader will write the names of the underlying struct fields as a row.
// It accepts a struct or a slice of structs.
func (w *Writer) WriteHeader(i interface{}) error {
	headers, err := GetFieldNames(i)
	if err != nil {
		return err
	}
	if err = w.Write(headers); err != nil {
		return err
	}
	return nil
}

func (w *Writer) getStrings(elem reflect.Value) ([]string, error) {
	output := make([]string, elem.NumField())
	for i := 0; i < elem.NumField(); i += 1 {
		f := elem.Field(i)

		// TODO What about using a type switch instead? benchmark it.
		switch f.Kind() {
		case reflect.String:
			output[i] = f.String()
		case reflect.Int64:
			// TODO additional base output
			output[i] = strconv.FormatInt(f.Int(), 10)
		case reflect.Float64:
			// TODO additional formats, precision
			output[i] = strconv.FormatFloat(f.Float(), 'f', -1, 64)
		case reflect.Bool:
			// Attempt to convert the value to a boolean
			output[i] = strconv.FormatBool(f.Bool())
		case reflect.Struct:
			switch f.Interface().(type) {
			case time.Time:
				// Get the underlying time
				t := f.Interface().(time.Time)

				// Check if an alternative layout should be used
				layout := w.layout[i]
				if layout == "" {
					layout = time.RFC3339
				}
				output[i] = t.Format(layout)
			default:
				return output, fmt.Errorf(
					"csv2: unsupported struct for field %d",
					i,
				)
			}
		default:
			return output, fmt.Errorf(
				"csv2: unsupported type %s for field %d",
				f.Kind(),
				i,
			)
		}
	}
	return output, nil
}

// Marshal writes a slice of structs to the Writer.
// The destination interface must of pointer of type slice.
func (w *Writer) Marshal(i interface{}) error {
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

	// Check the struct tags for any custom csv layout tags
	w.layout = setLayout(elem)

	// Read all
	for index := 0; index < sliceValue.Len(); index += 1 {
		s, err := w.getStrings(sliceValue.Index(index))
		if err != nil {
			return err
		}
		if err = w.Write(s); err != nil {
			return err
		}
	}

	// csv writer is buffered
	w.Flush()
	return nil
}

// NewWriter returns a new csv2 Writer by wrapping a csv Writer.
func NewWriter(r io.Writer) *Writer {
	return &Writer{Writer: csv.NewWriter(r)}
}
