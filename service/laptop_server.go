package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/daffarg/grpc-pcbook/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer
	Store LaptopStore
}

func NewLaptopServer(store LaptopStore) *LaptopServer {
	return &LaptopServer{Store: store}
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
	err := server.Store.Save(laptop)

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

	err := server.Store.Search(
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