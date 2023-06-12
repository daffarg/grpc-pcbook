package service_test

import (
	"context"
	"testing"

	"github.com/daffarg/grpc-pcbook/pb"
	"github.com/daffarg/grpc-pcbook/sample"
	"github.com/daffarg/grpc-pcbook/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerCreateLaptop(t *testing.T) {
	t.Parallel()
	laptopNoID := sample.NewLaptop()
	laptopNoID.Id = ""

	laptopInvalidID := sample.NewLaptop()
	laptopInvalidID.Id = "invalid-uuid"

	laptopDuplicateID := sample.NewLaptop()
	storeDuplicateID := service.NewInMemoryLaptopStore()
	storeDuplicateID.Save(laptopDuplicateID)

	testCases := []struct { // test table 
		name	string
		laptop *pb.Laptop
		store service.LaptopStore
		code codes.Code
	} {
		{
			name: "success_with_id",
			laptop: sample.NewLaptop(),
			store: service.NewInMemoryLaptopStore(),
			code: codes.OK,
		},
		{
			name: "success_no_id",
			laptop: laptopNoID,
			store: service.NewInMemoryLaptopStore(),
			code: codes.OK,
		},
		{
			name: "failure_invalid_id",
			laptop: laptopInvalidID,
			store: service.NewInMemoryLaptopStore(),
			code: codes.InvalidArgument,
		},
		{
			name: "failure_duplicate_id",
			laptop: laptopDuplicateID,
			store: storeDuplicateID,
			code: codes.AlreadyExists,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			req := &pb.CreateLaptopRequest{
				Laptop: test.laptop,
			}

			server := service.NewLaptopServer(test.store)
			res, err := server.CreateLaptop(context.Background(), req)

			if test.code == codes.OK {
				require.NoError(t, err) // should no error returned
				require.NotNil(t, res) // result should not be nil
				require.NotEmpty(t, res.Id) // id should not be nil

				if (len(test.laptop.Id) > 0) { // laptop ID provided by client
					require.Equal(t, test.laptop.Id, res.Id) // laptop ID provided by client should be equal to returned laptop ID
				}
			} else {
				require.Error(t, err)
				require.Nil(t, res)
				st, ok := status.FromError(err) // get status code from error
				require.True(t, ok) // should be true
				require.Equal(t, test.code, st.Code()) // status code should be equal to test code
			}
		})
	}
}