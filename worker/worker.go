package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"uk.ac.bris.cs/gameoflife/stubs"
)

var done bool

func parallelCalculateNextState(worldCopy [][]uint8, startY, endY, height, width int) [][]uint8 {
	fmt.Println("Next State Calculating!")
	//fmt.Println("--------NextStateCalculating------------")
	worldSection := make([][]uint8, endY-startY)
	for i := 0; i < (endY - startY); i++ {
		worldSection[i] = make([]uint8, width)
	}

	for j := startY; j < endY; j++ {
		for i := 0; i < width; i++ {
			sum := 0
			var neighbours [8]uint8
			top := j - 1
			bottom := j + 1
			left := i - 1
			right := i + 1
			if top == -1 {
				top = height - 1
			}
			if bottom == height {
				bottom = 0
			}
			if left == -1 {
				left = width - 1
			}
			if right == width {
				right = 0
			}
			neighbours[0] = worldCopy[bottom][left]
			neighbours[1] = worldCopy[bottom][i]
			neighbours[2] = worldCopy[bottom][right]
			neighbours[3] = worldCopy[j][left]
			neighbours[4] = worldCopy[j][right]
			neighbours[5] = worldCopy[top][left]
			neighbours[6] = worldCopy[top][i]
			neighbours[7] = worldCopy[top][right]

			for _, n := range neighbours {
				if n == 255 {
					sum = sum + 1
				}
			}

			if worldCopy[j][i] == 255 {
				if sum < 2 {
					worldSection[j-startY][i] = 0
					//c.events <- CellFlipped{CompletedTurns: *turns, Cell: util.Cell{X: i, Y: j}}
				} else if (sum == 2) || (sum == 3) {
					worldSection[j-startY][i] = 255
				} else if sum > 3 {
					worldSection[j-startY][i] = 0
					//c.events <- CellFlipped{CompletedTurns: *turns, Cell: util.Cell{X: i, Y: j}}
				}
			} else if worldCopy[j][i] == 0 {
				if sum == 3 {
					worldSection[j-startY][i] = 255
					//c.events <- CellFlipped{CompletedTurns: *turns, Cell: util.Cell{X: i, Y: j}}
				} else {
					worldSection[j-startY][i] = 0
				}
			}
		}
	}
	return worldSection
}

type RemoteWorker struct{}

func (r *RemoteWorker) CalculateNextState(request stubs.WorkerRequest, response *stubs.WorkerResponse) (err error) {
	//fmt.Println("Rpc call received!")
	done = false
	response.World = parallelCalculateNextState(request.WorldCopy, request.StartY, request.EndY, request.Height, request.Width)
	done = true
	return
}

func (r *RemoteWorker) Close(request stubs.CloseReq, response *stubs.CloseResp) (err error) {
	for !done {
	}
	fmt.Println(done)
	os.Exit(0)
	return
}

func (r *RemoteWorker) Test(request stubs.Request, response *stubs.Response) (err error) {
	fmt.Println("Test worked")
	return
}

func main() {
	pAddr := flag.String("port", ":8040", "Port to listen on")
	flag.Parse()

	listener, _ := net.Listen("tcp", *pAddr)
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("Error closing the listener")
		}
	}(listener)
	err := rpc.Register(&RemoteWorker{})
	if err != nil {
		fmt.Println("Error registering rpc")
	}
	rpc.Accept(listener)
	fmt.Println("Connection accepted")

}
