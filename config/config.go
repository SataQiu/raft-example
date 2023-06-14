package config

import (
	"log"

	"github.com/spf13/viper"
)

// RaftConf defines the configuration for Raft
type RaftConf struct {
	NodeId    string `mapstructure:"node_id"`
	IP        string `mapstructure:"ip"`
	Port      int    `mapstructure:"port"`
	VolumeDir string `mapstructure:"volume_dir"`
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
		},
		Server: ServerConf{
			IP:   v.GetString(ServerIP),
			Port: v.GetInt(ServerPort),
		},
	}
}
