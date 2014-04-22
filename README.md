csv2
====

Adds schema to Go's `encoding/csv` package.

Given a csv file:

    ID,NAME,ABBREV
    1,"United States",US
    2,"Canada",CA

It can be parsed using:

```go
package main

import (
    "fmt"
    "github.com/aodin/csv2" // Will import as csv!
    "os"
)

type Country struct {
    Id     int64
    Name   string
    Abbrev string
}

func (c Country) String() string {
    return fmt.Sprintf("%d: %s", c.Id, c.Name)
}

func main() {
    f, err := os.Open("./countries.csv")
    if err != nil {
        panic(err)
    }
    csvf := csv.NewReader(f)

    // Discard the header
    _, err = csvf.Read()
    if err != nil {
        panic(err)
    }

    var countries []Country
    err = csvf.Unmarshal(&countries)
    if err != nil {
        panic(err)
    }
    fmt.Println(countries)
}
```

By default, `time.Time` fields will be parsed with the `RFC3339` layout. Alternative layouts can be specified by adding a `csv` tag to the field:

```go
type holiday struct {
    Name string
    Day  time.Time `csv:"Jan _2"`
}
```

aodin, 2014
