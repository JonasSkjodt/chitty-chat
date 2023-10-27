package main

import (
	"context"
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net"
	"os"
	"sync"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.
	gRPC "github.com/JonasSkjodt/chitty-chat/proto"
	"google.golang.org/grpc"
)

type chatServer struct {
	gRPC.UnimplementedChatServer        // You need this line if you have a server
	name                         string // Not required but useful if you want to name your server
	port                         string // Not required but useful if your server needs to know what port it's listening to

	mutex sync.Mutex // used to lock the server to avoid race conditions.

}

// flags are used to get arguments from the terminal. Flags take a value, a default value and a description of the flag.
// to use a flag then just add it as an argument when running the program.
var serverName = flag.String("name", "default", "Senders name") // set with "-name <name>" in terminal
var port = flag.String("port", "5400", "Server port")           // set with "-port <port>" in terminal
var vectorClock = []int{0}
var clientID = 1
var nextNumClock = 0 // vector clock for the server

// Maps
var clientNames = make(map[string]gRPC.Chat_MessageStreamServer)
var clientIDs = make(map[string]int)

func main() {

	// f := setLog() //uncomment this line to log to a log.txt file instead of the console
	// defer f.Close()

	// This parses the flags and sets the correct/given corresponding values.
	flag.Parse()
	fmt.Println(".:server is starting:.")

	// launch the server
	launchServer()

	// code here is unreachable because launchServer occupies the current thread.
}

func launchServer() {
	log.Printf("Server %s: Attempts to create listener on port %s\n", *serverName, *port)

	// Create listener tcp on given port or default port 5400
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err) //If it fails to listen on the port, run launchServer method again with the next value/port in ports array
		return
	}

	// makes gRPC server using the options
	// you can add options here if you want or remove the options part entirely
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	// makes a new server instance using the name and port from the flags.
	server := &chatServer{
		name: *serverName,
		port: *port,
	}

	gRPC.RegisterChatServer(grpcServer, server) //Registers the server to the gRPC server.

	log.Printf("Server %s: Listening at %v\n", *serverName, list.Addr())

	if err := grpcServer.Serve(list); err != nil {
		log.Fatalf("failed to serve %v", err)
	}
	// code here is unreachable because grpcServer.Serve occupies the current thread.
}

func DeleteUser(clientName string) {
	if clientName != "" {
		//Deletes the index of the client from the vector clock
		// Curtiousy of https://www.geeksforgeeks.org/delete-elements-in-a-slice-in-golang/
		vectorClock = append(vectorClock[:clientIDs[clientName]], vectorClock[clientIDs[clientName]+1:]...)

		//Reduces the clientID of all clients with a higher ID than the deleted client
		for name := range clientIDs {
			if clientIDs[name] > clientIDs[clientName] {
				clientIDs[name]--
			}
		}

		//TODO Sends the vector clock to all clients with -1 on the deleted client

		//Deletes the client from the clientIDs map
		delete(clientIDs, clientName)

		//Updates the clientID of the next client to connect
		clientID = len(clientIDs) + 1
		//Deletes the client from the clientNames map
		delete(clientNames, clientName)

		log.Printf("%s has diconnected\n", clientName)
	}
}

func (s *chatServer) MessageStream(msgStream gRPC.Chat_MessageStreamServer) error {
	for {
		// get the next message from the stream
		msg, err := msgStream.Recv()
		if err == io.EOF {
			break
		}
		// some other error
		if err != nil {
			return err
		}
		hasher := fnv.New32()
		hasher.Write([]byte(msg.ClientName))
		if msg.Content == fmt.Sprint(hasher.Sum32()) {
			clientNames[msg.ClientName] = msgStream
			clientIDs[msg.ClientName] = clientID
			clientID++

			log.Printf("Participant %s: Connected to the server", msg.ClientName)

			/*
				algorithm for lamport timestamp should fullfill the following (if we dont use vector clock)
				1: When a process does work, increment the counter.
				2: When a process sends a message, include its counter.
				3: When a process receives a message, it takes the maximum of its own counter and the received counter from the message, then increments the counter by one.
			*/
			nextNumClock++                      //increment the vector clcok
			msg.Timestamp = int64(nextNumClock) //set timestamp of message

			SendMessages(&gRPC.ChatMessage{ClientName: "Server", Content: fmt.Sprintf("%s Connected at Lamport Timestamp %d", msg.ClientName, msg.Timestamp)})

			vectorClock = append(vectorClock, 0)

			hasher = nil

		} else {

			// the stream is closed so we can exit the loop
			// log the message
			log.Printf("Received message: from %s: %s", msg.ClientName, msg.Content)
			//lamports clock
			if msg.Timestamp > int64(nextNumClock) {
				nextNumClock = int(msg.Timestamp)
			}
			nextNumClock++
			msg.Timestamp = int64(nextNumClock)
			// send the message to all clients
			SendMessages(msg)

		}
	}

	return nil
}

func SendMessages(msg *gRPC.ChatMessage) {
	//var msg = <-newMessagesChannel
	for name := range clientNames {
		clientNames[name].Send(msg)
	}
}

// Method that disconnects a client from the server
func (s *chatServer) DisconnectFromServer(ctx context.Context, name *gRPC.ClientName) (*gRPC.Ack, error) {
	log.Printf("Participant %s: Disconnected from the server", name.ClientName)
	DeleteUser(name.ClientName)

	//increment vector clock
	nextNumClock++

	//sends a message to the rest of the clients logged into the server that a client who pressed exit left chitty chat
	msg := &gRPC.ChatMessage{
		ClientName: "Server",
		Content:    fmt.Sprintf("Participant %s left Chitty-Chat at Lamport time %d", name.ClientName, nextNumClock), Timestamp: int64(nextNumClock),
	}
	SendMessages(msg) // Broadcast the "left" message
	return &gRPC.Ack{Message: "success"}, nil
}

// Get preferred outbound ip of this machine
// Usefull if you have to know which ip you should dial, in a client running on an other computer
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate("log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
