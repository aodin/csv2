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

It can also write slices of structs to an output, including a header row derived from struct field names!

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

The previous code will output:

    ID,Name,Abbrev
    1,United States,US
    2,Canada,CA

aodin, 2014
