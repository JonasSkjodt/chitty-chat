package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.

	gRPC "github.com/JonasSkjodt/chitty-chat/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Same principle as in client. Flags allows for user specific arguments/values
var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", "5400", "Tcp server")

// var server gRPC.ChatClient  //the server
var ServerConn *grpc.ClientConn //the server connection
var chatServer gRPC.ChatClient  // new chat server client

func main() {
	//parse flag/arguments
	flag.Parse()

	fmt.Println("--- CLIENT APP ---")

	//log to file instead of console
	//f := setLog()
	//defer f.Close()

	//connect to server and close the connection when program closes
	fmt.Println("--- join Server ---")
	ConnectToServer()
	defer ServerConn.Close()

	//start the biding
	parseInput()
}

// connect to server
func ConnectToServer() {

	//dial options
	//the server is not using TLS, so we use insecure credentials
	//(should be fine for local testing but not in the real world)
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	//dial the server, with the flag "server", to get a connection to it
	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	//for the chat implementation
	// makes a client from the server connection and saves the connection
	// and prints rather or not the connection was is READY
	chatServer = gRPC.NewChatClient(conn)
	ServerConn = conn
	log.Println("the connection is: ", conn.GetState().String())
}

func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("--------------------")

	//Infinite loop to listen for clients input.
	for {
		fmt.Print("-> ")

		//Read input into var input and any errors into err
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input) //Trim input

		if !conReady(chatServer) {
			log.Printf("Client %s: something was wrong with the connection to the server :(", *clientsName)
			continue
		}

		//Convert string to int64, return error if the int is larger than 32bit or not a number
		//val, err := strconv.ParseInt(input, 10, 64)
		// if err != nil {
		// 	if input == "hi" {
		// 		sayHi()
		// 	}
		// 	continue
		// }
		sendMessage(input)
	}
}

// Function which returns a true boolean if the connection to the server is ready, and false if it's not.
func conReady(s gRPC.ChatClient) bool {
	return ServerConn.GetState().String() == "READY"
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}

// testing messaging system start
func sendMessage(text string) {
	// Check if the message length exceeds 128 characters
	/*if len(text) > 128 {
		log.Fatalf("Failed to send message: message length exceeds 128 characters.")
		return
	}*/

	msg := &gRPC.ChatMessage{
		ClientName: *clientsName,
		Content:    text,
	}

	ack, err := chatServer.SendMessage(context.Background(), msg)
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}
	log.Printf("Server acknowledged with: %s", ack.Status)
}

func receiveMessage() {
	details := &gRPC.ClientName{
		ClientName: *clientsName,
	}
	stream, err := chatServer.ReceiveMessageStream(context.Background(), details)
	if err != nil {
		log.Fatalf("Error on receive: %v", err)
	}
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v", err)
		}
		log.Printf("Received message %s from %s", msg.Content, msg.ClientName)
	}
}

//testing messaging system end
