package types

type ValueRow []interface{}

type ValueRows []ValueRow

type FeatureTable struct {
	Name    string    `json:"-"`
	Columns []string  `json:"columns"`
	Data    ValueRows `json:"data"`
}
