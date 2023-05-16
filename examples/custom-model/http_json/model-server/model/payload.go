package model

type Request struct {
	Instances [][]float64 `json:"instances"`
}

type Response struct {
	Predictions []float64 `json:"predictions"`
}
