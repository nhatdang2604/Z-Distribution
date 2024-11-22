package handler

import (
	"fmt"

	"github.com/go-zookeeper/zk"
	"github.com/nhatdang2604/z-distribution/engine/zk/config"
)

type GetHandler struct {
	zkConfig *config.ZkConfig
}

func (h *GetHandler) Key() string {
	return "GET"
}

func (h *GetHandler) Handle() (int32, *zk.Stat, error) {
	zkConnection := h.zkConfig.ZkConnection()
	counterPath := h.zkConfig.CounterPath()
	data, zkStat, err := zkConnection.Get(counterPath)
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

func NewGetHandler(zkConfig *config.ZkConfig) *GetHandler {
	return &GetHandler{zkConfig: zkConfig}
}
