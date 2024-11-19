package server

import (
	proto "Handin5/grpc"
	"context"
	"errors"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	proto.UnimplementedAuctionServer
	nrClients  int32
	highestBid int32
}

func (s *Server) Bid(ctx context.Context, in *proto.Amount) (*proto.Ack, error) {
	curMessage := in.Amount

	if curMessage <= 0 { //YOU BID NO MANIES
		return &proto.Ack{Ack: "CHANGE THIS !!!!"}, errors.New("message longer than 128 characters")
	} else if curMessage <= s.highestBid {
		s.highestBid = curMessage
	}

	return &proto.Ack{Ack: "CHANGE THIS !!!!"}, nil
}

func (s *Server) CreateClientIdentifier(ctx context.Context, in *proto.Empty) (*proto.ClientId, error) {
	s.nrClients += 1

	return &proto.ClientId{Clientid: s.nrClients}, nil
}

func Run() {
	server := &Server{
		nrClients:  0,
		highestBid: 0,
	}

	server.start_server()
}

func (s *Server) start_server() {

	gRPCserver := grpc.NewServer()

	log.Println("Server started")

	netListener, err := net.Listen("tcp", ":5050")
	if err != nil {
		panic(err)
	}

	proto.RegisterAuctionServer(gRPCserver, s)

	err = gRPCserver.Serve(netListener)
	if err != nil {
		panic(err)
	}
}
