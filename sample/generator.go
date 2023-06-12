package sample

import (
	"github.com/daffarg/grpc-pcbook/pb"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
)

// Generate new random keyboard
func NewKeyboard() *pb.Keyboard {
	keyboard := &pb.Keyboard{
		Layout:  randomKeyboardLayout(),
		Backlit: randomBool(),
	}
	return keyboard
}

// Generate new CPU
func NewCPU() *pb.CPU {
	brand := randomCPUBrand()
	cpuName := randomCPUName(brand)

	numberCores := randomInt(2, 8)
	numberThreads := randomInt(numberCores, 12)

	minGhz := randomFloat64(2.0, 3.5)
	maxGhz := randomFloat64(minGhz, 5.0)

	cpu := &pb.CPU{
		Brand: brand,
		Name:  cpuName,
		NumberCores: uint32(numberCores),
		NumberThreads: uint32(numberThreads),
		MinGhz: minGhz,
		MaxGhz: maxGhz,
	}

	return cpu
}

func NewGPU() *pb.GPU {
	brand := randomGPUBrand()
	gpuName := randomGPUName(brand)

	minGhz := randomFloat64(1.0, 1.5)
	maxGhz := randomFloat64(minGhz, 2.0)

	gpu := &pb.GPU{
		Brand: brand,
		Name:  gpuName,
		MinGhz: minGhz,
		MaxGhz: maxGhz,
		Memory: &pb.Memory{
			Value: uint64(randomInt(2, 6)),
			Unit: pb.Memory_GIGABYTE,
		},
	}

	return gpu
}

func NewStorage() *pb.Storage {
	storage := &pb.Storage{
		Driver: randomDriver(),
		Memory: &pb.Memory{
			Value: uint64(randomInt(128, 1024)),
			Unit: pb.Memory_GIGABYTE,
		},
	}

	return storage
}

func NewRAM() *pb.Memory {
	memory := &pb.Memory{
		Value: uint64(randomInt(4, 64)),
		Unit: pb.Memory_GIGABYTE,
	}

	return memory
}

func NewScreen() *pb.Screen {
	height := randomInt(1080, 4320)
	width := height * 16 / 9

	screen := &pb.Screen{
		SizeInch: float32(randomFloat64(13, 17)),
		Resolution: &pb.Screen_Resolution{
			Width: uint32(width),
			Height: uint32(height),
		},
		Panel: randomPanel(),
		Multitouch: randomBool(),
	}

	return screen
}

func NewLaptop() *pb.Laptop {
	brand := randomLaptopBrand()
	name := randomLaptopName(brand)

	laptop := &pb.Laptop{
		Id: uuid.New().String(),
		Name: name,
		Brand: brand,
		Cpu: NewCPU(),
		Ram: NewRAM(),
		Gpus: []*pb.GPU{NewGPU()},
		Storages: []*pb.Storage{NewStorage()},
		Screen: NewScreen(),
		Keyboard: NewKeyboard(),
		Weight: &pb.Laptop_WeightKg{
			WeightKg: randomFloat64(1.0, 3.0),
		},
		PriceUsd: randomFloat64(1500, 3000),
		ReleaseYear: uint32(randomInt(2015, 2023)),
		UpdatedAt: ptypes.TimestampNow(),
	}

	return laptop
}