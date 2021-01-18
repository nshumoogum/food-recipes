package models

// Recipes ...
type Recipes struct {
	Count      int      `json:"count"`
	Items      []Recipe `json:"items"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	TotalCount int      `json:"total_count"`
}

// Recipe ...
type Recipe struct {
	CookTime    int          `json:"cook_time"`
	Difficulty  string       `json:"difficulty"`
	Extras      []Ingredient `json:"extra_ingredients"`
	Favourite   bool         `json:"favourite"`
	Ingredients []Ingredient `json:"ingredients"`
	Location    Location     `json:"location"`
	Notes       string       `json:"notes,omitempty"`
	PortionSize int          `json:"portion_size"`
	Tags        []string     `json:"tags,omitempty"`
	Title       string       `json:"title"`
}

// Location ...
type Location struct {
	CookBook string `json:"cook_book"`
	Link     string `json:"link"`
	Page     int    `json:"page,omitempty"`
}

// Ingredient ...
type Ingredient struct {
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
	Unit     string `json:"unit,omitempty"`
}
