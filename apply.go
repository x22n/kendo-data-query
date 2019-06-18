package kendo

import (
	"fmt"
	"strings"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func (d *DataState) Apply(collection mgo.Collection) (dataResult DataResult) {
	if err := d.parse(); err != nil {
		return
	}

	total, err := d.getTotal(collection)
	if err != nil {
		return
	}

	aggregate := collection.Pipe(d.getPipeline())

	data := []interface{}{}
	err = aggregate.All(&data)
	fmt.Println(err)

	return DataResult{
		Data:  data,
		Total: total,
	}
}

func (d *DataState) getBasePipeline() (pipeline []bson.M) {

	pipeline = []bson.M{
		{
			"$addFields": bson.M{
				"id": "$_id",
			},
		},
		{
			"$project": bson.M{
				"_id": 0,
			},
		},
	}

	if len(d.preprocessing) > 0 {
		pipeline = append(pipeline, d.preprocessing...)
	}

	return
}

func (d *DataState) getPipeline() (pipeline []bson.M) {

	pipeline = d.getBasePipeline()

	if len(d.Lookup) > 0 {
		pipeline = append(pipeline, d.getLookups()...)
	}

	if len(d.Filter.Filters) > 0 {
		pipeline = append(pipeline, bson.M{"$match": d.getFilter()})
	}

	if len(d.Group) > 0 {
		pipeline = append(pipeline, d.getGroups()...)
		pipeline = append(pipeline, d.getProject())
	}

	if len(d.Sort) > 0 {
		pipeline = append(pipeline, d.getSortFields())
	}

	if d.PageSize > 0 {
		pipeline = append(pipeline, d.getPaging()...)
	}

	return
}

func (d *DataState) getTotalPipeline() (pipeline []bson.M) {

	lookupsMap := map[string]LookupDescriptor{}
	for _, lookup := range d.Lookup {
		lookupsMap[lookup.As] = lookup
	}

	// apply lookups only if needed by filter
	lookupsToApply := []LookupDescriptor{}
	for _, f := range d.Filter.Filters {
		rootKey := strings.Split(f.Field, ".")[0]
		if lookup, found := lookupsMap[rootKey]; found {
			lookupsToApply = append(lookupsToApply, lookup)
		}
	}

	pipeline = d.getBasePipeline()

	if len(lookupsToApply) > 0 {
		for _, l := range d.Lookup {
			pipeline = append(pipeline, bson.M{
				"$lookup": bson.M{
					"from":         l.From,
					"localField":   l.LocalField,
					"foreignField": l.ForeignField,
					"as":           l.As,
				},
			})
		}
	}

	if len(d.Filter.Filters) > 0 {
		pipeline = append(pipeline, bson.M{"$match": d.getFilter()})
	}

	pipeline = append(pipeline, bson.M{
		"$count": "total",
	})

	return
}

func (d *DataState) getTotal(collection mgo.Collection) (total int, err error) {

	var data struct {
		Total int `bson:"total"`
	}
	err = collection.Pipe(d.getTotalPipeline()).One(&data)

	return data.Total, err
}

func (d *DataState) getLookups() (lookups []bson.M) {

	lookups = []bson.M{}

	for _, l := range d.Lookup {
		lookups = append(lookups, bson.M{
			"$lookup": bson.M{
				"from":         l.From,
				"localField":   l.LocalField,
				"foreignField": l.ForeignField,
				"as":           l.As,
			},
		})
		if l.Single { // should be single doc instead of array
			lookups = append(lookups, bson.M{
				"$addFields": bson.M{
					l.As: bson.M{
						"$ifNull": []interface{}{
							bson.M{"$arrayElemAt": []interface{}{fmt.Sprintf("$%s", l.As), 0}},
							nil,
						},
					},
				},
			})
		}
	}

	return
}

func (d *DataState) getGroups() (groups []bson.M) {

	groups = []bson.M{}

	ids := bson.M{}
	for _, group := range d.Group {
		key := group.getKey()
		ids[key] = fmt.Sprintf("$_id.%s", key)
	}

	nbGroups := len(d.Group) - 1
	for i := nbGroups; i >= 0; i-- {
		group := d.Group[i]
		key := group.getKey()

		sortKey := fmt.Sprintf("_id.%s", key)

		if (nbGroups) == i {
			groups = append(groups, d.getFirstGrouping())
		} else {
			previousGroup := d.Group[i+1]
			previousField := previousGroup.Field
			previousKey := previousGroup.getKey()
			var groupKey interface{}
			if i == 0 {
				sortKey = "_id"
				groupKey = fmt.Sprintf("$_id.%s", key)
			} else {
				delete(ids, previousKey)
				groupKey = copyM(ids) //map elements are by reference we have to copy
			}
			groups = append(groups, d.getGroup(groupKey, previousKey, previousField, i))
		}

		groups = append(groups, bson.M{
			"$sort": bson.M{
				sortKey: group.getSort(),
			},
		})
	}

	return
}

func (d *DataState) getFirstGrouping() (group bson.M) {

	ids := bson.M{}
	for _, group := range d.Group {
		key := group.getKey()
		ids[key] = fmt.Sprintf("$%s", group.Field)
	}

	group = bson.M{
		"$group": bson.M{
			"_id": ids,
			"items": bson.M{
				"$push": "$$ROOT",
			},
		},
	}

	return
}

func (d *DataState) addAggregates(m bson.M, isLast bool) bson.M {

	aggregates := bson.M{}

	for _, a := range d.Aggregates {
		key := a.getKey()
		aggregate := bson.M{
			a.getAggregate(): a.getExpression(isLast),
		}

		if agg, ok := aggregates[key]; ok {
			m, _ := agg.(bson.M)
			m[a.Aggregate] = aggregate
			aggregates[key] = m
		} else {
			aggregates[key] = bson.M{
				a.Aggregate: aggregate,
			}
		}
	}

	if len(aggregates) == 0 {
		aggregates["_"] = nil //cannot project an empty object
	}

	m["aggregates"] = aggregates

	return m
}

func (d *DataState) getProject() (project bson.M) {
	firstGroup := d.Group[0]

	value := "$_id"
	isLast := (len(d.Group) == 1)
	if isLast {
		value = fmt.Sprintf("$_id.%s", firstGroup.getKey())
	}
	project = bson.M{
		"$project": d.addAggregates(bson.M{
			"_id":   0,
			"value": value,
			"items": "$items",
			"field": firstGroup.Field,
		}, isLast),
	}

	return
}

func (d *DataState) getSortFields() (sort bson.M) {
	var fields bson.M
	for _, s := range d.Sort {
		fields = bson.M{
			s.Field: 1,
		}

		if s.Dir == "desc" {
			fields[s.Field] = -1
		}
	}

	return bson.M{
		"$sort": fields,
	}
}

func (d *DataState) getFilter() (filter bson.M) {
	filter = bson.M{}

	for _, f := range d.Filter.Filters {
		f.Filter(filter)
	}

	return
}

func (d *DataState) getGroup(id interface{}, value string, field string, depth int) (group bson.M) {
	isLast := (len(d.Group) - 2) == depth
	group = bson.M{
		"$group": bson.M{
			"_id": id,
			"items": bson.M{
				"$push": d.addAggregates(bson.M{
					"value": fmt.Sprintf("$_id.%s", value),
					"items": "$items",
					"field": field,
				}, isLast),
			},
		},
	}

	return
}

func (d *DataState) getPaging() (paging []bson.M) {
	page := d.Page - 1

	return []bson.M{
		bson.M{"$skip": page * d.PageSize},
		bson.M{"$limit": d.PageSize},
	}
}
