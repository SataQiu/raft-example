package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

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

	// 启用基于 leader 选举的示例业务服务
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		started := false
		for leader := range raftServer.LeaderCh() {
			if leader && !started {
				log.Println("Become Leader")
				go DoBusiness(ctx)
			}
			if !leader && started {
				log.Println("Lost Leader")
				return
			}
		}
	}()

	srv := server.New(fmt.Sprintf("%s:%d", conf.Server.IP, conf.Server.Port), badgerDB, raftServer)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}

	return
}

func DoBusiness(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Business terminated")
			return
		case t := <-ticker.C:
			log.Printf("Do business at %v\n", t)
		}
	}
}
