// Package node
//
// @author: xwc1125
package node

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"

	apis "github.com/chain5j/chain5j-apis"
	blockchain "github.com/chain5j/chain5j-blockchain"
	broadcaster "github.com/chain5j/chain5j-broadcaster"
	config "github.com/chain5j/chain5j-config"
	kvstore "github.com/chain5j/chain5j-kvstore"
	nodekey "github.com/chain5j/chain5j-nodekey"
	p2p "github.com/chain5j/chain5j-p2p"
	packer "github.com/chain5j/chain5j-packer"
	pbft "github.com/chain5j/chain5j-pbft"
	permission "github.com/chain5j/chain5j-permission"
	"github.com/chain5j/chain5j-pkg/cli"
	"github.com/chain5j/chain5j-pkg/codec"
	kvstore2 "github.com/chain5j/chain5j-pkg/database/kvstore"
	"github.com/chain5j/chain5j-pkg/database/kvstore/leveldb"
	"github.com/chain5j/chain5j-pkg/types"
	"github.com/chain5j/chain5j-protocol/dispatch"
	"github.com/chain5j/chain5j-protocol/models"
	"github.com/chain5j/chain5j-protocol/protocol"
	syncer "github.com/chain5j/chain5j-syncer"
	"github.com/chain5j/logger"
	"github.com/fsnotify/fsnotify"
)

var (
	_ protocol.Node = new(node)
)

// node 节点组装器
type node struct {
	log    logger.Logger
	ctx    context.Context
	cancel context.CancelFunc

	configFile string
	codec      codec.Codec

	quiteCh chan struct{} // 用于wait
	lock    sync.RWMutex

	stateDataDB kvstore2.Database
	blockDataDB kvstore2.Database
	crudDataDB  kvstore2.Database

	config          protocol.Config
	database        protocol.Database
	blockReader     protocol.BlockReader
	blockWriter     protocol.BlockWriter
	blockReadWriter protocol.BlockReadWriter
	nodeKey         protocol.NodeKey
	apps            protocol.Apps
	p2pService      protocol.P2PService
	broadcaster     protocol.Broadcaster
	consensus       protocol.Consensus
	txPools         protocol.TxPools
	handshake       protocol.Handshake
	packer          protocol.Packer
	syncer          protocol.Syncer
	apis            protocol.APIs
	permission      protocol.Permission
}

// NewNode 创建节点
func NewNode(rootCtx context.Context, configFile string, initDB bool, opts ...option) (protocol.Node, error) {
	cpuNum := runtime.NumCPU() // 获得当前设备的cpu核心数
	log.Printf("node get CPU num: cpuNum=%d", cpuNum)
	runtime.GOMAXPROCS(cpuNum) // 设置需要用到的cpu数量

	if len(configFile) == 0 {
		return nil, fmt.Errorf("config file is nil")
	}
	ctx, cancel := context.WithCancel(rootCtx)
	node := &node{
		log:        logger.New("node"),
		configFile: configFile,
		ctx:        ctx,
		cancel:     cancel,

		quiteCh: make(chan struct{}),
	}
	if err := apply(node, opts...); err != nil {
		node.log.Error("apply is error", "err", err)
		return nil, err
	}
	if initDB {
		if err := node.initModulesWithDB(); err != nil {
			return nil, err
		}
	} else {
		if err := node.initModulesNoDB(); err != nil {
			return nil, err
		}
	}

	return node, nil
}

