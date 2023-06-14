package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"

	"github.com/SataQiu/raft-example/config"
	"github.com/SataQiu/raft-example/fsm"
	"github.com/SataQiu/raft-example/server"
)

func main() {
	conf := config.Load()
	log.Printf("%+v\n", conf)

	// Preparing badgerDB
	badgerOpt := badger.DefaultOptions(conf.Raft.VolumeDir)
	badgerDB, err := badger.Open(badgerOpt)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer func() {
		if err := badgerDB.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error close badgerDB: %s\n", err.Error())
		}
	}()

	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(conf.Raft.NodeId)
	raftConf.SnapshotThreshold = 1024

	fsmStore := fsm.NewBadger(badgerDB)

	store, err := raftboltdb.NewBoltStore(filepath.Join(conf.Raft.VolumeDir, "raft.dataRepo"))
	if err != nil {
		log.Fatal(err)
		return
	}

	// Wrap the store in a LogCache to improve performance.
	cacheStore, err := raft.NewLogCache(config.DefaultRaftLogCacheSize, store)
	if err != nil {
		log.Fatal(err)
		return
	}

	snapshotStore, err := raft.NewFileSnapshotStore(conf.Raft.VolumeDir, config.DefaultRaftSnapShotRetain, os.Stdout)
	if err != nil {
		log.Fatal(err)
		return
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", conf.RaftAddress())
	if err != nil {
		log.Fatal(err)
		return
	}

	transport, err := raft.NewTCPTransport(conf.RaftAddress(), tcpAddr, config.DefaultMaxPool, config.DefaultTcpTimeout, os.Stdout)
	if err != nil {
		log.Fatal(err)
		return
	}

	raftServer, err := raft.NewRaft(raftConf, fsmStore, cacheStore, store, snapshotStore, transport)
	if err != nil {
		log.Fatal(err)
		return
	}

	configuration := raft.Configuration{}
	for _, peer := range conf.Peers() {
		configuration.Servers = append(configuration.Servers, raft.Server{
			Suffrage: raft.Voter,
			ID:       raft.ServerID(peer.ID),
			Address:  raft.ServerAddress(peer.Address),
		})
	}

	raftServer.BootstrapCluster(configuration)

	srv := server.New(fmt.Sprintf("%s:%d", conf.Server.IP, conf.Server.Port), badgerDB, raftServer)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}

	return
}
