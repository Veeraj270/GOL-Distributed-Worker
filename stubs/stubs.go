package stubs

var RemoteCall string = "RemoteProcessor.CallRemoteDistributor"
var RemoteCellCount string = "RemoteProcessor.CallNumberOfAliveCells"

var RemotePause string = "RemoteProcessor.CallPause"

var RemoteSave string = "RemoteProcessor.CallSave"
var RemoteClose string = "RemoteProcessor.CallClose"

var WorkerCalculate string = "RemoteWorker.CalculateNextState"
var WorkerTest string = "RemoteWorker.Test"
var WorkerClose string = "RemoteWorker.Close"

type Request struct {
	World   [][]uint8
	Turns   int
	Threads int
}

type Response struct {
	World [][]uint8
}

type CellCountRequest struct {
}

type CellCountResponse struct {
	Turn      int
	CellCount int
}

type PauseReq struct {
	Paused bool
}

type PauseResp struct {
	Turn int
}

type SaveReq struct {
}

type SaveResp struct {
	World [][]uint8
	Turn  int
}

type CloseReq struct{}

type CloseResp struct{}

type WorkerRequest struct {
	WorldCopy                   [][]uint8
	StartY, EndY, Height, Width int
}

type WorkerResponse struct {
	World [][]uint8
}
