package kendo

import (
	"fmt"
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

func (ad AggregateDescriptor) getExpression(isRoot bool) string {
	expression := fmt.Sprintf("$items.aggregates.%s.%s", ad.getKey(), ad.Aggregate)
	if isRoot {
		expression = fmt.Sprintf("$items.%s", ad.Field)
	}

	return expression
}

func (ad AggregateDescriptor) getAggregate() string {
	key := ad.Aggregate
	if key == "average" {
		key = "avg"
	}

	return fmt.Sprintf("$%s", key)
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

func (gd GroupDescriptor) getSort() int {
	sort := 1
	if gd.Dir == "desc" {
		sort = -1
	}

	return sort
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

type LookupDescriptor struct { // TODO add toMongo method
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
