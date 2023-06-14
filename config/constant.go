package config

import "time"

const (
	// DefaultMaxPool controls how many connections we will pool.
	DefaultMaxPool = 3

	// DefaultTcpTimeout is used to apply I/O deadlines. For InstallSnapshot, we multiply
	// the timeout by (SnapshotSize / TimeoutScale).
	// https://github.com/hashicorp/raft/blob/v1.1.2/net_transport.go#L177-L181
	DefaultTcpTimeout = 10 * time.Second

	// DefaultRaftSnapShotRetain controls how many
	// snapshots are retained. Must be at least 1.
	DefaultRaftSnapShotRetain = 2

	// DefaultRaftLogCacheSize is the maximum number of logs to cache in-memory.
	// This is used to reduce disk I/O for the recently committed entries.
	DefaultRaftLogCacheSize = 512
)

const (
	ServerIP   = "SERVER_IP"
	ServerPort = "SERVER_PORT"
	RaftNodeId = "RAFT_NODE_ID"
	RaftIP     = "RAFT_IP"
	RaftPort   = "RAFT_PORT"
	RaftVolDir = "RAFT_VOL_DIR"
)

var ConfKeys = []string{
	ServerIP,
	ServerPort,
	RaftNodeId,
	RaftIP,
	RaftPort,
	RaftVolDir,
}
