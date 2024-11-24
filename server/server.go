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
)

type Server struct {
	proto.UnimplementedAuctionServer
	nrClients  int32
	highestBid int32
	isOver     bool
	isAuction  bool
	logger     *log.Logger
}

func (s *Server) timer() {
	time.Sleep(10 * time.Minute)
	s.isOver = true
}

func (s *Server) Results(ctx context.Context, in *proto.Empty) (*proto.Result, error) {
	return &proto.Result{Isover: s.isOver, Highestbid: s.highestBid}, nil
}

func (s *Server) Bid(ctx context.Context, in *proto.Amount) (*proto.Ack, error) {
	if !s.isAuction {
		s.isAuction = true
		go s.timer()
	}
	if s.isOver {
		return &proto.Ack{Ack: "Exception"}, nil
	}

	bidString := in.Amount
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

func Run(myPort int) {
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
		logger:     logger,
	}

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
