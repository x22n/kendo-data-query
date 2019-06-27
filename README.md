# Kendo data query for Go

[![Build Status](https://travis-ci.org/XavierTS/kendo-data-query.svg)](https://travis-ci.org/XavierTS/kendo-data-query)
[![codecov](https://codecov.io/gh/XavierTS/kendo-data-query/branch/master/graph/badge.svg)](https://codecov.io/gh/XavierTS/kendo-data-query)
[![GoDoc](https://godoc.org/github.com/XavierTS/kendo-data-query?status.svg)](https://godoc.org/github.com/XavierTS/kendo-data-query)

## Limitations

* Does not support multiple sorts on base columns BUT supports multiple sorted groups
* Only supports `and` logic between filters
* Only supports `avg` and `sum` aggregates

## Usage

#### Handler example
```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    ds, err := kendo.NewDataStateFromRequest(ctx.Request)
    if err != nil {
        // Error handling
    }
    ...
    // the following should not be directly in the handler, for reference only
    session, err := mgo.DialWithInfo(mongoDBDialInfo)
    collection := session.DB("db").C("collection")
    dr := ds.Apply(collection)
}
```

#### DataResult example

```json
{"data":[{"title":"cat","due":1.98},{"title":"dog","due":8.21},...],"total":325}
```

## Roadmap

* Better godoc
* Better path coverage
* Support for `or` logic between filters
* Support for complex/nested filters
* Better error handling on `Apply(collection)`
* Support for more aggregates