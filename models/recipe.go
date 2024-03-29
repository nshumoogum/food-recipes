package models

import (
	"strconv"
	"strings"

	errs "github.com/nshumoogum/food-recipes/apierrors"
	"github.com/nshumoogum/food-recipes/helpers"
)

// Recipes contains a list of recipes
type Recipes struct {
	Count      int      `json:"count"`
	Items      []Recipe `json:"items"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	TotalCount int64    `json:"total_count"`
}

// Recipe contains information of a recipe
type Recipe struct {
	ID          string       `bson:"_id,omitempty"               json:"id"`
	CookTime    int          `bson:"cook_time"                   json:"cook_time"`
	Difficulty  string       `bson:"difficulty"                  json:"difficulty"`
	Extras      []Ingredient `bson:"extra_ingredients,omitempty" json:"extra_ingredients,omitempty"`
	Favourite   bool         `bson:"favourite"                   json:"favourite"`
	Ingredients []Ingredient `bson:"ingredients"                 json:"ingredients"`
	Location    Location     `bson:"location"                    json:"location"`
	Notes       string       `bson:"notes,omitempty"             json:"notes,omitempty"`
	PortionSize int          `bson:"portion_size"                json:"portion_size"`
	Tags        []string     `bson:"tags,omitempty"              json:"tags,omitempty"`
	Title       string       `bson:"title"                       json:"title"`
}

// UpdateRecipe TODO probably needs to be removed and logic using this updated to use Patch
type UpdateRecipe struct {
	CookTime    int          `bson:"cook_time"                   json:"cook_time"`
	Difficulty  string       `bson:"difficulty"                  json:"difficulty"`
	Extras      []Ingredient `bson:"extra_ingredients,omitempty" json:"extra_ingredients,omitempty"`
	Favourite   bool         `bson:"favourite"                   json:"favourite"`
	Ingredients []Ingredient `bson:"ingredients"                 json:"ingredients"`
	Location    Location     `bson:"location"                    json:"location"`
	Notes       string       `bson:"notes,omitempty"             json:"notes,omitempty"`
	PortionSize int          `bson:"portion_size"                json:"portion_size"`
	Tags        []string     `bson:"tags,omitempty"              json:"tags,omitempty"`
	Title       string       `bson:"title"                       json:"title"`
}

// Location contains location information for recipe
type Location struct {
	CookBook string `bson:"cook_book,omitempty" json:"cook_book,omitempty"`
	Link     string `bson:"link,omitempty"      json:"link,omitempty"`
	Page     int    `bson:"page,omitempty"      json:"page,omitempty"`
}

// Ingredient contains the ingredient amount
type Ingredient struct {
	Item     string `bson:"item"           json:"item"`
	Quantity int    `bson:"quantity"       json:"quantity"`
	Unit     string `bson:"unit,omitempty" json:"unit,omitempty"`
}

var difficulty = map[string]bool{
	"easy":     true,
	"moderate": true,
	"hard":     true,
}

var validUnits = map[string]bool{
	"ml":   true,
	"l":    true,
	"g":    true,
	"kg":   true,
	"lbs":  true,
	"cups": true,
	"tbsp": true,
	"tsp":  true,
}

// Validate recipe creation
func (recipe *Recipe) Validate() []*ErrorObject {
	return validate(recipe, false)
}

// Validate recipe creation
func (updateRecipe *UpdateRecipe) Validate() []*ErrorObject {
	recipe := &Recipe{
		CookTime:    updateRecipe.CookTime,
		Difficulty:  updateRecipe.Difficulty,
		Extras:      updateRecipe.Extras,
		Favourite:   updateRecipe.Favourite,
		Ingredients: updateRecipe.Ingredients,
		Location:    updateRecipe.Location,
		Notes:       updateRecipe.Notes,
		PortionSize: updateRecipe.PortionSize,
		Tags:        updateRecipe.Tags,
		Title:       updateRecipe.Title,
	}

	return validate(recipe, true)
}

func validate(recipe *Recipe, isUpdate bool) []*ErrorObject {
	var (
		errorObjects  []*ErrorObject
		missingFields []string
		invalidUnits  = make(map[string]string)
	)

	if recipe.CookTime == 0 {
		missingFields = append(missingFields, "cook_time")
	}

	lcDiff := strings.ToLower(recipe.Difficulty)
	if !difficulty[lcDiff] {
		invalidDifficulty := map[string]string{"difficulty": recipe.Difficulty}
		errorObjects = append(errorObjects, &ErrorObject{Error: errs.ErrMissingFields.Error(), ErrorValues: invalidDifficulty})
	}

	// use lower case difficulty value
	recipe.Difficulty = lcDiff

	missingFields = append(missingFields, validateIngredients("extra_ingredients", recipe.Extras, invalidUnits)...)

	if len(recipe.Ingredients) == 0 {
		missingFields = append(missingFields, "ingredients")
	}

	missingFields = append(missingFields, validateIngredients("ingredients", recipe.Ingredients, invalidUnits)...)

	if errorObject := validateLocation(recipe.Location); errorObject != nil {
		errorObjects = append(errorObjects, errorObject)
	}

	if recipe.PortionSize == 0 {
		missingFields = append(missingFields, "portion_size")
	}

	if recipe.PortionSize < 0 {
		errorObjects = append(errorObjects, &ErrorObject{Error: errs.ErrInvalidPortionSize.Error(), ErrorValues: map[string]string{"portion_size": strconv.Itoa(recipe.PortionSize)}})
	}

	if !isUpdate && recipe.Title == "" {
		missingFields = append(missingFields, "title")
	} else if isUpdate && recipe.Title != "" {
		errorObjects = append(errorObjects, &ErrorObject{Error: errs.ErrUnableToChangeTitle.Error(), ErrorValues: map[string]string{"title": recipe.Title}})
	}

	if len(missingFields) > 0 {
		missingFieldList := map[string]string{"fields": helpers.StringifyWords(missingFields)}
		errorObjects = append(errorObjects, &ErrorObject{Error: errs.ErrMissingFields.Error(), ErrorValues: missingFieldList})
	}

	if len(invalidUnits) > 0 {
		errorObjects = append(errorObjects, &ErrorObject{Error: errs.ErrInvalidUnits.Error(), ErrorValues: invalidUnits})
	}

	if errorObjects != nil {
		return errorObjects
	}

	return nil
}

func validateIngredients(fieldName string, ingredients []Ingredient, invalidUnits map[string]string) (missingFields []string) {
	for i, ingredient := range ingredients {
		if ingredient.Item == "" {
			missingFields = append(missingFields, fieldName+".["+strconv.Itoa(i)+"].item")
		}

		if ingredient.Quantity == 0 {
			missingFields = append(missingFields, fieldName+".["+strconv.Itoa(i)+"].quantity")
		}

		if ingredient.Unit != "" {
			if !validUnits[ingredient.Unit] {
				invalidUnits[fieldName+".["+strconv.Itoa(i)+"].unit"] = ingredient.Item
			}
		}
	}

	return
}

func validateLocation(location Location) (err *ErrorObject) {
	var isLink bool

	if location.Link != "" {
		isLink = true
	}

	errorValues := map[string]string{
		"location.cook_book": location.CookBook,
		"location.link":      location.Link,
		"location.page":      strconv.Itoa(location.Page),
	}

	if location.CookBook != "" && location.Page > 0 {
		// have cookbook details
		if isLink {
			// cant have both link and cookbook
			err = &ErrorObject{Error: "cannot contain both cook book details and link", ErrorValues: errorValues}
		}
	} else if location.CookBook == "" && location.Page == 0 {
		// do not have cookbook details
		if !isLink {
			// cant have neither link or cookbook
			err = &ErrorObject{Error: "missing link or cook book details", ErrorValues: map[string]string{"location": "{}"}}
		}
	} else {
		// invalid cookbook details
		if isLink {
			err = &ErrorObject{Error: "invalid cookbook details and competing link", ErrorValues: errorValues}
		} else {
			delete(errorValues, "location.link")
			err = &ErrorObject{Error: "invalid cookbook details", ErrorValues: errorValues}
		}
	}

	return err
}
