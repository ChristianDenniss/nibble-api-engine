package entity

type Account struct {
	ID             string
	Name           string
	Email          string
	Phone          string
	Addresses      []SavedAddress
	PaymentMethods []PaymentMethod
}
