package rpc

type RpcCallData struct {
	Command string                 `json:"cmd"`
	Params  map[string]interface{} `json:"params"`
}

type RpcCallSimpleReply struct {
	Success bool `json:"success"`
}
