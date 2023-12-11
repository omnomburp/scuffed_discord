package user

import (
	"context"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// User Data Type
type User struct {
	Username string   `bson:"username"`
	Password string   `bson:"password"`
	Servers  []string `bson:"servers"`
}

// Creates a new User
func HandleCreateUser(w http.ResponseWriter, r *http.Request, ucollection *mongo.Collection) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusBadRequest)
		return
	}

	newUsername := r.FormValue("username")
	newPassword := r.FormValue("password")

	fmt.Println(newUsername)
	fmt.Println(newPassword)

	newUser := User{Username: newUsername, Password: newPassword, Servers: make([]string, 0)}

	insertResult, err := ucollection.InsertOne(context.Background(), newUser)
	if err != nil {
		http.Error(w, "Error creating user: "+err.Error(), http.StatusInternalServerError)
		fmt.Println("Error creating user", err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User Created"))
	fmt.Println("User Created Successfully", insertResult.InsertedID)
}

func HandleJoinServer(w http.ResponseWriter, r *http.Request, ucollection *mongo.Collection) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error joining server", http.StatusBadRequest)
	}

	newServer := r.Form.Get("servername")

	filter := bson.M{"username": r.Form.Get("username")}

	update := bson.M{
		"$push": bson.M{
			"servers": newServer,
		},
	}
	updateResult, err := ucollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		fmt.Println("Cannot update users Servers", err.Error())
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Joined"))

	fmt.Println("Server joined successfully", updateResult.UpsertedID)

}

func HandleLogin(w http.ResponseWriter, r *http.Request, ucollection *mongo.Collection) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		fmt.Println(w, "Failed to Parse form", http.StatusBadRequest)
		http.Error(w, "Failed to Parse form", http.StatusBadRequest)
		return
	}

	userName := r.FormValue("username")
	userPassword := r.FormValue("password")

	filter := bson.M{"username": userName}

	var loginuser User

	err = ucollection.FindOne(context.Background(), filter).Decode(&loginuser)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("User doesn't exist")
		w.Write([]byte("User doesn't exist"))
		return
	}

	if userPassword == loginuser.Password {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Password matches"))
		fmt.Println("Password matches")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Password doesn't match"))
		fmt.Println("Password doesn't match")
	}
}
