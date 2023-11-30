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
var initialised bool

var toGiveContacted []uint8

var receivedFromContacted []uint8

var toGiveContacter []uint8

var receivedFromContacter []uint8

var lower *rpc.Client
var err error

var readyForContact bool

var contacted bool

func createWorldCopy(world [][]uint8) [][]uint8 {
	worldCopy := make([][]uint8, len(world))
	for i := range worldCopy {
		worldCopy[i] = make([]uint8, len(world[i]))
		copy(worldCopy[i], world[i])
	}
	return worldCopy
}

func parallelCalculateNextState(world [][]uint8, startY, endY, height, width int) [][]uint8 {
	worldCopy := createWorldCopy(world)
	//fmt.Println("Next State Calculating!")
	//fmt.Println("--------NextStateCalculating------------")
	//fmt.Println("------------------------------------------", endY-startY, "--------------------------------------------------")
	//fmt.Println("Width:", width)

	//fmt.Println("len(worldCopy[0]):", len(worldCopy[0]))
	//fmt.Println("len(worldCopy):", len(worldCopy))

	worldSection := make([][]uint8, endY-startY)
	for i := 0; i < (endY - startY); i++ {
		worldSection[i] = make([]uint8, width)
	}

	for j := startY; j < endY; j++ {
		for i := 0; i < width; i++ {
			//fmt.Println("j:", j, " i:", i)
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

			//fmt.Println("bottom:", bottom, "top:", top, "left:", left, "right:", right)
			//fmt.Println("len(worldCopy[0]):", len(worldCopy[0]))
			//fmt.Println("len(worldCopy):", len(worldCopy))

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

func processChunk(world [][]uint8, threads int, startY int, endY int, turns int) [][]uint8 {

	if initialised == false {
		lower, err = rpc.Dial("tcp", "localhost:8070")
		if err != nil {
			fmt.Println("Couldn't dial worker")

		}
		initialised = true
	}

	worldCopy := createWorldCopy(world)
	height := endY - startY
	chunkSize := height / threads
	remainingChunk := height % threads
	fmt.Println("len(world):", len(world))
	fmt.Println("len(world[0])", len(world[0]))
	fmt.Printf("endY:%d, startY:%d, height:%d, chunksize:%d\n", endY, startY, height, chunkSize)
	var bufferedSliceChan = make([]chan [][]uint8, threads)

	//fmt.Println("------------------------Turn:", i, "--------------------------")
	fmt.Println(len(world), "x", len(world[0]))

	var parallelWorld [][]uint8
	for k := 0; k < threads; k++ {
		if k < threads-remainingChunk {
			Begin := (k * chunkSize) + startY
			End := ((k + 1) * chunkSize) + startY
			fmt.Println("Begin: ", Begin, " End: ", End)
			bufferedSliceChan[k] = make(chan [][]uint8)
			go func(worldCopy [][]uint8, StartY int, EndY int, out chan [][]uint8) {
				out <- parallelCalculateNextState(worldCopy, Begin, End, len(worldCopy), len(worldCopy[0]))
			}(worldCopy, Begin, End, bufferedSliceChan[k])
		} else if k == threads-remainingChunk {
			Begin := (k * chunkSize) + startY
			End := ((k+1)*chunkSize + 1) + startY
			fmt.Println("Begin: ", Begin, " End: ", End)

			bufferedSliceChan[k] = make(chan [][]uint8)
			go func(worldCopy [][]uint8, StartY int, EndY int, out chan [][]uint8) {
				out <- parallelCalculateNextState(worldCopy, Begin, End, len(worldCopy), len(worldCopy[0]))
			}(worldCopy, Begin, End, bufferedSliceChan[k])
		} else if k > threads-remainingChunk {
			Begin := ((k * chunkSize) + (k - (threads - remainingChunk))) + startY
			End := ((k+1)*chunkSize + (k + 1 - (threads - remainingChunk))) + startY
			fmt.Println("Begin: ", Begin, " End: ", End)

			bufferedSliceChan[k] = make(chan [][]uint8)
			go func(worldCopy [][]uint8, StartY int, EndY int, out chan [][]uint8) {
				out <- parallelCalculateNextState(worldCopy, Begin, End, len(worldCopy), len(worldCopy[0]))
			}(worldCopy, Begin, End, bufferedSliceChan[k])
		}
	}

	fmt.Println("Go routines deployed")
	for i := 0; i < threads; i++ {
		parallelWorld = append(parallelWorld, <-bufferedSliceChan[i]...)
	}
	fmt.Println("Go routines reassembled")
	worldCopy = parallelWorld
	world = parallelWorld

	return world
}

type RemoteWorker struct{}

func (r *RemoteWorker) HaloExchange(request stubs.HaloRequest, response *stubs.HaloResponse) (err error) {
	fmt.Println("for loop escaped")

	for !readyForContact {
	}
	readyForContact = false
	receivedFromContacter = request.Halo
	response.Halo = toGiveContacter
	contacted = true //after exchange
	return
}

func (r *RemoteWorker) CalculateNextState(request stubs.WorkerRequest, response *stubs.WorkerResponse) (err error) {
	//fmt.Println("Rpc call received!")
	done = false
	//response.World = parallelCalculateNextState(request.WorldCopy, request.StartY, request.EndY, request.Height, request.Width)
	response.World = processChunk(request.WorldCopy, 16, request.StartY, request.EndY, request.Turns)
	fmt.Println("Response made")
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
	pAddr := flag.String("port", ":8030", "Port to listen on")
	//pAddr2 := flag.String("top", ":8070", "Port to receive top halo")
	flag.Parse()

	contacted = false
	readyForContact = false

	listener, _ := net.Listen("tcp", *pAddr)

	/*
		defer func(listener net.Listener) {
			err := listener.Close()
			if err != nil {
				fmt.Println("Error closing the listener")
			}
		}(listener)

	*/
	err := rpc.Register(&RemoteWorker{})
	if err != nil {
		fmt.Println("Error registering rpc")
	}

	rpc.Accept(listener)

	fmt.Println("Connection accepted")

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("Error closing the listener2")
		}
	}(listener)

}
