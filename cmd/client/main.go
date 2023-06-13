package main

import (
	"context"
	"flag"
	"io"
	"log"
	"time"

	"github.com/daffarg/grpc-pcbook/pb"
	"github.com/daffarg/grpc-pcbook/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func createLaptop(laptopClient pb.LaptopServiceClient) {
	newLaptopObj := sample.NewLaptop()

	createLaptopReq := &pb.CreateLaptopRequest{
		Laptop: newLaptopObj,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := laptopClient.CreateLaptop(ctx, createLaptopReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Print("laptop already exists")
		} else {
			log.Fatal("cannot create laptop: ", err)
		}
	}

	log.Printf("created laptop with id: %s", res.Id)
}

func searchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	searchLaptopReq := &pb.SearchLaptopRequest{
		Filter: filter,
	}

	stream, err := laptopClient.SearchLaptop(ctx, searchLaptopReq)

	if err != nil {
		log.Fatal("cannot search laptop: ", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("cannot receive response: ", err)
		}

		laptop := res.GetLaptop()
		log.Print("Found a laptop with ID: ", laptop.GetId())
		log.Print(" - Brand: ", laptop.GetBrand())
		log.Print(" - Name: ", laptop.GetName())
		log.Print(" - CPU Cores: ", laptop.GetCpu().GetNumberCores())
		log.Print(" - CPU Min GHz: ", laptop.GetCpu().GetMinGhz())
		log.Print(" - RAM: ", laptop.GetRam().GetValue(), laptop.GetRam().GetUnit())
		log.Print(" - Price: ", laptop.GetPriceUsd())
	}
}

func main() {
	serverAddress := flag.String("address", "", "the server address")
	flag.Parse()
	log.Printf("dial server %s", *serverAddress)

	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	laptopClient := pb.NewLaptopServiceClient(conn)
	for i := 0; i < 10; i ++ {
		createLaptop(laptopClient)
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz: 2.5,
		MinRam: &pb.Memory{
			Unit: pb.Memory_GIGABYTE,
			Value: 4,
		},
	}

	searchLaptop(laptopClient, filter)
}