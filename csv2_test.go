package csv

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"
)

var example = []byte(`ID,NAME,ABBREV,POPULATION,GDP (trillions),FOUNDED,FREEDOM?
2,"United States","US",317808000,17.438,1776-07-04T00:00:00Z,true
3,"Canada","CA",35344962,1.518,1867-07-01T00:00:00Z,false`)

type country struct {
	Id         int64
	Name       string
	Abbrev     string
	Population int64
	GDP        float64
	Founded    time.Time
	Freedom    bool
}

func (c country) String() string {
	return fmt.Sprintf("%s, %s, (%d)", c.Name, c.Abbrev, c.Id)
}

var exampleHolidays = []byte(`Fourth of July,Jul 4
Halloween,Oct 31
Thanksgiving,Nov 27`)

type holiday struct {
	Name string
	Day  time.Time `csv:"Jan _2"`
}

func (h holiday) String() string {
	return fmt.Sprintf("%s (%s)", h.Name, h.Day.Format("01-02"))
}

func expectString(t *testing.T, a, b string) {
	if a != b {
		t.Errorf("Unexpected string: %s != %s", a, b)
	}
}

func expectInt64(t *testing.T, a, b int64) {
	if a != b {
		t.Errorf("Unexpected integer: %d != %d", a, b)
	}
}

func expectDate(t *testing.T, a, b time.Time) {
	if a != b {
		t.Errorf("Unexpected date: %s != %s", a, b)
	}
}

func TestReader_setLayouts(t *testing.T) {
	// Create a buffer with CSV format and a new csv2 reader
	r := NewReader(bytes.NewBuffer(exampleHolidays))

	var h holiday

	// Set the layouts
	typ := reflect.TypeOf(h)
	if typ.Kind() != reflect.Ptr {
		typ = reflect.PtrTo(typ)
	}
	r.setLayouts(typ.Elem())
	if len(r.layouts) != 1 {
		t.Fatalf("Unexpected length of layouts: %d != 1", len(r.layouts))
	}
	expectString(t, r.layouts[1], "Jan _2")

	// Also try with an array
	var holidays []holiday
	err := r.Unmarshal(&holidays)
	if err != nil {
		t.Fatal(err)
	}
	if len(holidays) != 3 {
		t.Fatalf("Unexpected length of holidays: %d != 2", len(holidays))
	}
	expectString(t, holidays[0].String(), "Fourth of July (07-04)")
}

func TestReader_Unmarshal(t *testing.T) {
	// Create a buffer with CSV format and a new csv2 reader
	r := NewReader(bytes.NewBuffer(example))

	// Get rid of the header
	_, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	// Unmarshal the whole file
	var countries []country
	err = r.Unmarshal(&countries)
	if err != nil {
		t.Fatal(err)
	}

	if len(countries) != 2 {
		t.Fatalf("Unexpected length of countries array: %d != %d", len(countries), 2)
	}

	c := countries[0]
	expectString(t, c.Name, "United States")
	expectString(t, c.Abbrev, "US")
	expectInt64(t, c.Id, 2)

	july4 := time.Date(1776, time.Month(7), 4, 0, 0, 0, 0, time.UTC)
	expectDate(t, c.Founded, july4)
	if !c.Freedom {
		t.Errorf("Unexpected boolean: false != true")
	}

	c = countries[1]
	expectString(t, c.Name, "Canada")
	expectString(t, c.Abbrev, "CA")
	expectInt64(t, c.Id, 3)

	// Pass some bad destinations
	r = NewReader(bytes.NewBuffer(example))

	// Not a pointer
	var cs []country
	if r.Unmarshal(cs) != ErrNotPointer {
		t.Error("Did not receive a non-pointer error during Unmarshal")
	}

	// Not a slice
	var cx country
	if r.Unmarshal(&cx) != ErrNotSlice {
		t.Fatal("Did not receive a non-slice error during Unmarshal")
	}
}

func TestReader_UnmarshalOne(t *testing.T) {
	// Create a buffer with CSV format and a new csv2 reader
	r := NewReader(bytes.NewBuffer(example))

	// Get rid of the header
	_, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	// Unmarshal one row
	var c country
	err = r.UnmarshalOne(&c)
	if err != nil {
		t.Fatal(err)
	}

	expectString(t, c.Name, "United States")
	expectString(t, c.Abbrev, "US")
	expectInt64(t, c.Id, 2)

	// Pass some bad destinations
	r = NewReader(bytes.NewBuffer(example))

	// Not a pointer
	if r.UnmarshalOne(c) != ErrNotPointer {
		t.Error("Did not receive a non-pointer error during UnmarshalOne")
	}

	// Not a struct
	var i int
	if r.UnmarshalOne(&i) != ErrNotStruct {
		t.Fatal("Did not receive a non-struct error during UnmarshalOne	")
	}
}
