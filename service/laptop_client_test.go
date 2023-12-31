package service_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/daffarg/grpc-pcbook/pb"
	"github.com/daffarg/grpc-pcbook/sample"
	"github.com/daffarg/grpc-pcbook/serializer"
	"github.com/daffarg/grpc-pcbook/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopStore := service.NewInMemoryLaptopStore()
	listenAddr := startTestLaptopServer(t, laptopStore, nil)
	laptopClient := newTestLaptopClient(t, listenAddr)

	laptop := sample.NewLaptop()
	expectedId := laptop.Id

	createLaptopReq := &pb.CreateLaptopRequest{Laptop: laptop}

	res, err := laptopClient.CreateLaptop(context.Background(), createLaptopReq)
	require.NotNil(t, res)
	require.NoError(t, err)
	require.Equal(t, expectedId, res.Id)

	// check the laptop is saved in the server
	returnedLaptop, err := laptopStore.FindById(res.Id)
	require.NoError(t, err)
	require.NotNil(t, returnedLaptop)

	// check that the saved laptop is same as the one we sent
	requireSameLaptop(t, laptop, returnedLaptop)
}	

func TestClientSearchLaptop(t *testing.T) {
	t.Parallel()

	filter := &pb.Filter{
		MaxPriceUsd: 2000,
		MinCpuCores: 4,
		MinCpuGhz: 2.2,
		MinRam: &pb.Memory{
			Value: 8,
			Unit: pb.Memory_GIGABYTE,
		},
	}

	store := service.NewInMemoryLaptopStore()
	expectedId := make(map[string]bool)
	laptop_sent := []*pb.Laptop{}

	for i := 0; i < 6; i ++ {
		laptop := sample.NewLaptop()
		switch i {
			case 0:
				laptop.PriceUsd = 2500
			case 1:
				laptop.Cpu.NumberCores = 2
			case 2:
				laptop.Cpu.MinGhz = 2.0
			case 3:
				laptop.Ram = &pb.Memory{Unit: pb.Memory_MEGABYTE, Value: 4096}
			case 4:
				laptop.PriceUsd = 1999
				laptop.Cpu.NumberCores = 4
				laptop.Cpu.MinGhz = 2.5;
				laptop.Cpu.MaxGhz = 4.5;
				laptop.Ram = &pb.Memory{Unit: pb.Memory_GIGABYTE, Value: 16}
				expectedId[laptop.Id] = true
			case 5:
				laptop.PriceUsd = 2000
				laptop.Cpu.NumberCores = 6
				laptop.Cpu.MinGhz = 2.8;
				laptop.Cpu.MaxGhz = 5.0;
				laptop.Ram = &pb.Memory{Unit: pb.Memory_GIGABYTE, Value: 64}
				expectedId[laptop.Id] = true
		}
		err := store.Save(laptop)
		require.NoError(t, err)
		laptop_sent = append(laptop_sent, laptop)
	}

	fmt.Println(laptop_sent)

	searchLaptopReq := &pb.SearchLaptopRequest{
		Filter: filter,
	}

	serverAddress := startTestLaptopServer(t, store, nil)
	laptopClient := newTestLaptopClient(t, serverAddress)

	stream, err := laptopClient.SearchLaptop(context.Background(), searchLaptopReq)
	require.NoError(t, err)

	found := 0
	iteration := 0
	for {
		fmt.Println("iteration", iteration)
		res, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("EOF")
			break
		}

		fmt.Println(res.GetLaptop())

		require.NoError(t, err)
		require.Contains(t, expectedId, res.GetLaptop().GetId())
		

		found += 1
		iteration += 1
	}

	require.Equal(t, len(expectedId), found)

}	

func TestClientUploadImage(t * testing.T) {
	t.Parallel()

	imageFolder := "../tmp"

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore(imageFolder)

	laptop := sample.NewLaptop()
	err := laptopStore.Save(laptop)
	require.NoError(t, err)

	serverAddress := startTestLaptopServer(t, laptopStore, imageStore)
	serviceClient := newTestLaptopClient(t, serverAddress)

	imagePath := fmt.Sprintf("%s/laptop.png", imageFolder)
	file, err := os.Open(imagePath)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	stream, err := serviceClient.UploadImage(ctx)
	require.NoError(t, err)

	err = stream.Send(&pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId: laptop.GetId(),
				ImageType: filepath.Ext(imagePath),
			},
		},
	})
	require.NoError(t, err)

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024) // create empty buffers (1 megabyte)
	size := 0
	
	for {
		n, err := reader.Read(buffer) // read 1 megabyte image chunk data into buffer
		if err == io.EOF {
			break
		}

		require.NoError(t, err)
		size += n

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		require.NoError(t, err)
	}

	res, err := stream.CloseAndRecv()

	require.NoError(t, err)
	require.NotZero(t, res.GetId())
	require.EqualValues(t, size, res.GetSize())

	err = file.Close()
	require.NoError(t, err)

	err = stream.CloseSend() // Close the stream before removing the file
	require.NoError(t, err)

	savedImagePath := fmt.Sprintf("%s/%s%s", imageFolder, res.GetId(), filepath.Ext(imagePath))
	require.FileExists(t, savedImagePath)
	require.NoError(t, os.Remove(savedImagePath))

	log.Printf("successfully uploaded image with ID = %s and size = %d", res.GetId(), res.GetSize())
} 

func startTestLaptopServer(t *testing.T, laptopStore service.LaptopStore, imageStore service.ImageStore) string {
	laptopServer := service.NewLaptopServer(laptopStore, imageStore)

	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer) // register laptopServer to grpcServer

	listener, err := net.Listen("tcp", ":0") // listen to random available port
	require.NoError(t, err)

	go grpcServer.Serve(listener) // start grpcServer, this method is blocking so it will use goroutine

	return listener.Addr().String() // return laptopServer and the address of the listener
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