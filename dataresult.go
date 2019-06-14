package kendo

type DataResult struct {
	Data  []interface{} `json:"data"`
	Total int           `json:"total"`
}
