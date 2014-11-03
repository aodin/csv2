package csv

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var example = []byte(`ID,NAME,ABBREV,POPULATION,GDP (trillions),FOUNDED,FREEDOM?
2,"United States","US",317808000,17.438,1776-07-04T00:00:00Z,true
3,"Canada","CA",35344962,1.518,1867-07-01T00:00:00Z,false`)

type country struct {
	ID         int64
	Name       string
	Abbrev     string
	Population int64
	GDP        float64
	Founded    time.Time
	Freedom    bool
}

var typedCountries = []country{
	{2, "United States", "US", 317808000, 17.438, time.Date(1776, 7, 4, 0, 0, 0, 0, time.UTC), true},
	{3, "Canada", "CA", 35344962, 1.518, time.Date(1867, 7, 1, 0, 0, 0, 0, time.UTC), false},
}

// Writer ends with a newline
var expectedCountries = `2,United States,US,317808000,17.438,1776-07-04T00:00:00Z,true
3,Canada,CA,35344962,1.518,1867-07-01T00:00:00Z,false
`

func (c country) String() string {
	return fmt.Sprintf("%s, %s, (%d)", c.Name, c.Abbrev, c.ID)
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

func TestGetFieldNames(t *testing.T) {
	assert := assert.New(t)
	expected := []string{
		"ID",
		"Name",
		"Abbrev",
		"Population",
		"GDP",
		"Founded",
		"Freedom",
	}
	var output []string
	var err error

	// Struct
	output, err = GetFieldNames(country{})
	assert.Nil(err)
	assert.Equal(expected, output)

	// Struct pointer
	output, err = GetFieldNames(&country{})
	assert.Nil(err)
	assert.Equal(expected, output)

	// Slice
	output, err = GetFieldNames([]country{})
	assert.Nil(err)
	assert.Equal(expected, output)

	// pointer to slice
	output, err = GetFieldNames(&[]country{})
	assert.Nil(err)
	assert.Equal(expected, output)

	// slice of pointers
	output, err = GetFieldNames([]*country{})
	assert.Nil(err)
	assert.Equal(expected, output)

	// slice of pointers
	output, err = GetFieldNames(&[]*country{})
	assert.Nil(err)
	assert.Equal(expected, output)
}

func TestSetLayout(t *testing.T) {
	assert := assert.New(t)

	// Create a buffer with CSV format and a new csv2 reader
	r := NewReader(bytes.NewBuffer(exampleHolidays))

	var h holiday

	// Set the layouts
	layout := setLayout(reflect.PtrTo(reflect.TypeOf(h)).Elem())
	assert.Equal(1, len(layout))
	assert.Equal("Jan _2", layout[1])

	// Also try with an array
	var holidays []holiday
	assert.Nil(r.Unmarshal(&holidays))
	assert.Equal(3, len(holidays))
	assert.Equal("Fourth of July (07-04)", holidays[0].String())
}

func TestReader_Unmarshal(t *testing.T) {
	assert := assert.New(t)

	// Create a buffer with CSV format and a new csv2 reader
	r := NewReader(bytes.NewBuffer(example))

	// Get rid of the header
	_, err := r.Read()
	assert.Nil(err)

	// Unmarshal the whole file
	var countries []country
	assert.Nil(r.Unmarshal(&countries))
	assert.Equal(2, len(countries))

	c := countries[0]
	assert.Equal("United States", c.Name)
	assert.Equal("US", c.Abbrev)
	assert.Equal(2, c.ID)
	assert.Equal(
		time.Date(1776, time.Month(7), 4, 0, 0, 0, 0, time.UTC),
		c.Founded,
	)
	assert.Equal(true, c.Freedom)

	c = countries[1]
	assert.Equal("Canada", c.Name)
	assert.Equal("CA", c.Abbrev)
	assert.Equal(3, c.ID)

	// Pass some bad destinations
	r = NewReader(bytes.NewBuffer(example))

	// Not a pointer
	var cs []country
	assert.Equal(ErrNotPointer, r.Unmarshal(cs))

	// Not a slice
	var cx country
	assert.Equal(ErrNotSlice, r.Unmarshal(&cx))
}

func TestReader_UnmarshalOne(t *testing.T) {
	assert := assert.New(t)

	// Create a buffer with CSV format and a new csv2 reader
	r := NewReader(bytes.NewBuffer(example))

	// Get rid of the header
	_, err := r.Read()
	assert.Nil(err)

	// Unmarshal one row
	var c country
	assert.Nil(r.UnmarshalOne(&c))
	assert.Equal("United States", c.Name)
	assert.Equal("US", c.Abbrev)
	assert.Equal(2, c.ID)
	assert.Equal(
		time.Date(1776, time.Month(7), 4, 0, 0, 0, 0, time.UTC),
		c.Founded,
	)

	// Pass some bad destinations
	r = NewReader(bytes.NewBuffer(example))

	// Not a pointer
	assert.Equal(ErrNotPointer, r.UnmarshalOne(c))

	// Not a struct
	var i int
	assert.Equal(ErrNotStruct, r.UnmarshalOne(&i))
}

func TestWriter_Marshal(t *testing.T) {
	assert := assert.New(t)

	// Create a buffer with CSV format and a new csv2 writer
	var b bytes.Buffer
	var w *Writer
	w = NewWriter(&b)

	// Marshal the countries array
	assert.Nil(w.Marshal(&typedCountries))
	assert.Equal(expectedCountries, b.String())
}
