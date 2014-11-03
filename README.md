csv2
====

Parse `csv` files directly into `struct` instances.

[![Build Status](https://travis-ci.org/aodin/csv2.svg)](https://travis-ci.org/aodin/csv2)


Quickstart
----------

Parse the `csv` file:

    ID,NAME,ABBREV
    1,"United States",US
    2,"Canada",CA

```go
package main

import (
    "fmt"
    "os"

    "github.com/aodin/csv2" // Will import as csv!
)

type Country struct {
    ID     int64
    Name   string
    Abbrev string
}

func (c Country) String() string {
    return fmt.Sprintf("%d: %s", c.ID, c.Name)
}

func main() {
    f, err := os.Open("./countries.csv")
    if err != nil {
        panic(err)
    }
    csvf := csv.NewReader(f)

    // Discard the header
    if _, err = csvf.Read(); err != nil {
        panic(err)
    }

    var countries []Country
    if err = csvf.Unmarshal(&countries); err != nil {
        panic(err)
    }
    fmt.Println(countries)
}
```

Destination Types
-----------------

Supported field types for destination structs include:

* `string`
* `int64`
* `float64`
* `bool`
* `time.Time`

By default, `time.Time` fields will be parsed with the `RFC3339` layout. Alternative layouts can be specified by adding a `csv` tag to the field:

```go
type holiday struct {
    Name string
    Day  time.Time `csv:"Jan _2"`
}
```

When a `csv` column is empty, the types above will error. You will want to use a pointer type, which will be set to `nil` when there is no content:

* `*int64`
* `*float64`
* `*bool`
* `*time.Time`


Writer
------

The package can also write slices of structs to an output, including a header row derived from struct field names!

```go
f, err := os.OpenFile("out.csv", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
if err != nil {
    panic(err)
}
defer f.Close()

writer := csv.NewWriter(f)
writer.WriteHeader(&countries)
writer.Marshal(&countries)
```

    ID,Name,Abbrev
    1,United States,US
    2,Canada,CA


Happy Hacking!

aodin, 2014
