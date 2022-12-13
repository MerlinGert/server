package main

import (
	"flag"
	"log"
	"net"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

const (
	dead  = 0
	alive = 255
)

type Client struct {
	world [][]uint8
	Turn  int
	Param stubs.Params
}

func (c *Client) NextStep(req stubs.Request, res *stubs.Response) error {
	c.Param = req.Params
	c.world = req.World
	c.Turn = 0
	for c.Turn < req.Params.Turns {

		newWorld := nextState(req.Params, c.world)
		c.Turn++
		c.world = newWorld

	}
	res.World = c.world
	res.Turns = c.Turn
	res.AliveCells = AliveCells(c.world)
	return nil
}

func (c *Client) GetAliveCells(req stubs.CellsRequest, res *stubs.CellsResponse) error {
	res.Turn = c.Turn
	res.AliveCellsCount = len(AliveCells(c.world))
	return nil
}

func RuleOfGame(startY, endY, startX, endX int, world [][]byte, p stubs.Params) [][]byte {
	h := endY - startY
	w := endX - startX
	ph := p.ImageHeight
	nextworld := make([][]byte, h)
	for i := range nextworld {
		nextworld[i] = make([]byte, w)
	}
	for y := startY; y < endY; y++ {
		for x := 0; x < w; x++ {
			var cell util.Cell
			cell.Y = y
			cell.X = x
			XR, XL := x+1, x-1
			YU, YD := y+1, y-1
			wy := y - startY
			neibours := (world[(YD+ph)%(ph)][(XL+w)%w] / alive) +
				(world[(YD+ph)%(ph)][(x+w)%w] / alive) +
				(world[(YD+ph)%(ph)][(XR+w)%w] / alive) +
				(world[(y+ph)%(ph)][(XL+w)%w] / alive) +
				(world[(y+ph)%(ph)][(XR+w)%w] / alive) +
				(world[(YU+ph)%(ph)][(XL+w)%w] / alive) +
				(world[(YU+ph)%(ph)][(x+w)%w] / alive) +
				(world[(YU+ph)%(ph)][(XR+w)%w] / alive)
			//update the matrix
			if world[y][x] == alive {
				if neibours < 2 || neibours > 3 {
					nextworld[wy][x] = dead
				} else if neibours == 2 || neibours == 3 {
					nextworld[wy][x] = alive
				} else {
					nextworld[wy][x] = dead
				}
			} else {
				if neibours == 3 {
					nextworld[wy][x] = alive
				} else {
					nextworld[wy][x] = dead
				}
			}
		}
	}
	return nextworld
}
func worker(startY, endY, startX, endX int, world [][]byte, out chan<- [][]byte, p stubs.Params) {
	pw := RuleOfGame(startY, endY, startX, endX, world, p)
	out <- pw
}

func nextState(p stubs.Params, world [][]byte) [][]byte {
	threads := p.Threads
	var newWorld [][]byte
	out := make([]chan [][]byte, threads)
	for i := 0; i < threads; i++ {
		out[i] = make(chan [][]byte)
	}
	if threads == 1 {
		newWorld = RuleOfGame(0, p.ImageHeight, 0, p.ImageWidth, world, p)
	} else {

		workerHeight := p.ImageHeight / p.Threads
		for i := 0; i < threads-1; i++ {
			go worker(i*workerHeight, (i+1)*workerHeight, 0, p.ImageWidth, world, out[i], p)
		}
		go worker(workerHeight*(p.Threads-1), p.ImageHeight, 0, p.ImageWidth, world, out[p.Threads-1], p)
		newWorld = nil
		for i := 0; i < threads; i++ {
			part := <-out[i]
			newWorld = append(newWorld, part...)
		}
	}
	return newWorld
}

func main() {
	port := flag.String("port", "8030", "port to listen on")
	flag.Parse()

	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()
	log.Printf("Listening on port %s", *port)
	rpc.Register(new(Client))
	rpc.Accept(listener)
}

func AliveCells(world [][]byte) []util.Cell {
	Cells := make([]util.Cell, 0)
	for y := range world {
		for x := range world[y] {
			if world[y][x] == alive {
				Cells = append(Cells, util.Cell{X: x, Y: y})
			}
		}
	}
	return Cells
}
