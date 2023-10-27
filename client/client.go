package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"os"
	"strings"
	"time"

	// this has to be the same as the go.mod module,
	// followed by the path to the folder the proto file is in.

	gRPC "github.com/JonasSkjodt/chitty-chat/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Same principle as in client. Flags allows for user specific arguments/values
var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", "5400", "Tcp server")

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
	//defer SendMessage("exit", ChatStream)
	defer ServerConn.Close()

	ChatStream, err := chatServer.MessageStream(context.Background())
	if err != nil {
		log.Fatalf("Error on receive: %v", err)
	}
	hasher := fnv.New32()
	hasher.Write([]byte(*clientsName))

	SendMessage(fmt.Sprint(hasher.Sum32()), ChatStream)

	//start the biding

	go listenForMessages(ChatStream)
	parseInput(ChatStream)
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
	log.Printf("Participant %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
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

func parseInput(stream gRPC.Chat_MessageStreamClient) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to Chitty Chat!")
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

		// check if the input is longer than 128 characters
		if len(input) > 128 {
			log.Println("Message is too long, please enter a message of up to 128 characters.")
			continue
		}

		// when one client disconnects, broadcast the event to other clients
		if input == "exit" {
			chatServer.DisconnectFromServer(context.Background(), &gRPC.ClientName{ClientName: *clientsName})
			os.Exit(1)
		}

		if !conReady(chatServer) {
			log.Printf("Participant %s: something was wrong with the connection to the server :(", *clientsName)
			continue
		}

		if input == "exit" {
			chatServer.DisconnectFromServer(stream.Context(), &gRPC.ClientName{ClientName: *clientsName})
			os.Exit(1)
		} else {
			SendMessage(input, stream)
		}

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

func SendMessage(content string, stream gRPC.Chat_MessageStreamClient) {

	message := &gRPC.ChatMessage{
		Content:    content,
		ClientName: *clientsName,
	}

	i := 0

	if i == 0 {
		i++
		stream.Send(message)
	} else {
		stream.Send(message)
		stream.Send(message) // Server for some reason only reads every second message sent so this is just to clear the "buffer"
	}
}

// watch the god
func listenForMessages(stream gRPC.Chat_MessageStreamClient) {
	for {
		time.Sleep(1 * time.Second)
		if stream != nil {
			msg, err := stream.Recv()
			if err == io.EOF {
				log.Printf("Error: io.EOF in listenForMessages in client.go")
				break
			}
			if err != nil {
				log.Fatalf("%v", err)
			}
			if msg.ClientName != *clientsName {
				log.Printf("%s: %s", msg.ClientName, msg.Content)
			}
		}
	}
}
