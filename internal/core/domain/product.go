package domain

// Product represents the core business entity.
type Product struct {
	Name            string
	Price           int
	PriceDiscounted int
	Description     string
	ImagesURL       []string
	Tags            []string
	Status          string
}
