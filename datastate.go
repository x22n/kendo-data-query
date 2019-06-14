package kendo

import (
	"net/url"
	"strings"

	"github.com/globalsign/mgo/bson"
)

type SortDescriptor struct {
	Dir   string // asc desc
	Field string
}

type AggregateDescriptor struct {
	Aggregate string //"count" | "sum" | "average" | "min" | "max"
	Field     string
}

func (ad AggregateDescriptor) getKey() string {
	return sanitizeKey(ad.Field)
}

type GroupDescriptor struct {
	Aggregates []AggregateDescriptor
	Dir        string // asc desc
	Field      string
}

func (gd GroupDescriptor) getKey() string {
	return sanitizeKey(gd.Field)
}


type FilterDescriptor struct {
	Field      string
	IgnoreCase bool
	Operator   string
	Value      interface{}
}

type CompositeFilterDescriptor struct {
	Logic   string             // or and
	Filters []FilterDescriptor //could also be a CompositeFilterDescriptor
}

//Not part of kendo data query api
type LookupDescriptor struct {  // TODO add toMongo method
	From         string
	LocalField   string
	ForeignField string
	As           string
	Single       bool
}

type DataState struct {
	Page          int
	PageSize      int
	Filter        CompositeFilterDescriptor
	Group         []GroupDescriptor
	Sort          []SortDescriptor
	Lookup        []LookupDescriptor
	Aggregates    []AggregateDescriptor
	values        url.Values
	replacements  map[string]string
	preprocessing []bson.M
}


func sanitizeKey(s string) string {
	return strings.Replace(s, ".", "", -1)
}