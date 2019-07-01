package kendo

import (
	"reflect"
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestDataState(t *testing.T) {

	t.Run("getPipeline", func(t *testing.T) {
		t.Run("Should return the correct complex pipeline", func(t *testing.T) {
			ds := DataState{
				Lookup: []LookupDescriptor{
					{
						From:         "vendors",
						LocalField:   "vendorId",
						ForeignField: "_id",
						As:           "vendor",
						Single:       true,
					},
					{
						From:         "resellers",
						LocalField:   "vendor.resellerId",
						ForeignField: "_id",
						As:           "reseller",
						Single:       true,
					},
				},
				Aggregates: []AggregateDescriptor{
					{
						Aggregate: "average",
						Field:     "commission.due",
					},
					{
						Aggregate: "sum",
						Field:     "commission.due",
					},
				},
				Filter: CompositeFilterDescriptor{
					Logic: "and",
					Filters: []FilterDescriptor{
						{
							Field:    "data.email",
							Operator: "contains",
							Value:    "a",
						},
					},
				},
				Group: []GroupDescriptor{
					{
						Field: "data.email",
						Dir:   "asc",
					},
					{
						Field: "vendor.email",
						Dir:   "desc",
					},
				},
			}

			wantPipeline := []bson.M{
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
				{
					"$lookup": bson.M{
						"as":           "vendor",
						"from":         "vendors",
						"localField":   "vendorId",
						"foreignField": "_id",
					},
				},
				{
					"$addFields": bson.M{
						"vendor": bson.M{
							"$ifNull": []interface{}{
								bson.M{"$arrayElemAt": []interface{}{"$vendor", 0}},
								nil,
							},
						},
					},
				},
				{
					"$lookup": bson.M{
						"from":         "resellers",
						"localField":   "vendor.resellerId",
						"as":           "reseller",
						"foreignField": "_id",
					},
				},
				{
					"$addFields": bson.M{
						"reseller": bson.M{
							"$ifNull": []interface{}{
								bson.M{"$arrayElemAt": []interface{}{"$reseller", 0}},
								nil,
							},
						},
					},
				},
				{
					"$match": bson.M{
						"data.email": bson.M{
							"$regex":   "a",
							"$options": "i",
						},
					},
				},
				{
					"$group": bson.M{
						"_id": bson.M{
							"dataemail":   "$data.email",
							"vendoremail": "$vendor.email",
						},
						"items": bson.M{
							"$push": "$$ROOT",
						},
					},
				},
				{
					"$sort": bson.M{
						"_id.vendoremail": -1,
					},
				},
				{
					"$group": bson.M{
						"_id": "$_id.dataemail",
						"items": bson.M{
							"$push": bson.M{
								"value": "$_id.vendoremail",
								"items": "$items",
								"field": "vendor.email",
								"aggregates": bson.M{
									"commissiondue": bson.M{
										"average": bson.M{
											"$avg": "$items.commission.due",
										},
										"sum": bson.M{
											"$sum": "$items.commission.due",
										},
									},
								},
							},
						},
					},
				},
				{
					"$sort": bson.M{
						"_id": 1,
					},
				},
				{
					"$project": bson.M{
						"value": "$_id",
						"items": "$items",
						"field": "data.email",
						"aggregates": bson.M{
							"commissiondue": bson.M{
								"average": bson.M{
									"$avg": "$items.aggregates.commissiondue.average",
								},
								"sum": bson.M{
									"$sum": "$items.aggregates.commissiondue.sum",
								},
							},
						},
						"_id": 0,
					},
				},
			}

			if gotPipeline := ds.getPipeline(); !reflect.DeepEqual(gotPipeline, wantPipeline) {
				t.Errorf("DataState.getPipeline() = %v, want %v", gotPipeline, wantPipeline)
			}
		})

		t.Run("Should return ascending sort", func(t *testing.T) {
			ds := DataState{
				Sort: []SortDescriptor{
					{
						Dir:   "asc",
						Field: "name",
					},
				},
			}
			wantPipeline := []bson.M{
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
				{
					"$sort": bson.M{
						"name": 1,
					},
				},
			}

			if gotPipeline := ds.getPipeline(); !reflect.DeepEqual(gotPipeline, wantPipeline) {
				t.Errorf("DataState.getPipeline() = %v, want %v", gotPipeline, wantPipeline)
			}
		})

		t.Run("Should return descending sort", func(t *testing.T) {
			ds := DataState{
				Sort: []SortDescriptor{
					{
						Dir:   "desc",
						Field: "name",
					},
				},
			}
			wantPipeline := []bson.M{
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
				{
					"$sort": bson.M{
						"name": -1,
					},
				},
			}

			if gotPipeline := ds.getPipeline(); !reflect.DeepEqual(gotPipeline, wantPipeline) {
				t.Errorf("DataState.getPipeline() = %v, want %v", gotPipeline, wantPipeline)
			}
		})
	})

	t.Run("getSortFields", func(t *testing.T) {
		t.Run("Should return ascending sort", func(t *testing.T) {
			ds := DataState{
				Sort: []SortDescriptor{
					{
						Dir:   "asc",
						Field: "name",
					},
				},
			}

			wantSortFields := bson.M{
				"$sort": bson.M{
					"name": 1,
				},
			}

			if gotSortFields := ds.getSortFields(); !reflect.DeepEqual(gotSortFields, wantSortFields) {
				t.Errorf("DataState.getSortFields() = %v, want %v", gotSortFields, wantSortFields)
			}
		})

		t.Run("Should return descending sort", func(t *testing.T) {
			ds := DataState{
				Sort: []SortDescriptor{
					{
						Dir:   "desc",
						Field: "name",
					},
				},
			}

			wantSortFields := bson.M{
				"$sort": bson.M{
					"name": -1,
				},
			}

			if gotSortFields := ds.getSortFields(); !reflect.DeepEqual(gotSortFields, wantSortFields) {
				t.Errorf("DataState.getSortFields() = %v, want %v", gotSortFields, wantSortFields)
			}
		})
	})

	t.Run("getPaging", func(t *testing.T) {
		t.Run("Should return skip and limit equal to requested page", func(t *testing.T) {
			ds := DataState{
				Page:     3,
				PageSize: 10,
			}

			wantPaging := []bson.M{
				{
					"$skip": 20,
				},
				{
					"$limit": 10,
				},
			}

			if gotPaging := ds.getPaging(); !reflect.DeepEqual(gotPaging, wantPaging) {
				t.Errorf("DataState.getPaging() = %v, want %v", gotPaging, wantPaging)
			}
		})
	})

	t.Run("getTotalPipeline", func(t *testing.T) {
		t.Run("Should return the base pipeline and $count if there is no filter or lookups", func(t *testing.T) {
			ds := DataState{}

			wantTotalPipeline := append(ds.getBasePipeline(), bson.M{
				"$count": "total",
			})

			if gotTotalPipeline := ds.getTotalPipeline(); !reflect.DeepEqual(gotTotalPipeline, wantTotalPipeline) {
				t.Errorf("DataState.getTotalPipeline() = %v, want %v", gotTotalPipeline, wantTotalPipeline)
			}
		})

		t.Run("Should return pipeline with filters if present", func(t *testing.T) {
			ds := DataState{
				Filter: CompositeFilterDescriptor{
					Logic: "and",
					Filters: []FilterDescriptor{
						{
							Field:    "name",
							Operator: "eq",
							Value:    "John",
						},
					},
				},
			}

			wantTotalPipeline := append(ds.getBasePipeline(), []bson.M{
				{"$match": ds.getFilter()},
				{"$count": "total"},
			}...)

			if gotTotalPipeline := ds.getTotalPipeline(); !reflect.DeepEqual(gotTotalPipeline, wantTotalPipeline) {
				t.Errorf("DataState.getTotalPipeline() = %v, want %v", gotTotalPipeline, wantTotalPipeline)
			}
		})

		t.Run("Should lookup if a filter is done on the lookup field", func(t *testing.T) {
			ds := DataState{
				Filter: CompositeFilterDescriptor{
					Logic: "and",
					Filters: []FilterDescriptor{
						{
							Field:    "owner.name",
							Operator: "eq",
							Value:    "John",
						},
					},
				},
				Lookup: []LookupDescriptor{
					{
						From:         "users",
						LocalField:   "owner",
						ForeignField: "_id",
						As:           "owner",
					},
				},
			}

			wantTotalPipeline := append(ds.getBasePipeline(), []bson.M{
				{
					"$lookup": bson.M{
						"from":         "users",
						"localField":   "owner",
						"foreignField": "_id",
						"as":           "owner",
					},
				},
				{"$match": ds.getFilter()},
				{"$count": "total"},
			}...)

			if gotTotalPipeline := ds.getTotalPipeline(); !reflect.DeepEqual(gotTotalPipeline, wantTotalPipeline) {
				t.Errorf("DataState.getTotalPipeline() = %v, want %v", gotTotalPipeline, wantTotalPipeline)
			}
		})

		t.Run("Should not lookup if there is no filters on the lookup field", func(t *testing.T) {
			ds := DataState{
				Filter: CompositeFilterDescriptor{
					Logic: "and",
					Filters: []FilterDescriptor{
						{
							Field:    "name",
							Operator: "eq",
							Value:    "John",
						},
					},
				},
				Lookup: []LookupDescriptor{
					{
						From:         "users",
						LocalField:   "owner",
						ForeignField: "_id",
						As:           "owner",
					},
				},
			}

			wantTotalPipeline := append(ds.getBasePipeline(), []bson.M{
				{"$match": ds.getFilter()},
				{"$count": "total"},
			}...)

			if gotTotalPipeline := ds.getTotalPipeline(); !reflect.DeepEqual(gotTotalPipeline, wantTotalPipeline) {
				t.Errorf("DataState.getTotalPipeline() = %v, want %v", gotTotalPipeline, wantTotalPipeline)
			}
		})
	})
}
