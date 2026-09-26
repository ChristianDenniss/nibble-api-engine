package entity

type Brand struct {
	ID   string
	Slug string
	Name string
}

type Alias struct {
	ID      string
	BrandID string
	Alias   string
}
