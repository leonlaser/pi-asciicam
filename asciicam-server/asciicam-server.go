package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/net/websocket"
)

var (
	channels       = []chan string{}
	levels         = []byte(" .,:;i1tLfCG08@")
	nl             = []byte("\n")
	numLevels      = float64(len(levels) - 1)
	contrastFactor int
	contrast       int
	brightness     int
	height         int
	width          int
	fps            int
	addr           string
	network        string
	help           bool
)

func arguments() {
	flag.BoolVar(&help, "help", false, "show help")
	flag.IntVar(&contrast, "c", 128, "contrast")
	flag.IntVar(&brightness, "b", 100, "brightness")
	flag.IntVar(&width, "w", 160, "raspivid width")
	flag.IntVar(&height, "h", 120, "raspivid height")
	flag.IntVar(&fps, "fps", 10, "raspivid frames per second")
	flag.StringVar(&addr, "addr", ":8000", "Address to allow websocket connection on e.g. :8000, 10.0.0.55:23123")
	flag.StringVar(&network, "net", "", "use a network MJPEG stream instead of local raspivid e.g. 10.0.0.1:5001")
	flag.Parse()

	contrastFactor = (259 * (contrast + 255)) / (255 * (259 - contrast))
}

func ascii(img image.Image) string {
	res := []byte{}

	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y += 2 {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			r, g, b := getRGB(img.At(x, y))
			rf := brightenColor(contrastColor(r))
			gf := brightenColor(contrastColor(g))
			bf := brightenColor(contrastColor(b))
			gc := ((0.299*float64(rf) + 0.587*float64(gf) + 0.114*float64(bf)) / 255)
			res = append(res, levels[level(gc)])
		}
		res = append(res, nl...)
	}

	return string(res)
}

func getRGB(c color.Color) (r, g, b int) {
	red, green, blue, _ := c.RGBA()
	r = int(red >> 8)
	g = int(green >> 8)
	b = int(blue >> 8)
	return
}

func level(b float64) int {
	return int(math.Abs(numLevels - math.Floor(b*numLevels)))
}

func contrastColor(c int) int {
	return truncate(((c - 128) * contrastFactor) + 128)
}

func brightenColor(c int) int {
	return truncate(c + brightness)
}

func truncate(f int) int {
	if f > 255 {
		return 255
	}
	if f < 0 {
		return 0
	}
	return f
}

func sourceRaspivid() io.ReadCloser {
	param := strings.Fields(fmt.Sprintf("-o - -w %d -h %d -n -t 0 -cd MJPEG -fps %d", width, height, fps))
	cmd := exec.Command("raspivid", param...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	cmd.Start()
	return stdout
}

func sourceNetwork() net.Conn {
	conn, err := net.Dial("tcp", network)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func startStream() {
	fmt.Print("Starting stream from ")

	var r io.Reader

	if network == "" {
		fmt.Println("raspivid")
		r = sourceRaspivid()
	} else {
		fmt.Println(network)
		r = sourceNetwork()
	}

	scanner := bufio.NewScanner(r)

	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		for i := 0; i < len(data)-1; i++ {
			// Scan for end of image marker
			if data[i] == byte(0xFF) && data[i+1] == byte(0xD9) {
				return i + 2, data[:i], nil
			}
		}
		return 0, nil, nil
	}

	scanner.Split(split)

	for scanner.Scan() {
		b := scanner.Bytes()
		if len(b) > 0 {
			// i failed to use the splitter correctly ... adding EOI marker ... maybe I am too tired
			b = append(b, []byte{0xFF, 0xD9}...)
			frame := decodeJPEG(bytes.NewReader(b))
			ascii := ascii(frame)
			// send frame to every active ws
			for _, c := range channels {
				c <- ascii
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Invalid input: %s", err)
	}
}

func decodeJPEG(reader io.Reader) image.Image {
	img, err := jpeg.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	return img
}

func stream(ws *websocket.Conn) {
	fmt.Println("New connection from", ws.Request().RemoteAddr)
	c := make(chan string)
	channels = append(channels, c)
	for ascii := range c {
		websocket.Message.Send(ws, ascii)
	}
	ws.Close()
}

func main() {
	arguments()

	fmt.Println("Settings")
	fmt.Println(fmt.Sprintf("  Size: %dx%d", width, height))
	fmt.Println("  Address:", addr)
	fmt.Println("  Contrast:", contrast)
	fmt.Println("  Brightness:", brightness)
	fmt.Println("  FPS:", fps)

	if help {
		flag.Usage()
		os.Exit(0)
	}

	go startStream()

	http.Handle("/", websocket.Handler(stream))

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: " + err.Error())
	}
}
