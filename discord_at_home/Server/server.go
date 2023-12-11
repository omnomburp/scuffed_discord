package server

import (
	"context"
	user "discord_at_home/DAHUser"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Server Data Type
// Server needs to track its own list of Users next time
// For now only Users will keep track of what Servers they are in
type Server struct {
	ServerName string    `bson:"servername"`
	ChatLog    []Message `bson:"chatlog"`
}

// Message Data Type
type Message struct {
	User    string `json:"user"`
	Content string `json:"content"`
}

type MessageChannel struct {
	User    string `json:"user"`
	Content string `json:"content"`
}

type Client struct {
	ID               string
	ResponseWriter   http.ResponseWriter
	MessageChannel   chan MessageChannel
	DisconnectNotify <-chan bool
}

type ServerChannel struct {
	messageChannel chan MessageChannel
	clients        map[string]*Client
}

var (
	servers         = make(map[string]*ServerChannel)
	serversMutex    sync.Mutex
	clientIDCounter int
	clientIDMutex   sync.Mutex
)

// Creates a newly created Server upon Request
func HandleServerCreate(w http.ResponseWriter, r *http.Request, scollection *mongo.Collection, ucollection *mongo.Collection) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error creating server", http.StatusBadRequest)
	}

	serverName := r.FormValue("servername")

	newServer := Server{ServerName: serverName, ChatLog: make([]Message, 0)}

	insertResult, err := scollection.InsertOne(context.Background(), newServer)
	if err != nil {
		fmt.Println("Unable to insert server")
		return
	}

	userName := r.FormValue("username")
	filter := bson.M{"username": userName}

	var user user.User

	err = ucollection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No matching user found")
			return
		} else {
			fmt.Println(err.Error())
			return
		}
	}

	update := bson.M{
		"$push": bson.M{
			"servers": serverName,
		},
	}

	updateResult, uperr := ucollection.UpdateOne(context.Background(), filter, update)
	if uperr != nil {
		fmt.Println("up error", uperr.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("halp"))

	fmt.Println("Server created successfully", insertResult.InsertedID)
	fmt.Println("Server added to user", updateResult.UpsertedID)

}

// Appends a new message into the Channel Chat Log
func HandleReceiveMessage(w http.ResponseWriter, r *http.Request, scollection *mongo.Collection) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error sending message", http.StatusBadRequest)
		return
	}

	message := r.FormValue("content")
	server := r.FormValue("servername")
	user := r.FormValue("username")

	fmt.Println("Message:", message)
	fmt.Println("Server:", server)
	fmt.Println("User:", user)

	newMessage := Message{User: user, Content: message}

	filter := bson.M{"servername": r.Form.Get("servername")}

	// Define an update to push a new entry to the "chatlog" of the specified channel
	update := bson.M{
		"$push": bson.M{
			"chatlog": newMessage,
		},
	}

	insertResult, err := scollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		fmt.Println("Cannot insert message")
		return
	}

	triggerEvent(server, message, user)

	w.WriteHeader(http.StatusAccepted)
	fmt.Println("Message sent successfully", insertResult.UpsertedID)
}

func HandleLoadServers(w http.ResponseWriter, r *http.Request, ucollection *mongo.Collection) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	userID := r.URL.Query().Get("id")

	filter := bson.M{"username": userID}
	var dahuser user.User

	err := ucollection.FindOne(context.Background(), filter).Decode(&dahuser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No matching user found.")
		} else {
			fmt.Println("Cannot find user to load server", err.Error())
		}
	}

	if dahuser.Servers == nil {
		fmt.Println("NO SERVERS")
		return
	}

	jsonData, jsonerr := json.Marshal(dahuser.Servers)
	if jsonerr != nil {
		http.Error(w, "Error encoding json", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(jsonData)
	fmt.Println("Server data sent back")
}

func HandleLoadChatLog(w http.ResponseWriter, r *http.Request, scollection *mongo.Collection) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	serverName := r.URL.Query().Get("id")

	filter := bson.M{"servername": serverName}

	var res Server

	reserr := scollection.FindOne(context.Background(), filter).Decode(&res)
	if reserr != nil {
		if reserr == mongo.ErrNoDocuments {
			fmt.Println("No matching server found")
		} else {
			http.Error(w, reserr.Error(), http.StatusBadRequest)
		}
	}

	chatlog := res.ChatLog

	if len(chatlog) == 0 {
		fmt.Println("There is no chat")
		return
	}

	jsonData, jsonerr := json.Marshal(chatlog)
	if jsonerr != nil {
		http.Error(w, "Error parsing json", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(jsonData)

	fmt.Println("Chatlog sent successfully")
}

func generateUniqueClientID() string {
	clientIDMutex.Lock()
	defer clientIDMutex.Unlock()

	clientIDCounter++
	return fmt.Sprintf("client_%d", clientIDCounter)
}

func HandleSSE(w http.ResponseWriter, r *http.Request) {
	serverID := r.URL.Query().Get("id")
	if serverID == "" || serverID == " " {
		return
	}

	clientID := generateUniqueClientID()

	client := &Client{
		ID:               clientID,
		ResponseWriter:   w,
		MessageChannel:   make(chan MessageChannel),
		DisconnectNotify: w.(http.CloseNotifier).CloseNotify(),
	}

	// Register the client to the server
	registerClient(serverID, client)

	// Set Content-Type to text/event-stream
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Stream messages to the client
	for {
		select {
		case message, ok := <-client.MessageChannel:
			if !ok {
				fmt.Println("Client channel closed")
				return
			}
			data := `{"content": "` + message.Content + `", "user": "` + message.User + `"}`
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()
		case <-client.DisconnectNotify:
			unregisterClient(serverID, client)
			fmt.Println("Client disconnected")
			return
		case <-r.Context().Done():
			unregisterClient(serverID, client)
			fmt.Println("Request canceled")
			return
		}
	}
}

func triggerEvent(serverID string, messageText string, userID string) {
	serversMutex.Lock()
	server, ok := servers[serverID]
	serversMutex.Unlock()

	if ok {
		// Broadcast the event message to all clients of the specific server
		newMessage := MessageChannel{User: userID, Content: messageText}
		for _, client := range server.clients {
			client.MessageChannel <- newMessage
		}
	}
}

// Functions to register and unregister clients
func registerClient(serverID string, client *Client) {
	serversMutex.Lock()
	server, ok := servers[serverID]
	if !ok {
		server = &ServerChannel{
			clients: make(map[string]*Client),
		}
		servers[serverID] = server
	}
	server.clients[client.ID] = client
	serversMutex.Unlock()
}

func unregisterClient(serverID string, client *Client) {
	serversMutex.Lock()
	server, ok := servers[serverID]
	if ok {
		delete(server.clients, client.ID)
		if len(server.clients) == 0 {
			delete(servers, serverID)
		}
	}
	serversMutex.Unlock()
}

func closeServerChannel(serverID string) {
	serversMutex.Lock()
	defer serversMutex.Unlock()
	if server, ok := servers[serverID]; ok {
		close(server.messageChannel)
	}
}
