// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package evmcore

import (
	"sync/atomic"

	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	"github.com/making-choice-personal/volary-lachesis/utils/gsignercache"
)

// statePrefetcher is a basic Prefetcher, which blindly executes a block on top
// of an arbitrary state with the goal of prefetching potentially useful state
// data from disk before the main block processor start executing.
type statePrefetcher struct {
	config *params.ChainConfig // Chain configuration options
	bc     DummyChain          // Canonical block chain
}

// newStatePrefetcher initialises a new statePrefetcher.
func newStatePrefetcher(config *params.ChainConfig, bc DummyChain) *statePrefetcher {
	return &statePrefetcher{
		config: config,
		bc:     bc,
	}
}

// Prefetch processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb, but any changes are discarded. The
// only goal is to pre-cache transaction signatures and state trie nodes.
func (p *statePrefetcher) Prefetch(block *EvmBlock, statedb *state.StateDB, cfg vm.Config, interrupt *uint32) {
	var (
		header       = block.Header()
		gaspool      = new(GasPool).AddGas(block.GasLimit)
		blockContext = NewEVMBlockContext(header, p.bc, nil)
		evm          = vm.NewEVM(blockContext, vm.TxContext{}, statedb, p.config, cfg)
		signer       = gsignercache.Wrap(types.MakeSigner(p.config, header.Number))
	)
	// Iterate over and process the individual transactions
	byzantium := p.config.IsByzantium(block.Number)
	for i, tx := range block.Transactions {
		// If block precaching was interrupted, abort
		if interrupt != nil && atomic.LoadUint32(interrupt) == 1 {
			return
		}
		// Convert the transaction into an executable message and pre-cache its sender
		msg, err := tx.AsMessage(signer, header.BaseFee)
		if err != nil {
			return // Also invalid block, bail out
		}
		statedb.Prepare(tx.Hash(), i)
		if err := precacheTransaction(msg, p.config, gaspool, statedb, header, evm); err != nil {
			return // Ugh, something went horribly wrong, bail out
		}
		// If we're pre-byzantium, pre-load trie nodes for the intermediate root
		if !byzantium {
			statedb.IntermediateRoot(true)
		}
	}
	// If were post-byzantium, pre-load trie nodes for the final root hash
	if byzantium {
		statedb.IntermediateRoot(true)
	}
}

// precacheTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. The goal is not to execute
// the transaction successfully, rather to warm up touched data slots.
func precacheTransaction(msg types.Message, config *params.ChainConfig, gaspool *GasPool, statedb *state.StateDB, header *EvmHeader, evm *vm.EVM) error {
	// Update the evm with the new transaction context.
	evm.Reset(NewEVMTxContext(msg), statedb)
	// Add addresses to access list if applicable
	_, err := ApplyMessage(evm, msg, gaspool)
	return err
}
