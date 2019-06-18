package kendo

import (
	"github.com/globalsign/mgo/bson"
)

func copyM(m bson.M) (copy bson.M) {
	copy = bson.M{}
	for k, v := range m {
		copy[k] = v
	}

	return
}