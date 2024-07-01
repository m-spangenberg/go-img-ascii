package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func main() {
	// Handle command line arguments
	imagePath := flag.String("i", "", "Path to the image file")
	output := flag.String("o", "stdout", "Output option: stdout or png or txt")
	width := flag.Int("w", 64, "Width to scale the image to")
	height := flag.Int("h", 32, "Height to scale the image to")

	// Override the default usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "  -i string")
		fmt.Fprintln(os.Stderr, "    	Path to the image file")
		fmt.Fprintln(os.Stderr, "  -o string")
		fmt.Fprintln(os.Stderr, "    	Output option: stdout or png or txt (default \"stdout\")")
		fmt.Fprintln(os.Stderr, "  -w int")
		fmt.Fprintln(os.Stderr, "    	Width to scale the image to (default 64)")
		fmt.Fprintln(os.Stderr, "  -h int")
		fmt.Fprintln(os.Stderr, "    	Height to scale the image to (default 32)")
	}

	flag.Parse()

	if *imagePath == "" {
		fmt.Println("No image provided. Quitting.")
		os.Exit(1)
	}

	img, err := decodeImage(*imagePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	scaled := scaleImage(img, *width, *height)
	gray := convertToGray(scaled)
	ascii := mapToASCII(gray)

	switch *output {
	case "stdout":
		printToSTDOUT(ascii)
	case "png":
		exportToPNG(ascii, "output.png")
	case "txt":
		exportToTXT(ascii, "output.txt")
	default:
		fmt.Println("Invalid output option. Quitting.")
		os.Exit(1)
	}
}

func decodeImage(imagePath string) (image.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

func scaleImage(img image.Image, width, height int) image.Image {
	bounds := img.Bounds()
	scaled := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := x * bounds.Dx() / width
			srcY := y * bounds.Dy() / height
			scaled.Set(x, y, img.At(srcX, srcY))
		}
	}

	return scaled
}

func convertToGray(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			originalColor := img.At(x, y)
			grayColor := color.GrayModel.Convert(originalColor).(color.Gray)
			gray.SetGray(x, y, grayColor)
		}
	}

	return gray
}

func mapToASCII(img *image.Gray) string {
	bounds := img.Bounds()
	ascii := " .:-=+*#%@"
	buf := make([]byte, 0, bounds.Dx()*bounds.Dy())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.GrayAt(x, y)
			i := int(float64(c.Y) * 10 / 255)
			buf = append(buf, ascii[i])
		}
		buf = append(buf, '\n')
	}

	return string(buf)
}

func printToSTDOUT(ascii string) {
	for _, char := range ascii {
		fmt.Print(string(char))
	}
}

func exportToTXT(ascii string, outputPath string) {
	file, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Error: File could not be created")
		os.Exit(1)
	}
	defer file.Close()

	if _, err := file.WriteString(ascii); err != nil {
		fmt.Println("Error: ASCII could not be written")
		os.Exit(1)
	}
}

func exportToPNG(ascii string, outputPath string) {
	lines := strings.Split(ascii, "\n")
	img := image.NewRGBA(image.Rect(0, 0, len(lines[0])*6, len(lines)*12))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	d := &font.Drawer{
		Dst:  img,
		Src:  image.Black,
		Face: basicfont.Face7x13,
	}

	for y, line := range lines {
		d.Dot = fixed.P(0, (y+1)*12)
		d.DrawString(line)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Error: File could not be created")
		os.Exit(1)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		fmt.Println("Error: Image could not be encoded")
		os.Exit(1)
	}
}
