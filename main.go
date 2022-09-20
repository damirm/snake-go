package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"golang.org/x/term"
)

const (
	FPS           = 4
	CH_BODY uint8 = 1
	CH_FOOD uint8 = 2
)

type Config struct {
	Width  int
	Height int
	Speed  float64
}

type Vec2i struct {
	x, y int
}

func NewVec2i(x, y int) Vec2i {
	return Vec2i{x: x, y: y}
}

func (v Vec2i) Add(other Vec2i) Vec2i {
	return NewVec2i(v.x+other.x, v.y+other.y)
}

func (v Vec2i) Mul(other Vec2i) Vec2i {
	return NewVec2i(v.x*other.x, v.y*other.y)
}

func (v Vec2i) Eq(other Vec2i) bool {
	return v.x == other.x && v.y == other.y
}

type Snake struct {
	// buf is a "canvas"
	buf [][]uint8
	// body is a snake body points
	body []Vec2i
	// food is a food coord
	food Vec2i

	direction Vec2i
	quit      bool

	w, h int
}

func NewSnake(c *Config) *Snake {
	buf := make([][]uint8, c.Height)
	for i := 0; i < c.Height; i++ {
		buf[i] = make([]uint8, c.Width)
	}

	// TODO: How to properly initialize the body?
	body := make([]Vec2i, 4, c.Width*c.Height)
	body[0] = NewVec2i(3, 1)
	body[1] = NewVec2i(2, 1)
	body[2] = NewVec2i(1, 1)
	body[3] = NewVec2i(0, 1)

	for _, p := range body {
		buf[p.y][p.x] = CH_BODY
	}

	food := NewVec2i(rand.Intn(c.Height), rand.Intn(c.Width))
	buf[food.y][food.x] = CH_FOOD

	return &Snake{
		quit:      false,
		body:      body,
		buf:       buf,
		food:      food,
		direction: NewVec2i(1, 0),
		w:         c.Width,
		h:         c.Height,
	}
}

func (s *Snake) spawnFood() {
	// TODO: Check body collision.
	s.food = NewVec2i(rand.Intn(s.h), rand.Intn(s.w))
	s.buf[s.food.y][s.food.x] = CH_FOOD
}

func (s *Snake) Update() {
	var prev Vec2i
	for i, p := range s.body {
		// Also clean the buf at p.
		s.buf[p.y][p.x] = 0

		// Head
		if i == 0 {
			prev = p
			s.body[i] = p.Add(s.direction)
		} else {
			s.body[i] = prev
			prev = p
		}

		np := s.body[i]

		// Check boundaries.
		if np.y < 0 || np.y >= len(s.buf) || np.x < 0 || np.x >= len(s.buf[np.y]) {
			// Gameover!
			// TODO: Moar friendly gameover.
			fmt.Println("Gameover")
			os.Exit(1)
		}

		// TODO: Check self eating.

		// If snake got food.
		if np.Eq(s.food) {
			// TODO: Increase body.

			// Spawn new food.
			s.spawnFood()
		}

		s.buf[np.y][np.x] = CH_BODY
	}
}

func (s *Snake) Render() {
	fmt.Print("\u001B[G")
	for i := 0; i < len(s.buf); i++ {
		for j := 0; j < len(s.buf[i]); j++ {
			var cell string
			switch s.buf[i][j] {
			case CH_BODY:
				cell = " * "
			case CH_FOOD:
				cell = " @ "
			default:
				cell = "---"
			}

			fmt.Print(cell)
		}
		fmt.Print("\u001B[G\n")
	}
	fmt.Printf("\u001B[%dD\u001B[%dA", s.w, s.h)
}

func (s *Snake) HandleKey(k byte) {
	switch k {
	case 'h':
		// left
		if s.direction.x == 0 {
			s.direction = NewVec2i(-1, 0)
		}
	case 'j':
		// down
		if s.direction.y == 0 {
			s.direction = NewVec2i(0, 1)
		}
	case 'k':
		// up
		if s.direction.y == 0 {
			s.direction = NewVec2i(0, -1)
		}
	case 'l':
		// right
		if s.direction.x == 0 {
			s.direction = NewVec2i(1, 0)
		}
	case 'q':
		s.quit = true
	default:
		// do nothing.
	}
}

func listenKeys(s *Snake, ch chan byte, r io.Reader) {
	reader := bufio.NewReaderSize(r, 1)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			// TODO: Panic or what?
		}
		s.HandleKey(b)
		ch <- b
	}
}

func loop(c *Config) {
	rand.Seed(time.Now().Unix())

	s := NewSnake(c)

	k := make(chan byte)
	go listenKeys(s, k, os.Stdin)

	for !s.quit {
		select {
		case <-time.After(time.Second / FPS):
		case <-k:
		}

		s.Update()
		s.Render()
	}
}

func parseArgs() *Config {
	c := &Config{}

	flag.IntVar(&c.Width, "width", 10, "")
	flag.IntVar(&c.Height, "height", 10, "")
	flag.Float64Var(&c.Speed, "speed", 1.0, "")
	flag.Parse()

	return c
}

func main() {
	state, err := term.MakeRaw(0)
	if err != nil {
		log.Fatal(err)
	}
	defer term.Restore(0, state)

	loop(parseArgs())
}
