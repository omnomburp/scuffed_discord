package main

import (
	"context"
	dahuser "discord_at_home/DAHUser"
	dahserver "discord_at_home/Server"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	serverCollection *mongo.Collection
	userCollection   *mongo.Collection
)

const (
	mongoURI         = "mongodb://localhost:27017"
	dbName           = "discordathome"
	servercollection = "server"
	usercollection   = "user"
)

func init() {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Println("Mongo Client failed to be initialised")
		return
	}

	db := client.Database(dbName)

	err = db.CreateCollection(context.Background(), servercollection)
	if err != nil {
		fmt.Println("Collection ", serverCollection, " already exists")
	} else {
		fmt.Println("Collection ", servercollection, " created")
	}

	err = db.CreateCollection(context.Background(), usercollection)
	if err != nil {
		fmt.Println("Collection ", usercollection, " already exists")
	} else {
		fmt.Println("Collection ", usercollection, " created")
	}

	serverCollection = db.Collection(servercollection)
	// Clear collection for testing
	//serverCollection.DeleteMany(context.Background(), bson.D{})
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"servername": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err = serverCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		fmt.Println("unable to make field unique")
	} else {
		fmt.Println("servername field is now unique")
	}

	userCollection = db.Collection(usercollection)
	// Clear collection for testing
	//userCollection.DeleteMany(context.Background(), bson.D{})
	userModel := mongo.IndexModel{
		Keys:    bson.M{"username": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err = userCollection.Indexes().CreateOne(context.Background(), userModel)
	if err != nil {
		fmt.Println("unable to make field unique")
	} else {
		fmt.Println("username field is now unique")
	}
}

func main() {

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		dahuser.HandleCreateUser(w, r, userCollection)
	})

	http.HandleFunc("/createserver", func(w http.ResponseWriter, r *http.Request) {
		dahserver.HandleServerCreate(w, r, serverCollection, userCollection)
	})

	http.HandleFunc("/sendmessage", func(w http.ResponseWriter, r *http.Request) {
		dahserver.HandleReceiveMessage(w, r, serverCollection)
	})

	http.HandleFunc("/joinserver", func(w http.ResponseWriter, r *http.Request) {
		dahuser.HandleJoinServer(w, r, userCollection)
	})

	http.HandleFunc("/loadservers", func(w http.ResponseWriter, r *http.Request) {
		dahserver.HandleLoadServers(w, r, userCollection)
	})

	http.HandleFunc("/loadchat", func(w http.ResponseWriter, r *http.Request) {
		dahserver.HandleLoadChatLog(w, r, serverCollection)
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		dahuser.HandleLogin(w, r, userCollection)
	})

	http.HandleFunc("/sse", dahserver.HandleSSE)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server", err)
	}

}
