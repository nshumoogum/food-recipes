package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ONSdigital/log.go/log"
	"github.com/nshumoogum/food-recipes/config"
	"github.com/nshumoogum/food-recipes/models"
	"github.com/nshumoogum/food-recipes/service"
	"github.com/pkg/errors"
)

const serviceName = "food-recipes"

var recipeData = make(map[string]models.Recipe)

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Event(ctx, "application unexpectedly failed", log.ERROR, log.Error(err))
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Read config
	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "failed to retrieve configuration", log.FATAL, log.Error(err))
		return err
	}
	log.Event(ctx, "config on startup", log.INFO, log.Data{"config": cfg})

	if cfg.DownloadData {
		if err = Download(ctx, cfg.GSURL, cfg.DownloadTimeout); err != nil {
			return err
		}
	}

	// Create the service, providing an error channel for fatal errors
	svcErrors := make(chan error, 1)

	// Run the service
	svc := service.New(cfg)
	if err := svc.Run(ctx, recipeData, svcErrors); err != nil {
		return errors.Wrap(err, "running service failed")
	}

	// Blocks until an os interrupt or a fatal error occurs
	select {
	case err := <-svcErrors:
		log.Event(ctx, "service error received", log.ERROR, log.Error(err))
	case sig := <-signals:
		log.Event(ctx, "os signal received", log.Data{"signal": sig}, log.INFO)
	}

	return svc.Close(ctx)
}

// Download data on initialisation
func Download(ctx context.Context, url string, timeout time.Duration) error {
	logData := log.Data{"url": url}
	log.Event(ctx, "downloading data", log.INFO, logData)

	if url == "" {
		log.Event(ctx, "missing google sheets url, no data loaded", log.WARN, logData)
		return nil
	}

	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Event(ctx, "cannot download file from the given url", log.ERROR, log.Error(err), logData)
		return err
	}

	if resp.StatusCode != 200 {
		err := errors.New("response from the URL was" + strconv.Itoa(resp.StatusCode) + "but expecting 200")
		log.Event(ctx, "unexpected response code", log.ERROR, log.Error(err), logData)
		return err
	}

	if resp.Header["Content-Type"][0] != "text/csv" {
		err := fmt.Errorf("the file downloaded has content type '%s', expected 'text/csv'", resp.Header["Content-Type"])
		log.Event(ctx, "unexpected response header", log.ERROR, log.Error(err), logData)
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Event(ctx, "unable to read response body", log.ERROR, log.Error(err), logData)
		return err
	}

	// Store data in-memory
	csvReader := csv.NewReader(strings.NewReader(string(b)))

	// Scan header row
	_, err = csvReader.Read()
	if err != nil {
		log.Event(ctx, "encountered error when processing header row of csv", log.ERROR, log.Error(err))
		return err
	}

	count := 0
	for {
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		recipeLogData := log.Data{"line_count": count, "csv_line": line}
		if err != nil {
			log.Event(ctx, "encountered error reading csv", log.ERROR, log.Error(err), recipeLogData)
			break
		}

		extras := getIngredients(ctx, line[11], recipeLogData)
		ingredients := getIngredients(ctx, line[10], recipeLogData)

		location := models.Location{
			CookBook: line[3],
			Link:     line[2],
		}

		location.Page, err = strconv.Atoi(line[4])
		if err != nil {
			recipeLogData["page"] = line[4]
			log.Event(ctx, "page value unreadable", log.WARN, recipeLogData)
		}

		var favourite bool
		if line[6] == "TRUE" {
			favourite = true
		}

		tags := strings.Split(line[5], "/")

		title := strings.ReplaceAll(line[0], " ", "-")
		lcTitle := strings.ToLower(title)

		recipe := models.Recipe{
			ID:          lcTitle,
			Difficulty:  line[8],
			Extras:      extras,
			Favourite:   favourite,
			Ingredients: ingredients,
			Location:    location,
			Notes:       line[9],
			Tags:        tags,
			Title:       line[0],
		}

		recipe.CookTime, err = strconv.Atoi(line[7])
		if err != nil {
			recipeLogData["cook_time"] = line[3]
			log.Event(ctx, "cook_time value unreadable", log.WARN, recipeLogData)
		}

		recipe.PortionSize, err = strconv.Atoi(line[1])
		if err != nil {
			recipeLogData["portion_size"] = line[1]
			log.Event(ctx, "portion_size value unreadable", log.WARN, recipeLogData)
		}

		recipeData[lcTitle] = recipe

		count++
	}

	logData["count"] = count
	log.Event(ctx, "successfuly loaded recipe data", log.INFO, logData)

	return nil
}

func getIngredients(ctx context.Context, cell string, logData log.Data) (ingredientList []models.Ingredient) {
	if cell == "" {
		return
	}

	ingredients := strings.Split(strings.ReplaceAll(cell, ")", ""), "(")
	log.Event(ctx, "what do we have", log.Data{"ingredients": ingredients})

	for _, ingredient := range ingredients {
		if ingredient == "" {
			continue
		}

		logData["ingredient"] = ingredient
		log.Event(ctx, "what do we have", logData)
		ingredientParts := strings.Split(ingredient, ":")

		quantity, err := strconv.Atoi(ingredientParts[1])
		if err != nil {
			logData["quantity"] = ingredientParts[1]
			log.Event(ctx, "quantity value unreadable", log.WARN, logData)
		}

		ingredientList = append(ingredientList, models.Ingredient{
			Item:     ingredientParts[0],
			Quantity: quantity,
			Unit:     ingredientParts[2],
		})
	}

	return
}
