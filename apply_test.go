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
					AggregateDescriptor{
						Aggregate: "average",
						Field:     "commission.due",
					},
					AggregateDescriptor{
						Aggregate: "sum",
						Field:     "commission.due",
					},
				},
				Filter: CompositeFilterDescriptor{
					Logic: "and",
					Filters: []FilterDescriptor{
						FilterDescriptor{
							Field:    "data.email",
							Operator: "contains",
							Value:    "a",
						},
					},
				},
				Group: []GroupDescriptor{
					GroupDescriptor{
						Field: "data.email",
						Dir:   "asc",
					},
					GroupDescriptor{
						Field: "vendor.email",
						Dir:   "desc",
					},
				},
			}

			wantPipeline := []bson.M{
				bson.M{
					"$addFields": bson.M{
						"id": "$_id",
					},
				},
				bson.M{
					"$project": bson.M{
						"_id": 0,
					},
				},
				bson.M{
					"$lookup": bson.M{
						"as":           "vendor",
						"from":         "vendors",
						"localField":   "vendorId",
						"foreignField": "_id",
					},
				},
				bson.M{
					"$addFields": bson.M{
						"vendor": bson.M{
							"$ifNull": []interface{}{
								bson.M{"$arrayElemAt": []interface{}{"$vendor", 0}},
								nil,
							},
						},
					},
				},
				bson.M{
					"$lookup": bson.M{
						"from":         "resellers",
						"localField":   "vendor.resellerId",
						"as":           "reseller",
						"foreignField": "_id",
					},
				},
				bson.M{
					"$addFields": bson.M{
						"reseller": bson.M{
							"$ifNull": []interface{}{
								bson.M{"$arrayElemAt": []interface{}{"$reseller", 0}},
								nil,
							},
						},
					},
				},
				bson.M{
					"$match": bson.M{
						"data.email": bson.M{
							"$regex":   "a",
							"$options": "i",
						},
					},
				},
				bson.M{
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
				bson.M{
					"$sort": bson.M{
						"_id.vendoremail": -1,
					},
				},
				bson.M{
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
				bson.M{
					"$sort": bson.M{
						"_id": 1,
					},
				},
				bson.M{
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
				bson.M{
					"$addFields": bson.M{
						"id": "$_id",
					},
				},
				bson.M{
					"$project": bson.M{
						"_id": 0,
					},
				},
				bson.M{
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
				bson.M{
					"$addFields": bson.M{
						"id": "$_id",
					},
				},
				bson.M{
					"$project": bson.M{
						"_id": 0,
					},
				},
				bson.M{
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
}
