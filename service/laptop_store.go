package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/daffarg/grpc-pcbook/pb"
	"github.com/jinzhu/copier"
)

var ErrAlreadyExists = errors.New("record already exists")

type LaptopStore interface {
	Save(laptop *pb.Laptop) error
	FindById(laptopId string) (*pb.Laptop, error)
	Search(ctx context.Context, filter *pb.Filter, found func(*pb.Laptop) error)  error // param2: callback function
}

type InMemoryLaptopStore struct {
	Mutex sync.RWMutex
	Data  map[string]*pb.Laptop
}

func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		Data: make(map[string]*pb.Laptop),
	}
}

func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.Mutex.Lock()
	defer store.Mutex.Unlock()

	if store.Data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	other, err := deepCopy(laptop)

	if err != nil {
		return err
	}

	store.Data[laptop.Id] = other
	return nil
}

func (store *InMemoryLaptopStore) FindById(laptopId string) (*pb.Laptop, error) {
	store.Mutex.RLock()
	defer store.Mutex.RUnlock()

	laptop := store.Data[laptopId]
	if laptop == nil {
		return nil, nil
	}

	return deepCopy(laptop)
}

func (store *InMemoryLaptopStore) Search(ctx context.Context, filter *pb.Filter, found func(*pb.Laptop) error) error {
	store.Mutex.RLock()
	defer store.Mutex.RLocker().Unlock()

	for _, laptop := range store.Data {
		log.Print("checking laptop id: ", laptop.Id)

		if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
			log.Print("context is cancelled")
			return errors.New("context is cancelled")
		}

		if isQualified(filter, laptop) {
			fmt.Println("QUALIFIED")
			other, err := deepCopy(laptop)
			if err != nil {
				return err
			}
			
			err = found(other)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}

	if laptop.GetCpu().GetNumberCores() < filter.GetMinCpuCores() {
		return false
	}

	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}	

	if toBit(laptop.GetRam()) < toBit(filter.GetMinRam()) {
		return false
	}

	return true
}

func toBit(memory *pb.Memory) uint64 {
	value := memory.GetValue()

	switch memory.GetUnit() {
		case pb.Memory_BIT:
			return value
		case pb.Memory_BYTE:
			return value << 3
		case pb.Memory_KILOBYTE:
			return value << 13
		case pb.Memory_MEGABYTE:
			return value << 13
		case pb.Memory_GIGABYTE:
			return value << 23
		case pb.Memory_TERABYTE:
			return value << 33
		default:
			return 0
	}
}

func deepCopy(laptop *pb.Laptop) (*pb.Laptop, error) {
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data : %v", err)
	}
	return other, nil
}