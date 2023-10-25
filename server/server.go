package main

import (
	"flag"
	"fmt"
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
var vectorClock = []int{}
var nextID = 0 // vector clock for the server

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

// The method format can be found in the pb.go file. If the format is wrong, the server type will give an error.
// func (s *Server) Increment(ctx context.Context, Amount *gRPC.Amount) (*gRPC.Ack, error) {
// 	// locks the server ensuring no one else can increment the value at the same time.
// 	// and unlocks the server when the method is done.
// 	s.mutex.Lock()
// 	defer s.mutex.Unlock()

// 	// increments the value by the amount given in the request,
// 	// and returns the new value.
// 	s.incrementValue += int64(Amount.GetValue())
// 	return &gRPC.Ack{NewValue: s.incrementValue}, nil
// }

func (s *chatServer) MessageStream(msgStream gRPC.Chat_MessageStreamServer) error {
	for {
		// get the next message from the stream
		msg, err := msgStream.Recv()

		// the stream is closed so we can exit the loop
		if err == io.EOF {
			break
		}
		// some other error
		if err != nil {
			return err
		}
		// log the message
		log.Printf("Received message: from %s: %s", msg.ClientName, msg.Content)

		msgStream.Send(msg)
	}

	// be a nice server and say goodbye to the client :)

	return nil
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

// testing messaging system start

var chatHistory []*gRPC.ChatMessage
var newMessagesChannel = make(chan *gRPC.ChatMessage)

// func (s *chatServer) SendMessage(ctx context.Context, msg *gRPC.ChatMessage) (*gRPC.Ack, error) {
// 	log.Printf("Received message: from %s: %s", msg.ClientName, msg.Content) // Log the message
// 	chatHistory = append(chatHistory, msg)
// 	newMessagesChannel <- msg

// }

// func (s *chatServer) ReceiveMessageStream(details *gRPC.ChatMessage, stream gRPC.Chat_ReceiveMessageStreamServer) error {
// 	newClient := &client{
// 		stream: stream,
// 		stop:   make(chan bool),
// 	}

// 	clients = append(clients, newClient) // Add new client to the slice

// 	go func() {
// 		for {
// 			select {
// 			case msg := <-newMessagesChannel:
// 				log.Printf("New message received from %s: %s", msg.ClientName, msg.Content)
// 				for _, client := range clients {
// 					if err := client.stream.Send(msg); err != nil {
// 						log.Printf("Error while sending message to client: %v", err)
// 					}
// 				}
// 			case <-newClient.stop:
// 				return
// 			}
// 		}
// 	}()

// 	<-stream.Context().Done()

// 	newClient.stop <- true // Stop the go routine when client disconnects

// 	return stream.Context().Err()
// }

//testing messaging system end
