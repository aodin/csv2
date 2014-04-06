package csv

import (
	"bytes"
	"fmt"
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

func TestReader_Unmarshal(t *testing.T) {
	// Create a buffer with CSV format
	b := bytes.NewBuffer(example)

	// Create the csv2 reader
	r := NewReader(b)

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
}

func TestReader_UnmarshalOne(t *testing.T) {
	// Create a buffer with CSV format
	b := bytes.NewBuffer(example)

	// Create the csv2 reader
	r := NewReader(b)

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
}
