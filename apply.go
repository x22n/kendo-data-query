package kendo

import (
	"fmt"

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

	aggregate := collection.Pipe(d.GetPipeline())

	data := []interface{}{}
	err = aggregate.All(&data)
	fmt.Println(err)

	return DataResult{
		Data:  data,
		Total: total,
	}
}

func (d *DataState) GetPipeline() (pipeline []bson.M) {

	pipeline = []bson.M{}

	if len(d.preprocessing) > 0 {
		pipeline = append(pipeline, d.preprocessing...)
	}

	if len(d.Lookup) > 0 {
		pipeline = append(pipeline, d.getLookup()...)
	}

	//replace _id by id
	pipeline = append(pipeline, bson.M{"$addFields": bson.M{
		"id": "$_id",
	}})
	pipeline = append(pipeline, bson.M{"$project": bson.M{
		"_id": 0,
	}})

	if len(d.Filter.Filters) > 0 {
		pipeline = append(pipeline, bson.M{"$match": d.getFilter()})
	}

	if len(d.Group) > 0 {
		pipeline = append(pipeline, d.getAggregate()...)
		pipeline = append(pipeline, d.getProject())
	}

	if len(d.Sort) > 0 {
		pipeline = append(pipeline, d.getSortFields())
	}

	if d.PageSize > 0 {
		page := d.Page - 1
		pipeline = append(pipeline, []bson.M{
			bson.M{"$skip": page * d.PageSize},
			bson.M{"$limit": d.PageSize},
		}...)
	}

	return
}

func (d *DataState) getTotal(collection mgo.Collection) (total int, err error) {

	lookupsMap := map[string]LookupDescriptor{}
	for _, lookup := range d.Lookup {
		lookupsMap[lookup.As] = lookup
	}

	// apply lookups only if needed by filter
	lookupsToApply := []LookupDescriptor{}
	for _, f := range d.Filter.Filters {
		if lookup, found := lookupsMap[f.Field]; found {
			lookupsToApply = append(lookupsToApply, lookup)
		}
	}

	filter := d.getFilter()
	if len(lookupsToApply) > 0 {
		pipeline := []bson.M{}
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

		data := []interface{}{} // TODO wrong data type
		err = collection.Pipe(pipeline).All(&data)
	} else {
		total, err = collection.Find(filter).Count()
	}

	return
}

func (d *DataState) getLookup() (lookup []bson.M) {

	lookup = []bson.M{}

	for _, l := range d.Lookup {
		lookup = append(lookup, bson.M{
			"$lookup": bson.M{
				"from":         l.From,
				"localField":   l.LocalField,
				"foreignField": l.ForeignField,
				"as":           l.As,
			},
		})
		if l.Single { // should be single doc instead of array
			lookup = append(lookup, bson.M{
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

func (d *DataState) getAggregate() (aggregate []bson.M) {

	aggregate = []bson.M{}

	ids := bson.M{}
	for _, group := range d.Group {
		key := group.getKey()
		ids[key] = fmt.Sprintf("$_id.%s", key)
	}

	nbGroups := len(d.Group) - 1
	for i := nbGroups; i >= 0; i-- {
		group := d.Group[i]
		key := group.getKey()

		sortId := fmt.Sprintf("_id.%s", key)

		if (nbGroups) == i {
			aggregate = append(aggregate, d.getFirstGrouping())
		} else {
			previousGroup := d.Group[i+1]
			previousField := previousGroup.Field
			previousKey := previousGroup.getKey()
			var groupKey interface{}
			if i == 0 {
				sortId = fmt.Sprintf("_id")
				groupKey = fmt.Sprintf("$_id.%s", key)
			} else {
				delete(ids, previousKey)
				groupKey = copyM(ids) //map elements are by reference we have to copy
			}
			aggregate = append(aggregate, d.getGroup(groupKey, previousKey, previousField, i))
		}

		aggregate = append(aggregate, bson.M{
			"$sort": bson.M{
				sortId: getSort(group.Dir),
			},
		})
	}

	return
}

func (d *DataState) getFirstGrouping() (group bson.M) {

	ids := bson.M{}
	for _, group := range d.Group {
		f := group.Field
		key := group.getKey()
		ids[key] = fmt.Sprintf("$%s", f)
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

func (d *DataState) addAggregates(m bson.M, firstlevel bool) bson.M {

	aggregates := bson.M{}

	for _, a := range d.Aggregates {
		key := a.getKey()
		aggregate := bson.M{
			fmt.Sprintf("$%s", d.toMongoAggregate(a.Aggregate)): getAggregateExpression(a, firstlevel),
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

func getAggregateExpression(a AggregateDescriptor, firstlevel bool) (expression string) {

	if firstlevel {
		expression = fmt.Sprintf("$items.%s", a.Field)
	} else {
		expression = fmt.Sprintf("$items.aggregates.%s.%s", a.getKey(), a.Aggregate)
	}

	return
}

func (d *DataState) toMongoAggregate(s string) (a string) {

	switch s {
	case "average":
		a = "avg"
	default:
		a = s
	}

	return
}

func (d *DataState) getProject() (project bson.M) {
	firstGroup := d.Group[0]

	value := "$_id"
	singleGroup := (len(d.Group) == 1)
	if singleGroup {
		value = fmt.Sprintf("$_id.%s", firstGroup.getKey())
	}
	project = bson.M{
		"$project": d.addAggregates(bson.M{
			"_id":   0,
			"value": value,
			"items": "$items",
			"field": firstGroup.Field,
		}, singleGroup),
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
	isSecondGroup := (len(d.Group) - 2) == depth
	group = bson.M{
		"$group": bson.M{
			"_id": id,
			"items": bson.M{
				"$push": d.addAggregates(bson.M{
					"value": fmt.Sprintf("$_id.%s", value),
					"items": "$items",
					"field": field,
				}, isSecondGroup),
			},
		},
	}

	return
}

func copyM(m bson.M) (copy bson.M) {
	copy = bson.M{}
	for k, v := range m {
		copy[k] = v
	}

	return
}

func getSort(s string) (i int) {

	i = 1
	if s == "desc" {
		i = -1
	}

	return
}
