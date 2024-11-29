package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/joho/godotenv"
)

type ZkConfig struct {
	servers             []string
	sessionTimeoutAfter time.Duration
	zkConnection        *zk.Conn
	counterPath         string
	lockPath            string
}

func (c *ZkConfig) Start() error {
	conn, err := c.establishedConnection()
	if err != nil {
		return err
	}
	c.zkConnection = conn
	c.createNodeIfNotExists(c.counterPath, "0", zk.FlagPersistent)
	c.createNodeIfNotExists(c.lockPath, "", zk.FlagPersistent)
	return nil
}

func (c *ZkConfig) establishedConnection() (*zk.Conn, error) {
	conn, _, err := zk.Connect(c.servers, c.sessionTimeoutAfter)
	if err != nil {
		c.zkConnection = nil
		return nil, fmt.Errorf("unable to connect to Zookeeper: %v", err)
	}

	return conn, nil
}

func (c *ZkConfig) createNodeIfNotExists(node string, defaultVal string, zkFlags int32) error {
	exists, _, err := c.zkConnection.Exists(node)
	if err != nil {
		return fmt.Errorf("error on checking node on %v with error: %v", node, err)
	}
	if !exists {
		_, err = c.zkConnection.Create(
			node,
			[]byte(defaultVal),
			zkFlags,
			zk.WorldACL(zk.PermAll),
		)

		if err != nil {
			return fmt.Errorf("error on initializing node %v with error: %v", node, err)
		}
	}

	return nil
}

func (c *ZkConfig) Stop() {
	if c.zkConnection != nil {
		c.zkConnection.Close()
	}
}

// Getters
func (c *ZkConfig) ZkConnection() *zk.Conn {
	return c.zkConnection
}

func (c *ZkConfig) LockPath() string {
	return c.lockPath
}

func (c *ZkConfig) CounterPath() string {
	return c.counterPath
}

// \Getters

func NewZkConfig(
	sessionTimeoutAfter time.Duration,
	counterPath string,
	lockPath string,
) *ZkConfig {

	// Init config
	env, _ := godotenv.Read(".env")
	envVal := env["ZK_SERVER"]
	servers := strings.Split(envVal, ",")

	return &ZkConfig{
		servers:             servers,
		sessionTimeoutAfter: sessionTimeoutAfter,
		counterPath:         counterPath,
		lockPath:            lockPath,
	}
}
