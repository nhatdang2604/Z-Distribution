package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/joho/godotenv"
	"github.com/nhatdang2604/z-distribution/handler"
)

type ZkConfig struct {
	servers             []string
	sessionTimeoutAfter time.Duration
	zkConnection        *zk.Conn
	getHandler          *handler.GetHandler
	incHandler          *handler.IncHandler
}

func (c *ZkConfig) Connect() (*zk.Conn, error) {
	conn, _, err := zk.Connect(c.servers, c.sessionTimeoutAfter)
	if err != nil {
		c.zkConnection = nil
		return nil, fmt.Errorf("unable to connect to Zookeeper: %v", err)
	}

	c.zkConnection = conn
	return c.zkConnection, nil
}

func (c *ZkConfig) Init(
	counterPath string,
	lockPath string,
) error {

	//Init the counter if it doesn't exist
	c.createNodeIfNotExists(counterPath, "0")
	c.createNodeIfNotExists(lockPath, "")

	// Init handlers
	var getHandler *handler.GetHandler = handler.NewGetHandler(
		c.zkConnection,
		counterPath,
	)
	var incHandler *handler.IncHandler = handler.NewIncHandler(
		c.zkConnection,
		lockPath,
		getHandler,
	)
	c.getHandler = getHandler
	c.incHandler = incHandler

	return nil
}

func (c *ZkConfig) createNodeIfNotExists(node string, defaultVal string) error {
	exists, _, err := c.zkConnection.Exists(node)
	if err != nil {
		return fmt.Errorf("error on checking node on %v with error: %v", node, err)
	}
	if !exists {
		_, err = c.zkConnection.Create(
			node,
			[]byte(defaultVal),
			zk.FlagPersistent,
			zk.WorldACL(zk.PermAll),
		)

		if err != nil {
			return fmt.Errorf("error on initializing node %v with error: %v", node, err)
		}
	}

	return nil
}

func (c *ZkConfig) Close() {
	c.zkConnection.Close()
}

// Handle commands from the user
func (c *ZkConfig) Handle(command string) {
	switch command {
	case c.getHandler.Key():
		counter, _, err := c.getHandler.Handle()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Current counter value: %d\n", counter)
		}
	case "INC":
		err := c.incHandler.Handle()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Increase counter successfully")
		}
	default:
		fmt.Println("Invalid command. Use 'GET' or 'INC'.")
	}
}

func NewZkConfig() *ZkConfig {

	// Init config
	env, _ := godotenv.Read(".env")
	envVal := env["ZK_SERVER"]
	servers := strings.Split(envVal, ",")

	return &ZkConfig{
		servers:             servers,
		sessionTimeoutAfter: 10 * time.Second,
	}
}
