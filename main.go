package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// function for accessing the database
func GetMongoDbConnection() (*mongo.Client, error) {

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb+srv://admin:password1A@cyrptocluster.lmp7s.mongodb.net/sample_airbnb?retryWrites=true&w=majority"))

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	return client, nil
}

// function for connection to specific db and collection
func getMongoDbCollection(DbName string, CollectionName string) (*mongo.Collection, error) {

	client, err := GetMongoDbConnection()

	// handle errors
	if err != nil {
		return nil, err
	}

	collection := client.Database(DbName).Collection(CollectionName)

	return collection, nil
}

func main() {
	app := fiber.New()

	// allow for cors
	app.Use(cors.New())

	// get a single house
	app.Get("/house/:id?", func(c *fiber.Ctx) error {
		// connect to the database
		collection, err := getMongoDbCollection("sample_airbnb", "listingsAndReviews")
		if err != nil {
			// error in connection return error
			return c.Status(500).Send([]byte(err.Error()))
		}

		// filter to only get one key
		var filter bson.M = bson.M{}
		if c.Params("id") != "" {
			id := c.Params("id")
			filter = bson.M{"_id": id}
		}

		// make the results be in the correct format
		var results []bson.M

		// actually make the request using the cursor
		cur, err := collection.Find(context.Background(), filter)
		defer cur.Close(context.Background())

		// handle errors
		if err != nil {
			return c.Status(500).Send([]byte(err.Error()))
		}

		// grab all of the results from the quesry
		cur.All(context.Background(), &results)

		// handle errors
		if results == nil {
			return c.Status(404).SendString("not Found")
		}

		// send in json format the results
		json, _ := json.Marshal(results)
		return c.Send(json)
	})

	// main list
	app.Get("/main", func(c *fiber.Ctx) error {
		// connect to the database
		collection, err := getMongoDbCollection("sample_airbnb", "listingsAndReviews")
		if err != nil {
			// error in connection return error
			return c.Status(500).Send([]byte(err.Error()))

		}

		var results []bson.M

		cursor, err := collection.Find(context.Background(), bson.M{"$and": []bson.M{{"price": bson.M{"$gt": 80}}, {"price": bson.M{"$lt": 100}}, {"price": bson.M{"$exists": true}}}})

		if err != nil {
			return c.Status(500).Send([]byte(err.Error()))
		}

		cursor.All(context.Background(), &results)

		// no results send not found error
		if results == nil {
			return c.SendStatus(404)
		}

		json, _ := json.Marshal(results)
		return c.Send(json)
	})
	log.Fatal(app.Listen(":8080"))
}
