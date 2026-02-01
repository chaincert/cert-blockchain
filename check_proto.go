package main

import (
    "fmt"
    evmapi "github.com/evmos/evmos/v20/api/ethermint/evm/v1"
    "google.golang.org/protobuf/proto"
)

func main() {
    msg := &evmapi.MsgEthereumTx{}
    // Assert it implements proto.Message
    var _ proto.Message = msg
    
    fmt.Printf("FullName: %s\n", msg.ProtoReflect().Descriptor().FullName())
}
