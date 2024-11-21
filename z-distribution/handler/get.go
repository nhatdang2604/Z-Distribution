package handler

import (
	"fmt"

	"github.com/go-zookeeper/zk"
)

type GetHandler struct {
	zkConnection *zk.Conn
	counterPath  string
}

func (h *GetHandler) Key() string {
	return "GET"
}

func (h *GetHandler) Handle() (int32, *zk.Stat, error) {
	data, zkStat, err := h.zkConnection.Get(h.counterPath)
	if err != nil {
		return 0, zkStat, fmt.Errorf("get counter value with error on connection: %v", err)
	}

	var counter int32
	_, err = fmt.Sscanf(string(data), "%d", &counter)
	if err != nil {
		return 0, zkStat, fmt.Errorf("get counter value with error on parsing: %v", err)
	}

	return counter, zkStat, nil
}

func NewGetHandler(zkConnection *zk.Conn, counterPath string) *GetHandler {
	return &GetHandler{
		zkConnection: zkConnection,
		counterPath:  counterPath,
	}
}
