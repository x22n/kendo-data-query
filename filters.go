package kendo

import (
	"fmt"
	"regexp"

	"github.com/globalsign/mgo/bson"
)

//TODO use list of operator
func (f *FilterDescriptor) Filter(filter bson.M) {
	operator := f.Operator
	value := f.Value
	field := f.Field

	var escapedValue string
	if v, ok := value.(string); ok {
		escapedValue = regexp.QuoteMeta(v)
	}

	switch operator {
	case "eq":
		filter[field] = value
	case "ne":
		filter[field] = bson.M{
			"$ne": value,
		}
	case "isnull":
		filter[field] = nil
	case "isnotnull":
		filter[field] = bson.M{
			"$ne": nil,
		}
	case "lt", "lte", "gt", "gte":
		filter[field] = bson.M{
			("$" + operator): value,
		}
	case "startswith":
		filter[field] = bson.M{
			"$regex": fmt.Sprintf("^%s", escapedValue), "$options": "i",
		}
	case "endswith":
		filter[field] = bson.M{
			"$regex": fmt.Sprintf("%s$", escapedValue), "$options": "i",
		}
	case "contains":
		filter[field] = bson.M{
			"$regex": fmt.Sprintf("%s", escapedValue), "$options": "i",
		}
	case "doesnotcontain":
		//TODO case insensitive
		filter[field] = bson.M{
			"$not": fmt.Sprintf("%s", value),
		}
	case "isempty":
		filter[field] = ""
	case "isnotempty":
		filter[field] = bson.M{
			"$ne": "",
		}
	}
}
