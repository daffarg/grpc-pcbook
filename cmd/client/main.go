package main

import (
	"bufio"
	"context"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/daffarg/grpc-pcbook/pb"
	"github.com/daffarg/grpc-pcbook/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func createLaptop(laptopClient pb.LaptopServiceClient, laptop *pb.Laptop) {
	createLaptopReq := &pb.CreateLaptopRequest{
		Laptop: laptop,
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

func uploadImage(laptopClient pb.LaptopServiceClient, laptopId string, imagePath string) {
	// open an image from folder
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatalf("cannot open image file : %v", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	stream, err := laptopClient.UploadImage(ctx)
	if err != nil {
		log.Fatalf("cannot upload image : %v", err)
	}

	// first send an image info to the server
	req := &pb.UploadImageRequest {
		Data: &pb.UploadImageRequest_Info {
			Info: &pb.ImageInfo {
			LaptopId: laptopId,
			ImageType: filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)

	if err != nil {
		log.Fatalf("cannot send image info : %v %v", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024) // create empty buffers (1 megabyte)
	
	for {
		n, err := reader.Read(buffer) // read 1 megabyte image chunk data into buffer
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("cannot read chunk into buffer : %v", err)
		}

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatalf("cannot send chunk to server : %v %v", err, stream.RecvMsg(nil))
		}
	}

	res, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("cannot receive response from server : %v", err)
	}

	log.Printf("successfully uploaded image with ID = %s and size = %d", res.GetId(), res.GetSize())
}

func testCreateLaptop(laptopClient pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	createLaptop(laptopClient, laptop)
}

func testSearchLaptop(laptopClient pb.LaptopServiceClient) {
	for i := 0; i < 10; i ++ {
		createLaptop(laptopClient, sample.NewLaptop())
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

func testUploadImage(laptopClient pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	log.Printf("make a request to server to create laptop with ID = %s", laptop.Id)
	createLaptop(laptopClient, laptop)
	uploadImage(laptopClient, laptop.Id, "tmp/laptop.png")
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

	testCreateLaptop(laptopClient)
	testSearchLaptop(laptopClient)
	testUploadImage(laptopClient)
}