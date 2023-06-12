package serializer_test

import (
	"fmt"
	"testing"

	"github.com/daffarg/grpc-pcbook/pb"
	"github.com/daffarg/grpc-pcbook/sample"
	"github.com/daffarg/grpc-pcbook/serializer"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestWriteProtobufToBinaryFile(t *testing.T) {
	t.Parallel()

	binaryFile := "../tmp/laptop.bin"
	jsonFile := "../tmp/laptop.json"

	laptop1 := sample.NewLaptop()

	fmt.Println("Sent laptop")
	fmt.Println(laptop1)
	
	err := serializer.WriteProtobufToBinaryFile(laptop1, binaryFile)
	require.NoError(t, err)

	laptop2 := &pb.Laptop{}

	err = serializer.ReadProtobufFromBinaryFile(binaryFile, laptop2)
	require.NoError(t, err)

	fmt.Println("Received laptop")
	fmt.Println(laptop2)
	require.True(t, proto.Equal(laptop1, laptop2))

	err = serializer.WriteProtobufToJSONFile(laptop1, jsonFile)
	require.NoError(t, err)
}
