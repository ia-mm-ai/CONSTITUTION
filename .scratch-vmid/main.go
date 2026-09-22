package main

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/version"
)

func vmID(name string) (ids.ID, error) {
	if len(name) > 32 {
		return ids.Empty, fmt.Errorf("too long")
	}
	b := make([]byte, 32)
	copy(b, []byte(name))
	return ids.ToID(b)
}

func main() {
	id, err := vmID("PRESENCE_AVALANCHE_VM_001")
	fmt.Println(id, err)
	fmt.Println("rpcchainvm protocol:", version.RPCChainVMProtocol)
	old, _ := vmID("LOCALITY_VM_003")
	fmt.Println("padded LOCALITY_VM_003:", old)
}
