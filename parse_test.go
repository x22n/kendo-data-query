package kendo

import (
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestNewDataStateFromRequest2(t *testing.T) {
	t.Run("Should return DataState with values filled", func(t *testing.T) {
		u, _ := url.Parse("https://test.test?filter=title~contains~%27aaaa%27&group=title-asc~owner.firstName-asc&page=1&sort=owner.firstName-asc&pageSize=10")

		request := new(http.Request)
		request.URL = u

		wantValues, _ := url.ParseQuery(u.RawQuery)

		gotDataState, _ := NewDataStateFromRequest(request)
		gotValues := gotDataState.values
		if !reflect.DeepEqual(gotValues, wantValues) {
			t.Errorf("NewDataStateFromRequest() = %v, want %v", gotValues, wantValues)
		}
	})

	t.Run("Should return error if request URL RawQuery is not parseable", func(t *testing.T) {
		u, _ := url.Parse("")
		u.RawQuery = "%" // invalid

		request := new(http.Request)
		request.URL = u

		_, err := NewDataStateFromRequest(request)
		if err == nil {
			t.Errorf("NewDataStateFromRequest() error = %v, wantErr", err)
			return
		}
	})
}

func TestDataState_WithReplacements1(t *testing.T) {
	t.Run("Should set DataState replacements field", func(t *testing.T) {
		replacements := map[string]string{
			"_id": "id",
		}
		d := DataState{}
		d.WithReplacements(replacements)

		if !reflect.DeepEqual(d.replacements, replacements) {
			t.Errorf("DataState.WithReplacements() = %v, want %v", d.replacements, replacements)
		}
	})
}

func TestDataState_WithLookups(t *testing.T) {
	t.Run("Should set DataState Lookup field", func(t *testing.T) {
		lookups := []LookupDescriptor{
			{
				From:         "vendors",
				LocalField:   "vendorId",
				ForeignField: "_id",
				As:           "vendor",
				Single:       true,
			},
		}
		d := DataState{}
		d.WithLookups(lookups)

		if !reflect.DeepEqual(d.Lookup, lookups) {
			t.Errorf("DataState.WithLookups() = %v, want %v", d.Lookup, lookups)
		}
	})
}

func TestDataState_WithPreprocessing(t *testing.T) {
	t.Run("Should set DataState preprocessing field", func(t *testing.T) {
		preprocessing := []bson.M{
			{
				"$addFields": bson.M{
					"cats": 20,
				},
			},
		}
		d := DataState{}
		d.WithPreprocessing(preprocessing)

		if !reflect.DeepEqual(d.preprocessing, preprocessing) {
			t.Errorf("DataState.WithPreprocessing() = %v, want %v", d.preprocessing, preprocessing)
		}
	})
}

func TestDataState_parse(t *testing.T) {
	t.Run("parsePage", func(t *testing.T) {
		t.Run("Should parse page in DataState values and set Page field", func(t *testing.T) {
			wantPage := 5
			v := url.Values{}
			v.Set("page", strconv.Itoa(wantPage))
			d := DataState{}
			d.values = v

			d.parse()

			if !reflect.DeepEqual(d.Page, wantPage) {
				t.Errorf("DataState.parse() = %v, want %v", d.Page, wantPage)
			}
		})

		t.Run("Should return err if page value cannot be parsed as int", func(t *testing.T) {
			v := url.Values{}
			v.Set("page", "one")
			d := DataState{}
			d.values = v

			err := d.parse()

			if _, ok := err.(*strconv.NumError); !ok {
				t.Errorf("DataState.parse() error = %v, want NumError", err)
				return
			}
		})
	})

	t.Run("parsePageSize", func(t *testing.T) {
		t.Run("Should parse page size in DataState values and set PageSize field", func(t *testing.T) {
			wantPageSize := 5
			v := url.Values{}
			v.Set("pageSize", strconv.Itoa(wantPageSize))
			d := DataState{}
			d.values = v

			d.parse()

			if !reflect.DeepEqual(d.PageSize, wantPageSize) {
				t.Errorf("DataState.parse() = %v, want %v", d.PageSize, wantPageSize)
			}
		})

		t.Run("Should return err if page size value cannot be parsed as int", func(t *testing.T) {
			v := url.Values{}
			v.Set("pageSize", "one")
			d := DataState{}
			d.values = v

			err := d.parse()

			if _, ok := err.(*strconv.NumError); !ok {
				t.Errorf("DataState.parse() error = %v, want NumError", err)
				return
			}
		})
	})

	t.Run("parseSortDescriptors", func(t *testing.T) {
		t.Run("Should parse sort in DataState values and set Sort field", func(t *testing.T) {
			v := url.Values{}
			v.Set("sort", "firstName-asc")
			d := DataState{}
			d.values = v

			d.parse()

			wantSort := []SortDescriptor{
				{
					Field: "firstName",
					Dir:   "asc",
				},
			}
			if !reflect.DeepEqual(d.Sort, wantSort) {
				t.Errorf("DataState.parse() = %v, want %v", d.Sort, wantSort)
			}
		})

		t.Run("Should replace sort field if the field has a replacement", func(t *testing.T) {
			v := url.Values{}
			v.Set("sort", "firstName-asc")
			d := DataState{}
			d.values = v
			d.replacements = map[string]string{
				"firstName": "fName",
			}

			d.parse()

			wantSort := []SortDescriptor{
				{
					Field: d.replacements["firstName"],
					Dir:   "asc",
				},
			}
			if !reflect.DeepEqual(d.Sort, wantSort) {
				t.Errorf("DataState.parse() = %v, want %v", d.Sort, wantSort)
			}
		})
	})

	t.Run("parseAggregateDescriptors", func(t *testing.T) {
		t.Run("Should parse aggregate in DataState values and set Aggregates field", func(t *testing.T) {
			v := url.Values{}
			v.Set("aggregate", "due-sum~total-sum")
			d := DataState{}
			d.values = v

			d.parse()

			wantAggregates := []AggregateDescriptor{
				{
					Field:     "due",
					Aggregate: "sum",
				},
				{
					Field:     "total",
					Aggregate: "sum",
				},
			}
			if !reflect.DeepEqual(d.Aggregates, wantAggregates) {
				t.Errorf("DataState.parse() = %v, want %v", d.Aggregates, wantAggregates)
			}
		})

		t.Run("Should replace sort field if the field has a replacement", func(t *testing.T) {
			v := url.Values{}
			v.Set("aggregate", "due-sum~total-sum")
			d := DataState{}
			d.values = v
			d.replacements = map[string]string{
				"due": "unpaid",
			}

			d.parse()

			wantAggregates := []AggregateDescriptor{
				{
					Field:     d.replacements["due"],
					Aggregate: "sum",
				},
				{
					Field:     "total",
					Aggregate: "sum",
				},
			}
			if !reflect.DeepEqual(d.Aggregates, wantAggregates) {
				t.Errorf("DataState.parse() = %v, want %v", d.Aggregates, wantAggregates)
			}
		})
	})

	t.Run("parseGroupDescriptors", func(t *testing.T) {
		t.Run("Should parse group in DataState values and set Group field", func(t *testing.T) {
			v := url.Values{}
			v.Set("group", "title-asc~due-desc")
			d := DataState{}
			d.values = v

			d.parse()

			wantGroups := []GroupDescriptor{
				{
					Field: "title",
					Dir:   "asc",
				},
				{
					Field: "due",
					Dir:   "desc",
				},
			}
			if !reflect.DeepEqual(d.Group, wantGroups) {
				t.Errorf("DataState.parse() = %v, want %v", d.Group, wantGroups)
			}
		})

		t.Run("Should replace sort field if the field has a replacement", func(t *testing.T) {
			v := url.Values{}
			v.Set("group", "title-asc~due-desc")
			d := DataState{}
			d.values = v
			d.replacements = map[string]string{
				"due": "unpaid",
			}

			d.parse()

			wantGroups := []GroupDescriptor{
				{
					Field: "title",
					Dir:   "asc",
				},
				{
					Field: d.replacements["due"],
					Dir:   "desc",
				},
			}
			if !reflect.DeepEqual(d.Group, wantGroups) {
				t.Errorf("DataState.parse() = %v, want %v", d.Group, wantGroups)
			}
		})
	})

	t.Run("parseFilterDescriptors", func(t *testing.T) {
		t.Run("Should parse filter in DataState values and set Filter field", func(t *testing.T) {
			v := url.Values{}
			v.Set("filter", "(title~contains~'hello'~and~firstName~eq~'world')")
			d := DataState{}
			d.values = v

			d.parse()

			wantFilters := []FilterDescriptor{
				{
					Field:    "title",
					Operator: "contains",
					Value:    "hello",
				},
				{
					Field:    "firstName",
					Operator: "eq",
					Value:    "world",
				},
			}
			if !reflect.DeepEqual(d.Filter.Filters, wantFilters) {
				t.Errorf("DataState.parse() = %v, want %v", d.Filter.Filters, wantFilters)
			}
		})
	})
}
