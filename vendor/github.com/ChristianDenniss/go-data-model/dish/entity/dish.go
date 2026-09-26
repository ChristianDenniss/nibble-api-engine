package entity

type Dish struct {
	ID             string
	BrandID        string
	Name           string
	CanonicalName  string
	Description    string
}

type Alias struct {
	ID     string
	DishID string
	Alias  string
}
