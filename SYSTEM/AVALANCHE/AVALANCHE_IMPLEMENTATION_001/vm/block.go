package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow/choices"
)

const (
	blockSchema            = "PRESENCE_AVALANCHE_BLOCK_001"
	maxTransitionsPerBlock = 64
	maxClockSkew           = 10 * time.Second
)

var errCannotRejectGenesis = errors.New("the definitionally accepted genesis block cannot be rejected")

type WireBlock struct {
	Schema      string        `json:"schema"`
	ParentID    string        `json:"parent_id"`
	Timestamp   int64         `json:"timestamp"`
	Height      uint64        `json:"height"`
	Transitions []*Transition `json:"transitions"`
}

type localityBlock struct {
	vm           *VM
	wire         *WireBlock
	id           ids.ID
	bytes        []byte
	parentID     ids.ID
	height       uint64
	timestamp    time.Time
	status       choices.Status
	postState    *RuntimeState
	receipts     []*Receipt
	genesis      bool
	builtLocally bool
}

func newGenesisBlock(vm *VM, genesisBytes []byte, state *RuntimeState) *localityBlock {
	digest := sha256.Sum256(genesisBytes)
	return &localityBlock{
		vm:        vm,
		id:        ids.ID(digest),
		bytes:     bytes.Clone(genesisBytes),
		parentID:  ids.Empty,
		height:    0,
		timestamp: time.Unix(0, 0).UTC(),
		status:    choices.Accepted,
		postState: state,
		genesis:   true,
	}
}

func parseWireBlock(blockBytes []byte) (*WireBlock, ids.ID, ids.ID, error) {
	wire := new(WireBlock)
	if err := decodeCanonical(blockBytes, wire); err != nil {
		return nil, ids.Empty, ids.Empty, err
	}
	if wire.Schema != blockSchema {
		return nil, ids.Empty, ids.Empty, fmt.Errorf("unsupported block schema %q", wire.Schema)
	}
	parentID, err := ids.FromString(wire.ParentID)
	if err != nil {
		return nil, ids.Empty, ids.Empty, fmt.Errorf("parent_id: %w", err)
	}
	if wire.Timestamp < 0 || wire.Height == 0 {
		return nil, ids.Empty, ids.Empty, errors.New("non-genesis block must have a non-negative timestamp and positive height")
	}
	if len(wire.Transitions) == 0 || len(wire.Transitions) > maxTransitionsPerBlock {
		return nil, ids.Empty, ids.Empty, fmt.Errorf("block must contain 1..%d transitions", maxTransitionsPerBlock)
	}
	seen := make(map[string]struct{}, len(wire.Transitions))
	for i, transition := range wire.Transitions {
		if transition == nil {
			return nil, ids.Empty, ids.Empty, fmt.Errorf("transition %d is nil", i)
		}
		if err := transition.ValidateSyntax(); err != nil {
			return nil, ids.Empty, ids.Empty, fmt.Errorf("transition %d: %w", i, err)
		}
		transitionID, err := transition.ID()
		if err != nil {
			return nil, ids.Empty, ids.Empty, err
		}
		idHex := hex.EncodeToString(transitionID[:])
		if _, exists := seen[idHex]; exists {
			return nil, ids.Empty, ids.Empty, fmt.Errorf("duplicate transition %s", idHex)
		}
		seen[idHex] = struct{}{}
	}
	digest := sha256.Sum256(blockBytes)
	return wire, ids.ID(digest), parentID, nil
}

func newWireBlock(vm *VM, wire *WireBlock, builtLocally bool) (*localityBlock, error) {
	blockBytes, err := json.Marshal(wire)
	if err != nil {
		return nil, err
	}
	parsed, blockID, parentID, err := parseWireBlock(blockBytes)
	if err != nil {
		return nil, err
	}
	return &localityBlock{
		vm:           vm,
		wire:         parsed,
		id:           blockID,
		bytes:        blockBytes,
		parentID:     parentID,
		height:       wire.Height,
		timestamp:    time.Unix(wire.Timestamp, 0).UTC(),
		status:       choices.Processing,
		builtLocally: builtLocally,
	}, nil
}

func blockFromBytes(vm *VM, blockBytes []byte, status choices.Status) (*localityBlock, error) {
	wire, blockID, parentID, err := parseWireBlock(blockBytes)
	if err != nil {
		return nil, err
	}
	return &localityBlock{
		vm:        vm,
		wire:      wire,
		id:        blockID,
		bytes:     bytes.Clone(blockBytes),
		parentID:  parentID,
		height:    wire.Height,
		timestamp: time.Unix(wire.Timestamp, 0).UTC(),
		status:    status,
	}, nil
}

func (b *localityBlock) ID() ids.ID           { return b.id }
func (b *localityBlock) Parent() ids.ID       { return b.parentID }
func (b *localityBlock) Height() uint64       { return b.height }
func (b *localityBlock) Timestamp() time.Time { return b.timestamp }
func (b *localityBlock) Bytes() []byte        { return bytes.Clone(b.bytes) }
func (b *localityBlock) Status() choices.Status {
	b.vm.lock.RLock()
	defer b.vm.lock.RUnlock()
	return b.status
}

func (b *localityBlock) Verify(context.Context) error {
	if b.genesis {
		return nil
	}
	return b.vm.verifyBlock(b)
}

func (b *localityBlock) Accept(context.Context) error {
	if b.genesis {
		return nil
	}
	return b.vm.acceptBlock(b)
}

func (b *localityBlock) Reject(context.Context) error {
	if b.genesis {
		return errCannotRejectGenesis
	}
	return b.vm.rejectBlock(b)
}
