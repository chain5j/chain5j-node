// Package node
//
// @author: xwc1125
package node

import (
	"fmt"

	"github.com/chain5j/chain5j-pkg/codec"
	"github.com/chain5j/chain5j-protocol/protocol"
)

type option func(opt *node) error

func apply(f *node, opts ...option) error {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(f); err != nil {
			return fmt.Errorf("option apply err:%v", err)
		}
	}
	return nil
}

func WithCodec(codec codec.Codec) option {
	return func(opt *node) error {
		opt.codec = codec
		return nil
	}
}
func WithConsensus(consensus protocol.Consensus) option {
	return func(opt *node) error {
		opt.consensus = consensus
		return nil
	}
}
func WithPacker(packer protocol.Packer) option {
	return func(opt *node) error {
		opt.packer = packer
		return nil
	}
}
func WithDatabase(database protocol.Database) option {
	return func(opt *node) error {
		opt.database = database
		return nil
	}
}
func WithBlockReader(blockReader protocol.BlockReader) option {
	return func(opt *node) error {
		opt.blockReader = blockReader
		return nil
	}
}
func WithBlockWriter(blockWriter protocol.BlockWriter) option {
	return func(opt *node) error {
		opt.blockWriter = blockWriter
		return nil
	}
}
func WithBlockReadWriter(blockReadWriter protocol.BlockReadWriter) option {
	return func(opt *node) error {
		opt.blockReadWriter = blockReadWriter
		return nil
	}
}
func WithNodeKey(nodeKey protocol.NodeKey) option {
	return func(opt *node) error {
		opt.nodeKey = nodeKey
		return nil
	}
}
func WithApps(apps protocol.Apps) option {
	return func(opt *node) error {
		opt.apps = apps
		return nil
	}
}
func WithP2PService(p2pService protocol.P2PService) option {
	return func(opt *node) error {
		opt.p2pService = p2pService
		return nil
	}
}
func WithBroadcaster(broadcaster protocol.Broadcaster) option {
	return func(opt *node) error {
		opt.broadcaster = broadcaster
		return nil
	}
}
func WithTxPools(txPools protocol.TxPools) option {
	return func(opt *node) error {
		opt.txPools = txPools
		return nil
	}
}
func WithSyncer(syncer protocol.Syncer) option {
	return func(opt *node) error {
		opt.syncer = syncer
		return nil
	}
}
func WithPermission(permission protocol.Permission) option {
	return func(opt *node) error {
		opt.permission = permission
		return nil
	}
}
