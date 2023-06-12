package service

import (
	"errors"
	"fmt"
	"sync"

	"github.com/daffarg/grpc-pcbook/pb"
	"github.com/jinzhu/copier"
)

var ErrAlreadyExists = errors.New("record already exists")

type LaptopStore interface {
	Save(laptop *pb.Laptop) error
	FindById(laptopId string) (*pb.Laptop, error)
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

	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)

	if err != nil {
		return fmt.Errorf("cannot copy laptop data : %w", err)
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

	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data : %v", err)
	}
	return other, nil
}