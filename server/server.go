package server

import (
	proto "Handin5/grpc"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"strconv"
	"time"
	"syscall"
	"os/signal"
)

type Server struct {
	proto.UnimplementedAuctionServer
	nrClients  int32
	highestBid int32
	higestBidder	int32
	isOver     bool
	isAuction  bool
	myPort	   int32
	logger     *log.Logger
}

func (s *Server) timer() {
	time.Sleep(10 * time.Minute)
	s.isOver = true
}

func (s *Server) Results(ctx context.Context, in *proto.Empty) (*proto.Result, error) {
	return &proto.Result{Isover: s.isOver, Highestbid: s.highestBid, Clientid: s.higestBidder}, nil
}

func (s *Server) Bid(ctx context.Context, in *proto.BidPackage) (*proto.Ack, error) {
	if !s.isAuction {
		s.isAuction = true
		go s.timer()
	}
	if s.isOver {
		return &proto.Ack{Ack: "Exception"}, nil
	}
	s.higestBidder = in.Clientid.Clientid
	bidString := in.Amount.Amount
	currentBid, err := strconv.Atoi(bidString)
	if err != nil {
		return &proto.Ack{Ack: "Exception"}, nil
	}

	currentBidInt32 := int32(currentBid)

	if currentBidInt32 <= s.highestBid {
		s.logger.Printf("Bid of %d failed\n", currentBidInt32)

		return &proto.Ack{Ack: "Fail"}, nil
	}

	s.logger.Printf("Bid of %d was a success\n", currentBidInt32)

	s.highestBid = currentBidInt32
	return &proto.Ack{Ack: "Success"}, nil
}

func (s *Server) CreateClientIdentifier(ctx context.Context, in *proto.Empty) (*proto.ClientId, error) {
	s.nrClients += 1

	return &proto.ClientId{Clientid: s.nrClients}, nil
}

var sigChan = make(chan os.Signal, 1)

func (s *Server) Disconnect() {
	<-sigChan
	log.Fatalf("Server with port %d has been shutdown", s.myPort)
	os.Exit(0)
}

func Run(myPort int) {
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	logger := log.New(logFile, "", log.LstdFlags)

	server := &Server{
		nrClients:  0,
		highestBid: 0,
		isOver:     false,
		isAuction:  false,
		myPort: 	int32(myPort),
		logger:     logger,
	}

	server.Disconnect()
	server.start_server(myPort)
}

func (s *Server) start_server(myPort int) {
	gRPCserver := grpc.NewServer()

	netListener, err := net.Listen("tcp", fmt.Sprintf(":%d", myPort))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Server started on port: %d \n", myPort)

	s.logger.Printf("Server started on port: %d \n", myPort)

	proto.RegisterAuctionServer(gRPCserver, s)

	err = gRPCserver.Serve(netListener)
	if err != nil {
		panic(err)
	}
}
