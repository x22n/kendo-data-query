package kendo

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
)

const (
	TimeLayout      = "2006-01-02T15-04-05"
	TokensPerFilter = 4
)

// NewDataStateFromRequest creates a *DataState from the query parameters of the request
func NewDataStateFromRequest(request *http.Request) (dataState *DataState, err error) {

	values, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		return
	}

	dataState = new(DataState)
	dataState.values = values

	return
}

// WithReplacements adds field replacements to the DataState.
// For example map[string]string{ "_id": "id" }
func (d *DataState) WithReplacements(replacements map[string]string) {
	d.replacements = replacements
}

// WithLookups adds LookupDescriptors to the DataState
func (d *DataState) WithLookups(lookups []LookupDescriptor) {
	d.Lookup = lookups
}

// WithPreprocessing adds pipeline steps to the DataState that are executed before any other steps
func (d *DataState) WithPreprocessing(preprocessing []bson.M) {
	d.preprocessing = preprocessing
}

func (d *DataState) parse() (err error) {

	if err = d.parsePage(); err != nil {
		return
	}

	if err = d.parsePageSize(); err != nil {
		return
	}

	if err = d.parseFilterDescriptors(); err != nil {
		return
	}

	d.parseSortDescriptors()
	d.parseGroupDescriptors()
	d.parseAggregateDescriptors()

	return
}

func (d *DataState) parseAggregateDescriptors() {
	aggregrate := d.values.Get("aggregate")
	if aggregrate == "" {
		return
	}

	tokens := tokenize(aggregrate, "~")
	aggregates := make([]AggregateDescriptor, len(tokens))
	for i, t := range tokens {
		values := tokenize(t, "-")
		aggregates[i].Field = d.replaceField(values[0])
		aggregates[i].Aggregate = values[1]
	}
	d.Aggregates = aggregates
}

func (d *DataState) parseFilterDescriptors() (err error) {
	filter := d.values.Get("filter")
	if filter == "" {
		return
	}

	tokens := tokenize(filter, "~")
	nbTokens := len(tokens)
	nbFilters := (nbTokens / TokensPerFilter) + 1
	filters := make([]FilterDescriptor, nbFilters)
	for i := 0; i < nbTokens; i += TokensPerFilter {
		idx := i / TokensPerFilter

		var value interface{}
		value, err = toValue(tokens[i+2])
		if err != nil {
			return
		}

		filters[idx] = FilterDescriptor{
			Field:    d.replaceField(tokens[i]),
			Operator: tokens[i+1],
			Value:    value,
		}
	}

	d.Filter.Filters = filters
	if nbFilters > 1 {
		d.Filter.Logic = "and" //TODO support "or" logic
	}

	return
}

func (d *DataState) parseGroupDescriptors() {
	group := d.values.Get("group")
	if group == "" {
		return
	}

	tokens := tokenize(group, "~")
	groups := make([]GroupDescriptor, len(tokens))
	for i, t := range tokens {
		group := tokenize(t, "-")
		field := d.replaceField(group[0])
		direction := group[1]
		groups[i] = GroupDescriptor{
			Field: field,
			Dir:   direction,
		}
	}

	d.Group = groups
}

func (d *DataState) parsePage() (err error) {
	page := d.values.Get("page")
	if page == "" {
		return
	}

	d.Page, err = strconv.Atoi(page)

	return
}

func (d *DataState) parsePageSize() (err error) {
	size := d.values.Get("pageSize")
	if size == "" {
		return
	}

	d.PageSize, err = strconv.Atoi(size)

	return
}

func (d *DataState) parseSortDescriptors() {
	sort := d.values.Get("sort")
	if sort == "" {
		return
	}

	s := tokenize(sort, "-")
	if len(s) >= 2 { //id-asc
		field := d.replaceField(s[0])
		direction := s[1]
		d.Sort = append(d.Sort, SortDescriptor{
			Field: field,
			Dir:   direction,
		})
	}
}

func (d *DataState) replaceField(field string) (replaced string) {

	replaced = field
	if replacement, ok := d.replacements[field]; ok {
		replaced = replacement
	}

	return
}

func tokenize(s string, sep string) (tokens []string) {

	i := strings.Index(s, "(")
	j := strings.Index(s, ")")

	if i != -1 && j != 1 {
		s = s[i+1 : j]
	}

	return strings.Split(s, sep)
}

func toValue(s string) (value interface{}, err error) {

	s = strings.TrimSuffix(s, "'")
	if d := strings.TrimPrefix(s, "datetime'"); d != s {
		value, err = time.Parse(TimeLayout, d)
	} else if strings.HasPrefix(s, "'") == false {
		value, err = strconv.ParseFloat(s, 64)
	} else {
		value = strings.TrimPrefix(s, "'")
	}

	return
}
