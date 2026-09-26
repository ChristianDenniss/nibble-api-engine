package entity

type StoreMatch struct {
	ID            string
	SourceStoreID string
	PlaceID       string
	Confidence    float64
	Status        string
	Method        string
}

type ItemMatch struct {
	ID           string
	SourceItemID string
	DishID       string
	Confidence   float64
	Status       string
	Method       string
}

type Evidence struct {
	ID           string
	StoreMatchID string
	ItemMatchID  string
	SignalKind   string
	Score        float64
	Note         string
}
