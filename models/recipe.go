package models

import (
	"context"
	"strconv"

	"github.com/ONSdigital/log.go/log"
	errs "github.com/nshumoogum/food-recipes/apierrors"
	"github.com/nshumoogum/food-recipes/helpers"
)

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
	ID          string       `json:"id"`
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

var difficulty = map[string]bool{
	"easy":     true,
	"moderate": true,
	"hard":     true,
}

var validUnits = map[string]bool{
	"none": true,
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
	var (
		errorObjects  []*ErrorObject
		missingFields []string
		invalidUnits  = make(map[string]string)
	)
	log.Event(context.Background(), "got error object 1", log.WARN, log.Data{"error": errorObjects})

	if recipe.CookTime == 0 {
		missingFields = append(missingFields, "cook_time")
	}

	if !difficulty[recipe.Difficulty] {
		invalidDifficulty := map[string]string{"difficulty": recipe.Difficulty}
		errorObjects = append(errorObjects, &ErrorObject{Error: errs.ErrMissingFields.Error(), ErrorValues: invalidDifficulty})
	}

	missingFields = append(missingFields, validateIngredients("extra_ingredients", recipe.Extras, invalidUnits)...)

	missingFields = append(missingFields, validateIngredients("ingredients", recipe.Ingredients, invalidUnits)...)
	log.Event(context.Background(), "got error object 2", log.WARN, log.Data{"error": errorObjects})

	if errorObject := validateLocation(recipe.Location); errorObject != nil {
		errorObjects = append(errorObjects, errorObject)
	}

	if recipe.PortionSize == 0 {
		missingFields = append(missingFields, "portion_size")
	}

	if recipe.PortionSize < 0 {
		errorObjects = append(errorObjects, &ErrorObject{Error: errs.ErrInvalidPortionSize.Error(), ErrorValues: map[string]string{"portion_size": strconv.Itoa(recipe.PortionSize)}})
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

		if !validUnits[ingredient.Unit] {
			invalidUnits[fieldName+".["+strconv.Itoa(i)+"].unit"] = ingredient.Item
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

	return
}
