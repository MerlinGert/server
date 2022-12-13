package stubs

import "uk.ac.bris.cs/gameoflife/util"

var NextStep = "Client.NextStep"
var GetAliveCells = "Client.GetAliveCells"

type Params struct {
	ImageWidth  int
	ImageHeight int
	Turns       int
	Threads     int
}
type Request struct {
	World  [][]uint8
	Params Params
}

type Response struct {
	World      [][]uint8
	Turns      int
	AliveCells []util.Cell
}

type CellsRequest struct {
}

type CellsResponse struct {
	Turn            int
	AliveCellsCount int
}
