package gol

import (
	"fmt"
	"log"
	"net/rpc"
	"strconv"
	"strings"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

const (
	alive = 255
)

func distributor(p Params, c distributorChannels) {
	c.ioCommand <- ioInput
	filename := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)
	c.ioFilename <- filename
	closed := false
	// TODO: Create a 2D slice to store the world.
	client, err := rpc.Dial("tcp", "0.0.0.0:8030")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer client.Close()
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
		for j := range world[i] {

			world[i][j] = <-c.ioInput
			if world[i][j] == alive {
				c.events <- CellFlipped{
					Cell:           util.Cell{X: j, Y: i},
					CompletedTurns: 0,
				}
			}
		}
	}
	go timer(client, c.events, &closed)
	var res stubs.Response
	req := stubs.Request{
		World: world,
		Params: stubs.Params{
			ImageWidth:  p.ImageWidth,
			ImageHeight: p.ImageHeight,
			Turns:       p.Turns,
			Threads:     p.Threads,
		},
	}

	err = client.Call(stubs.NextStep, req, &res)
	if err != nil {
		panic(err)
	}
	world = res.World
	turn := res.Turns
	Output(p, c, world, turn)
	c.events <- FinalTurnComplete{p.Turns, res.AliveCells}
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}
	closed = true
	close(c.events)
}

/* func nextState(p Params, world [][]byte, turn int, c distributorChannels) [][]byte {
	threads := p.Threads
	var newWorld [][]byte
	out := make([]chan [][]byte, threads)
	for i := 0; i < threads; i++ {
		out[i] = make(chan [][]byte)
	}
	if threads == 1 {
		newWorld = RuleOfGame(0, p.ImageHeight, 0, p.ImageWidth, world, p, c, turn)
	} else {

		workerHeight := p.ImageHeight / p.Threads
		for i := 0; i < threads-1; i++ {
			go worker(i*workerHeight, (i+1)*workerHeight, 0, p.ImageWidth, world, out[i], p, c, turn)
		}

		go worker(workerHeight*(p.Threads-1), p.ImageHeight, 0, p.ImageWidth, world, out[p.Threads-1], p, c, turn)

		newWorld = nil
		for i := 0; i < threads; i++ {
			part := <-out[i]
			newWorld = append(newWorld, part...)
		}
	}
	return newWorld
}

func worker(startY, endY, startX, endX int, world [][]byte, out chan<- [][]byte, p Params, c distributorChannels, turn int) {
	pw := RuleOfGame(startY, endY, startX, endX, world, p, c, turn)
	out <- pw
}

func RuleOfGame(startY, endY, startX, endX int, world [][]byte, p Params, c distributorChannels, turn int) [][]byte {
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
					c.events <- CellFlipped{turn, cell}
				} else if neibours == 2 || neibours == 3 {
					nextworld[wy][x] = alive
				} else {
					nextworld[wy][x] = dead
					c.events <- CellFlipped{turn, cell}
				}
			} else {
				if neibours == 3 {
					nextworld[wy][x] = alive
					c.events <- CellFlipped{turn, cell}
				} else {
					nextworld[wy][x] = dead
				}
			}
		}
	}
	return nextworld
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
} */

/*func Timeticker(world [][]byte, ticker <-chan time.Time, c distributorChannels, turn int) {
	select {
	case <-ticker:
		Cells := len(AliveCells(world))
		c.events <- AliveCellsCount{
			CellsCount:     Cells,
			CompletedTurns: turn,
		}
	default:
		return
	}
}*/

func Output(p Params, c distributorChannels, world [][]byte, turn int) {
	c.ioCommand <- ioOutput
	pgmFilename := strings.Join([]string{strconv.Itoa(p.ImageWidth), strconv.Itoa(p.ImageHeight), strconv.Itoa(turn)}, "x")
	c.ioFilename <- pgmFilename
	for y := range world { //send world via output channel byte by byte
		for x := range world[y] {
			c.ioOutput <- world[y][x]
		}
	}
	c.events <- ImageOutputComplete{turn, pgmFilename}
}

/*func keyPress(turn *int, world [][]byte, p Params, c distributorChannels) {
	var Locker sync.Mutex
	for {
		key := <-c.keyPresses
		switch key {
		case 's':
			Output(p, c, world, *turn)

		case 'q':

			Output(p, c, world, *turn)
			aliveCells := AliveCells(world)
			c.events <- FinalTurnComplete{p.Turns, aliveCells}
			c.ioCommand <- ioCheckIdle
			<-c.ioIdle
			c.events <- StateChange{*turn, Quitting}
			os.Exit(0)

		case 'p':
			Locker.Lock()
			c.events <- StateChange{*turn, Paused}

			for {
				Key := <-c.keyPresses
				if Key == 'p' {
					break
				}
			}
			Locker.Unlock()
			c.events <- StateChange{*turn, Executing}
		}
	}
}*/

/*func AliveCells(world [][]byte) []util.Cell {
	Cells := make([]util.Cell, 0)
	for y := range world {
		for x := range world[y] {
			if world[y][x] == alive {
				Cells = append(Cells, util.Cell{X: x, Y: y})
			}
		}
	}
	return Cells
}*/

func timer(client *rpc.Client, eventChan chan<- Event, Closed *bool) {
	for {
		time.Sleep(time.Second * 2)
		if !*Closed {
			var res stubs.CellsResponse
			err := client.Call(stubs.GetAliveCells, stubs.CellsRequest{}, &res)
			if err != nil {
				log.Printf("Error: %v", err)
			}
			eventChan <- AliveCellsCount{CellsCount: res.AliveCellsCount, CompletedTurns: res.Turn}
		}
	}
}
