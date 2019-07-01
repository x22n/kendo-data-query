# Kendo data query for Go

[![Build Status](https://travis-ci.org/x22n/kendo-data-query.svg)](https://travis-ci.org/x22n/kendo-data-query)
[![codecov](https://codecov.io/gh/x22n/kendo-data-query/branch/master/graph/badge.svg)](https://codecov.io/gh/x22n/kendo-data-query)
[![GoDoc](https://godoc.org/github.com/x22n/kendo-data-query?status.svg)](https://godoc.org/github.com/x22n/kendo-data-query)
[![Go Report Card](https://goreportcard.com/badge/github.com/x22n/kendo-data-query)](https://goreportcard.com/report/github.com/x22n/kendo-data-query)

Go (Golang) library to parse and apply Kendo data query on a MongoDB database using mgo.

---

- [Kendo data query for Go](#kendo-data-query-for-go)
  - [Install](#install)
  - [Examples](#examples)
    - [Handler example](#handler-example)
      - [DataResult example](#dataresult-example)
  - [Limitations](#limitations)
  - [Roadmap](#roadmap)

---

## Install

```sh
go get -u github.com/x22n/kendo-data-query
```

## Examples

### Handler example

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

## Limitations

- Does not support multiple sorts on base columns BUT supports multiple sorted groups
- Only supports `and` logic between filters
- Only supports `avg` and `sum` aggregates

## Roadmap

- Better godoc
- Support for `or` logic between filters
- Support for complex/nested filters
- Support for more aggregates
