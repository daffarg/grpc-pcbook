package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"time"

	"github.com/daffarg/grpc-pcbook/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxImageSize = 1 << 20 // one megabyte

type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer
	LaptopStore LaptopStore
	ImageStore ImageStore
}

func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore) *LaptopServer {
	return &LaptopServer{LaptopStore: laptopStore, ImageStore: imageStore}
}

func (server *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("Receiving create laptop request with id : %s", laptop.Id)

	if (len(laptop.Id) > 0) { // laptop id provided by the client
		_, err := uuid.Parse(laptop.Id) // check if laptop id valid

		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Laptop ID is not a valid UUID : %v", err)
		}
	} else {
		id, err := uuid.NewRandom()
		
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to create new UUID for the laptop : %v", err)
		} 	
		laptop.Id = id.String() // set the laptop ID with new generated id
	}

	// pretending do heavy computation
	// time.Sleep(6 * time.Second)

	// check if the context is cancelled
	if ctx.Err() == context.Canceled {
		log.Print("Request is canceled")
		return nil, status.Error(codes.Canceled, "request is cancelled")
	}

	// check context error to see if the deadline exceeded
	if ctx.Err() == context.DeadlineExceeded {
		log.Print("deadline is exceeded")
		return nil, status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	}

	// save the laptop to in-memory store
	err := server.LaptopStore.Save(laptop)

	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExists) {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "Failed to save new laptop : %v", err)
	}

	log.Printf("Successfully saved new laptop with id : %s", laptop.Id)
	response := &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}

	return response, nil
}

func (server *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream pb.LaptopService_SearchLaptopServer) error {
	filter := req.GetFilter()

	log.Printf("receive search laptop request with filter : %v", filter)

	err := server.LaptopStore.Search(
		stream.Context(),
		filter, 
		func (laptop *pb.Laptop) error { // call back function: send laptop stream to client
			log.Print(time.Now())
			
			res := &pb.SearchLaptopResponse{Laptop: laptop}
		
			err := stream.Send(res)

			if err != nil {
				return err
			}

			log.Printf("sent laptop with id : %s", laptop.GetId())
			return nil
		},
	)

	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error %v", err)
	}

	return nil
}	

func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
	req, err := stream.Recv() // receive image info from client

	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive image info"))
	}

	laptopId := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()

	// check if laptop exists
	laptop, err := server.LaptopStore.FindById(laptopId)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot find laptop with ID = %s : %v", laptopId, err))
	}
	if laptop == nil { // laptop doesn't exists
		return logError(status.Errorf(codes.InvalidArgument, "laptop with ID = %s doesn't exists", laptopId, err))
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		log.Printf("waiting to receive more image data")

		req, err := stream.Recv() 

		if err == io.EOF {
			log.Printf("no more image data")
			break
		}
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
		}

		chunk := req.GetChunkData() // get image chunk data
		size := len(chunk)

		imageSize += size // increase total image size
		if imageSize > maxImageSize {
			return logError(status.Errorf(codes.InvalidArgument, "image size larger than maximum size : %d > %d", imageSize, maxImageSize))
		}

		_, err = imageData.Write(chunk) // write chunk data received from client to image data buffer
		if err != nil {
			return logError(status.Errorf(codes.Unknown, "cannot write chunk data into the buffer : %v", err))
		}
	}

	// after successfully received all chunk data, store the image
	imageId, err := server.ImageStore.Save(laptopId, imageType, imageData)
	
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot write chunk data into the buffer : %v", err))
	}

	res := &pb.UploadImageResponse{
		Id: imageId,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send response : %v", err))
	}

	log.Printf("saved an image with id = %s and size = %d", imageId, imageSize)
	return nil
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}