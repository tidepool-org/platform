package mongo

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/page"
)

type ListResult[T any] struct {
	Data  []T `json:"data" bson:"data"`
	Count int `json:"count" bson:"count"`
}

// ListResultQueryPipeline returns an aggregation pipeline which will match, sort and aggregate the results
// so the final result can be unmarshalled into ListResult.
func ListResultQueryPipeline(selector bson.M, sort bson.M, pagination page.Pagination) []bson.M {
	pipeline := []bson.M{
		{"$match": selector},
	}
	if sort != nil {
		pipeline = append(pipeline, bson.M{"$sort": sort})
	}
	pipeline = append(pipeline, PaginationFacetPipelineStages(pagination)...)
	return pipeline
}

func PaginationFacetPipelineStages(pagination page.Pagination) []bson.M {
	return []bson.M{
		{
			"$facet": bson.M{
				"data": []bson.M{
					{"$match": bson.M{}},
					{"$skip": pagination.Page * pagination.Size},
					{"$limit": pagination.Size},
				},
				"meta": []bson.M{
					{"$count": "count"},
				},
			},
		},
		// The facet above returns the count in an object as first element of the array, e.g.:
		// {
		//   "data": [...],
		//   "meta": [{"count": 1}]
		// }
		// The projections below lifts it up to the top level, e.g.:
		// {
		//   "data": [],
		//   "count": 1,
		// }
		{
			"$project": bson.M{
				"data": "$data",
				"count": bson.M{
					"$arrayElemAt": bson.A{"$meta.count", 0},
				},
			},
		},
	}
}
