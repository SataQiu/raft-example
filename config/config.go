package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// RaftConf defines the configuration for Raft
type RaftConf struct {
	NodeId    string `mapstructure:"node_id"`
	IP        string `mapstructure:"ip"`
	Port      int    `mapstructure:"port"`
	VolumeDir string `mapstructure:"volume_dir"`
	Peers     string `mapstructure:"peers"`
}

// ServerConf defines the configuration for HTTP server
type ServerConf struct {
	IP   string `mapstructure:"ip"`
	Port int    `mapstructure:"port"`
}

// Config defines the raft-example configuration
type Config struct {
	Raft   RaftConf   `mapstructure:"raft"`
	Server ServerConf `mapstructure:"server"`
}

// Load inits and returns the Config
func Load() *Config {
	var v = viper.New()

	v.AutomaticEnv()
	for _, env := range ConfKeys {
		if err := v.BindEnv(env); err != nil {
			log.Fatal(err)
		}
	}

	v.SetDefault(ServerIP, "127.0.0.1")
	v.SetDefault(RaftIP, "127.0.0.1")

	return &Config{
		Raft: RaftConf{
			NodeId:    v.GetString(RaftNodeId),
			IP:        v.GetString(RaftIP),
			Port:      v.GetInt(RaftPort),
			VolumeDir: v.GetString(RaftVolDir),
			Peers:     v.GetString(RaftPeers),
		},
		Server: ServerConf{
			IP:   v.GetString(ServerIP),
			Port: v.GetInt(ServerPort),
		},
	}
}

type Peer struct {
	ID      string
	Address string
}

func (c *Config) Peers() []Peer {
	peers := []Peer{
		{
			ID:      c.Raft.NodeId,
			Address: c.RaftAddress(),
		},
	}

	if len(c.Raft.Peers) == 0 {
		return peers
	}

	peerItems := strings.Split(strings.TrimSpace(c.Raft.Peers), ",")
	for _, peerItem := range peerItems {
		element := strings.Split(strings.TrimSpace(peerItem), "#")
		if len(element) != 2 {
			log.Fatal("failed to parse peers, must be ID1#Address1,ID2#Address2 format")
		}
		if element[1] != c.RaftAddress() {
			peers = append(peers, Peer{
				ID:      element[0],
				Address: element[1],
			})
		}
	}

	return peers
}

func (c *Config) RaftAddress() string {
	return fmt.Sprintf("%s:%d", c.Raft.IP, c.Raft.Port)
}
