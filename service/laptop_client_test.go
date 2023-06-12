package service_test

import (
	"context"
	"net"
	"testing"

	"github.com/daffarg/grpc-pcbook/pb"
	"github.com/daffarg/grpc-pcbook/sample"
	"github.com/daffarg/grpc-pcbook/serializer"
	"github.com/daffarg/grpc-pcbook/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopServer, listenAddr := startTestLaptopServer(t)
	laptopClient := newTestLaptopClient(t, listenAddr)

	laptop := sample.NewLaptop()
	expectedId := laptop.Id

	createLaptopReq := &pb.CreateLaptopRequest{Laptop: laptop}

	res, err := laptopClient.CreateLaptop(context.Background(), createLaptopReq)
	require.NotNil(t, res)
	require.NoError(t, err)
	require.Equal(t, expectedId, res.Id)

	// check the laptop is saved in the server
	returnedLaptop, err := laptopServer.Store.FindById(res.Id)
	require.NoError(t, err)
	require.NotNil(t, returnedLaptop)

	// check that the saved laptop is same as the one we sent
	requireSameLaptop(t, laptop, returnedLaptop)
}	

func startTestLaptopServer(t *testing.T) (*service.LaptopServer, string) {
	laptopServer := service.NewLaptopServer(service.NewInMemoryLaptopStore())

	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer) // register laptopServer to grpcServer

	listener, err := net.Listen("tcp", ":0") // listen to random available port
	require.NoError(t, err)

	go grpcServer.Serve(listener) // start grpcServer, this method is blocking so it will use goroutine

	return laptopServer, listener.Addr().String() // return laptopServer and the address of the listener
}

func newTestLaptopClient(t *testing.T, serverAddress string) pb.LaptopServiceClient {
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	require.NoError(t, err)
	return pb.NewLaptopServiceClient(conn)
}

func requireSameLaptop(t *testing.T, first *pb.Laptop, second *pb.Laptop) {
	firstLaptopJson, err := serializer.ProtobufToJSON(first)
	require.NoError(t, err)

	secondLaptopJson, err := serializer.ProtobufToJSON(second)
	require.NoError(t, err)

	require.Equal(t, firstLaptopJson, secondLaptopJson)
}