func (n *node) initModulesNoDB() error {
	var err error
	// config
	{
		if n.config == nil {
			n.config, err = config.NewConfig(n.configFile)
			if err != nil {
				return err
			}
		}
	}

	// 	nodeKey 需要先从指定的路径中读取，如果没有生成后，写入指定位置
	{
		if n.nodeKey == nil {
			n.nodeKey, err = nodekey.NewNodeKey(n.ctx,
				nodekey.WithConfig(n.config),
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (n *node) initModulesWithDB() error {
	var err error
	// database
	{
		var dbConfig = new(models.DatabaseConfig)
		err := cli.LoadConfig(n.configFile, "database", dbConfig, func(e fsnotify.Event) {})
		if err != nil {
			return err
		}

		switch dbConfig.Driver {
		case "leveldb":
			n.stateDataDB, err = leveldb.New(filepath.Join(dbConfig.Source, "statedata"), 0, 0, "")
			if err != nil {
				n.Stop()
				return err
			}
			n.blockDataDB, err = leveldb.New(filepath.Join(dbConfig.Source, "blockdata"), 0, 0, "")
			if err != nil {
				n.Stop()
				return err
			}

			n.crudDataDB, err = leveldb.New(filepath.Join(dbConfig.Source, "cruddata"), 0, 0, "")
			if err != nil {
				n.Stop()
				return err
			}
			n.database, err = kvstore.NewKvStore(n.ctx,
				kvstore.WithDB(n.stateDataDB))
			if err != nil {
				n.Stop()
				return err
			}
		default:
			return fmt.Errorf("unsupported the db driver: %s", dbConfig.Driver)
		}
	}

	// config
	{
		if n.config == nil {
			n.config, err = config.NewConfig(n.configFile,
				config.WithDB(n.database))
			if err != nil {
				return err
			}
		}
	}

	// 	nodeKey 需要先从指定的路径中读取，如果没有生成后，写入指定位置
	{
		if n.nodeKey == nil {
			n.nodeKey, err = nodekey.NewNodeKey(n.ctx,
				nodekey.WithConfig(n.config),
			)
			if err != nil {
				return err
			}
		}
	}

	// apps
	{
		if n.apps == nil {
			n.apps = dispatch.NewApps(n.nodeKey)
		}
	}

	// apis

	// apis
	{
		n.apis, err = apis.NewApis(
			n.ctx,
			apis.WithBlockDB(n.database),
			apis.WithStateDB(n.database),
		)
		if err != nil {
			return err
		}
	}

	// p2p
	{
		if n.p2pService == nil {
			n.p2pService, err = p2p.NewP2P(
				n.ctx,
				p2p.WithConfig(n.config),
				p2p.WithAPIs(n.apis),
			)
			if err != nil {
				return err
			}
		}
	}

	// broadcast
	{
		if n.broadcaster == nil {
			n.broadcaster, err = broadcaster.NewBroadcaster(n.ctx,
				broadcaster.WithP2PService(n.p2pService),
				broadcaster.WithConfig(n.config),
			)
			if err != nil {
				return err
			}
		}
	}

	// blockReader
	{
		// 需要在consensus前
		if n.blockReader == nil {
			n.blockReader, err = blockchain.NewBlockReader(n.config, n.database, n.apis)
			if err != nil {
				n.Stop()
				return err
			}
		}
	}

	// txpool
	{
		if n.txPools == nil {
			n.txPools, err = dispatch.NewTxPools(n.ctx,
				dispatch.WithConfig(n.config),
				dispatch.WithBlockReader(n.blockReader),
				dispatch.WithBroadcaster(n.broadcaster),
				dispatch.WithApps(n.apps),
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *node) Init() (err error) {
	// consensus
	{
		if n.consensus == nil {
			if n.config.ChainConfig().Consensus != nil {
				if n.config.ChainConfig().Consensus.Name != "pbft" {
					return fmt.Errorf("pbft config is empty")
				}

				n.consensus, err = pbft.NewConsensus(context.Background(),
					pbft.WithConfig(n.config),
					pbft.WithBlockReader(n.blockReader),
					pbft.WithDatabaseReader(n.database),
					pbft.WithBroadcaster(n.broadcaster),
					pbft.WithNodeKey(n.nodeKey),
					pbft.WithKVDB(n.stateDataDB),
				)
				if err != nil {
					return err
				}
			}
		}
	}

	// blockchain
	{
		if n.blockWriter == nil {
			n.blockWriter = blockchain.NewBlockWriter(
				n.blockReader,
				n.consensus, // 使用共识验证区块头
				n.apps,
			)
		}
		if n.blockReadWriter == nil {
			n.blockReadWriter, err = blockchain.NewBlockRW(
				n.blockReader,
				n.blockWriter,
			)
			if err != nil {
				return err
			}
		}
	}

	// handshake
	{
		n.handshake, err = permission.NewProtocolManager(n.ctx,
			0,
			permission.WithP2PService(n.p2pService),
			permission.WithBlockRW(n.blockReadWriter),
			permission.WithBroadcaster(n.broadcaster),
		)
	}

	// syncer
	{
		if n.syncer == nil {
			n.syncer, err = syncer.NewSyncer(n.ctx,
				syncer.WithBlockRW(n.blockReadWriter),
				syncer.WithApps(n.apps),
				syncer.WithP2PService(n.p2pService),
				syncer.WithHandshake(n.handshake),
			)
			if err != nil {
				return err
			}
		}
	}

	// packer
	{
		if n.packer == nil {
			n.packer, err = packer.NewPacker(n.ctx,
				packer.WithConfig(n.config),
				packer.WithBlockRW(n.blockReadWriter),
				packer.WithTxPools(n.txPools),
				packer.WithApps(n.apps),
				packer.WithEngine(n.consensus),
			)
			if err != nil {
				return err
			}
		}
	}

	apis.WithTxPools(n.apis, n.txPools)
	apis.WithBlockReader(n.apis, n.blockReader)
	return nil
}

// Start 启动
func (n *node) Start() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	// 注册编码
	if n.codec != nil {
		codec.RegisterCodec(n.codec)
	}
	// 启动database
	n.log.Debug("start database ...")
	if err := n.blockReader.Start(); err != nil {
		n.log.Error("start database err", "err", err)
		return err
	}
	// 启动p2p
	n.log.Debug("start p2pService ...")
	if err := n.p2pService.Start(); err != nil {
		n.log.Error("start p2p service err", "err", err)
		return err
	}

	// 启动txPool
	n.log.Debug("start txPool ...")
	if err := n.txPools.Start(); err != nil {
		n.log.Error("start txPool err", "err", err)
		return err
	}

	// 启动握手
	n.log.Debug("start handshake ...")
	if err := n.handshake.Start(); err != nil {
		n.log.Error("start handshake err", "err", err)
		return err
	}

	// 启动共识
	n.log.Debug("start engine ...")
	if err := n.consensus.Start(); err != nil {
		n.log.Error("start engine err", "err", err)
		return err
	}

	// 启动同步
	n.log.Debug("start syncer ...")
	if err := n.syncer.Start(); err != nil {
		n.log.Error("start syncer err", "err", err)
		return err
	}

	// 启动区块打包功能。【启动命令行中--mine】
	if n.config.EnablePacker() {
		n.log.Debug("start packer ...")
		if err := n.packer.Start(); err != nil {
			n.log.Error("start packer err", "err", err)
			return err
		}
	}

	go n.handleSig()

	return nil
}

func (n *node) handleSig() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigc)

	<-sigc
	n.log.Info("Got interrupt, shutting down...")
	go n.Stop()
}

// Stop 停止
func (n *node) Stop() error {
	n.lock.Lock()
	defer n.lock.Unlock()

	n.cancel()

	if n.packer != nil {
		n.packer.Stop()
	}

	if n.consensus != nil {
		n.consensus.Stop()
	}
	if n.syncer != nil {
		n.syncer.Stop()
	}
	// if n.handshake != nil {
	//	n.handshake.Stop()
	// }
	if n.txPools != nil {
		n.txPools.Stop()
	}
	if n.p2pService != nil {
		n.p2pService.Stop()
	}
	if n.blockReader != nil {
		n.blockReader.Stop()
	}
	if n.stateDataDB != nil {
		n.stateDataDB.Close()
	}
	if n.blockDataDB != nil {
		n.blockDataDB.Close()
	}
	if n.crudDataDB != nil {
		n.crudDataDB.Close()
	}

	close(n.quiteCh)

	return nil
}

func (n *node) Wait() {
	n.lock.RLock()
	stop := n.quiteCh
	n.lock.RUnlock()

	<-stop
}

func (n *node) BlockReadWriter() protocol.BlockReadWriter {
	return n.blockReadWriter
}

func (n *node) ChainConf() protocol.Config {
	return n.config
}

func (n *node) GetStateDataDB() kvstore2.Database {
	return n.stateDataDB
}

func (n *node) GetTxPools() protocol.TxPools {
	return n.txPools
}

func (n *node) Apps() protocol.Apps {
	return n.apps
}

func (n *node) NodeKey() protocol.NodeKey {
	return n.nodeKey
}

func (n *node) Database() protocol.Database {
	return n.database
}

func (n *node) KVDatabase() kvstore2.Database {
	return n.stateDataDB
}

func (n *node) Config() protocol.Config {
	return n.config
}

func (n *node) APIs() protocol.APIs {
	return n.apis
}

func (n *node) AddTxPool(txType types.TxType, txPool protocol.TxPool) {
	n.txPools.Register(txType, txPool)
}

func (n *node) SetConsensus(consensus protocol.Consensus) error {
	n.consensus = consensus
	return nil
}
