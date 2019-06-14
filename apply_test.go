package kendo

import (
	"reflect"
	"testing"

	"github.com/globalsign/mgo/bson"
)

func TestDataState(t *testing.T) {

	t.Run("GetPipeline", func(t *testing.T) {
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

		t.Run("Should return the correct pipeline", func(t *testing.T) {
			wantPipeline := []bson.M{
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

			if gotPipeline := ds.GetPipeline(); !reflect.DeepEqual(gotPipeline, wantPipeline) {
				t.Errorf("DataState.GetPipeline() = %v, want %v", gotPipeline, wantPipeline)
			}
		})
	})
}
