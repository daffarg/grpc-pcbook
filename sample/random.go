package sample

import (
	"math/rand"
	"time"

	"github.com/daffarg/grpc-pcbook/pb"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomBool() bool {
	// generate a random int from 0 to 1 and return true if that int equals to 1
	return rand.Intn(2) == 1 
}

func randomLaptopBrand() string {
	return randomStringFromSet("Apple", "Dell", "Lenovo", "ASUS", "HP", "Acer", "MSI")
}

func randomLaptopName(brand string) string {
	switch brand {
		case "Apple":
			return randomStringFromSet("Macbook Air", "Macbook Pro")
		case "Dell":
			return randomStringFromSet("Latitude", "Vostro", "XPS", "Alienware")
		case "Lenovo":
			return randomStringFromSet("Thinkpad X1", "Thinkpad P1", "Thinkpad P53", "Thinkpad P73")
		case "ASUS":
			return randomStringFromSet("ROG", "Vivobook")
		case "HP":
			return randomStringFromSet("Omen", "Pavoilion", "Elitebook", "Spectre")
		case "Acer":
			return randomStringFromSet("Swift", "Aspire", "Spin", "Nitro")
		case "MSI":
			return randomStringFromSet("Titan", "Raider")
		default:
			return "New Laptop Name"
	}
}

func randomDriver() pb.Storage_Driver {
	if rand.Intn(2) == 1 {
		return pb.Storage_SSD
	}

	return pb.Storage_HDD
}

func randomPanel() pb.Screen_Panel {
	if (rand.Intn(2) == 1) {
		return pb.Screen_IPS
	}

	return pb.Screen_OLED
}

func randomKeyboardLayout() pb.Keyboard_Layout {
	switch rand.Intn(3) {
		case 1:
			return pb.Keyboard_QWERTY
		case 2:
			return pb.Keyboard_QWERTZ
		default:
			return pb.Keyboard_AZERTY
	}
}

func randomGPUBrand() string {
	return randomStringFromSet("NVIDIA", "AMD")
}

func randomCPUBrand() string {
	return randomStringFromSet("Intel", "AMD")
}

func randomStringFromSet(strings ... string) string {
	return strings[rand.Intn(len(strings))]
}

func randomCPUName(brand string) string {
	if brand == "Intel" {
		return randomStringFromSet(
			"Xeon E-2286M",
			"Core i9-9980HK",
			"Core i7-9750H",
			"Core i5-9400F",
			"Core i3-1005G1",
		)
	} 

	return randomStringFromSet(
		"Ryzen 7 PRO 2700U",
		"Ryzen 5 PRO 3500U",
		"Ryzen 3 PRO 3200GE",
	)
}

func randomGPUName(brand string) string {
	if brand == "NVIDIA" {
		return randomStringFromSet(
			"GTX 1060 TI",
			"GTX 1050 TI",
			"GTX 1660 TI",
			"RTX 2060",
		)
	}

	return randomStringFromSet(
		"RX 590",
		"RX 5700 XT",
		"RX 5500 XT",
		"RX 5600 XT",
	)
}

func randomInt(min int, max int) int {
	return min + rand.Intn(max-min+1)
}

func randomFloat64(min float64, max float64) float64 {
	return min + rand.Float64()*(max-min)
}