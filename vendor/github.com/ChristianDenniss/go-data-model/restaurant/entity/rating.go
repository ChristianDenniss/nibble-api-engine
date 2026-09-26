package entity

// Rating is an observed aggregate for a restaurant. The zero value means unknown.
type Rating struct {
	Average float64
	Count   int
}
