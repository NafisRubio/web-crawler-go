package shopline

// ProductResponse is the top-level structure for the JSON response.
type ProductResponse struct {
	Data ProductShopLine `json:"data"`
}

// ProductShopLine contains the main product details.
type ProductShopLine struct {
	ID                      string            `json:"_id"`
	TitleTranslations       map[string]string `json:"title_translations"`
	DescriptionTranslations map[string]string `json:"description_translations"`
	Media                   []MediaItem       `json:"media"`
	CategoryIDs             []string          `json:"category_ids"`
	Price                   Price             `json:"price"`
	PriceSale               Price             `json:"price_sale"`
	Variations              []Variation       `json:"variations"`
	VariantOptions          []VariantOption   `json:"variant_options"`
	Quantity                int               `json:"quantity"`
}

// MediaItem represents a media object associated with the product.
type MediaItem struct {
	Images          ImageSet       `json:"images"`
	ID              string         `json:"_id"`
	AltTranslations map[string]any `json:"alt_translations"`
	Blurhash        string         `json:"blurhash"`
}

// ImageSet contains different image sizes.
type ImageSet struct {
	Original ImageDetails `json:"original"`
}

// ImageDetails contains the specific details of an image.
type ImageDetails struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	URL    string  `json:"url"`
}

// Price represents a price object with currency details.
type Price struct {
	Cents          int     `json:"cents"`
	CurrencySymbol string  `json:"currency_symbol"`
	CurrencyISO    string  `json:"currency_iso"`
	Label          string  `json:"label"`
	Dollars        float64 `json:"dollars"`
}

// Variation represents a product variation.
type Variation struct {
	Price                  Price               `json:"price"`
	LocationID             any                 `json:"location_id"`
	SKU                    any                 `json:"sku"`
	FieldsTranslations     map[string][]string `json:"fields_translations"`
	Key                    string              `json:"key"`
	MediaID                any                 `json:"media_id"`
	StockIDs               []string            `json:"stock_ids"`
	RootProductVariationID string              `json:"root_product_variation_id"`
	PriceSale              Price               `json:"price_sale"`
	Weight                 float64             `json:"weight"`
	FeedVariations         FeedVariations      `json:"feed_variations"`
	Quantity               int                 `json:"quantity"`
	StockID                string              `json:"stock_id"`
	Warehouse              Warehouse           `json:"warehouse"`
	PreorderLimit          int                 `json:"preorder_limit"`
	MaxOrderQuantity       int                 `json:"max_order_quantity"`
	MemberPrice            Price               `json:"member_price"`
	Fields                 []VariationField    `json:"fields"`
	MPN                    any                 `json:"mpn"`
	GTIN                   any                 `json:"gtin"`
	VariantOptionIDs       []string            `json:"variant_option_ids"`
	ProductPriceTiers      any                 `json:"product_price_tiers"`
}

// FeedVariations contains custom feed variation data.
type FeedVariations struct {
	Custom map[string]string `json:"custom"`
}

// Warehouse represents a warehouse location.
type Warehouse struct {
	ID               string            `json:"_id"`
	NameTranslations map[string]string `json:"name_translations"`
}

// VariationField represents a field within a variation.
type VariationField struct {
	NameTranslations map[string]string `json:"name_translations"`
}

// VariantOption represents an option for a product variant.
type VariantOption struct {
	Key              string            `json:"key"`
	MediaID          any               `json:"media_id"`
	Index            int               `json:"index"`
	NameTranslations map[string]string `json:"name_translations"`
	Type             string            `json:"type"`
}
