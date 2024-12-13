package handler

import (
	"fmt"

	"z-election/zk/config"

	"github.com/go-zookeeper/zk"
)

type GetHandler struct {
	zkConfig *config.ZkConfig
}

func (h *GetHandler) Key() string {
	return "GET"
}

func (h *GetHandler) Handle(concensusPath string) (int32, *zk.Stat, error) {
	nodeId := h.zkConfig.NodeId()
	zkConnection := h.zkConfig.ZkConnection()
	data, zkStat, err := zkConnection.Get(concensusPath)
	if err != nil {
		return 0, zkStat, fmt.Errorf("node with id=%v had error on getting concensus value: %v", nodeId, err)
	}

	var consensusVal int32
	_, err = fmt.Sscanf(string(data), "%d", &consensusVal)
	if err != nil {
		return 0, zkStat, fmt.Errorf("node with id=%v had error on getting consensus value with parsing: %v", nodeId, err)
	}

	return consensusVal, zkStat, nil
}

func NewGetHandler(zkConfig *config.ZkConfig) *GetHandler {
	return &GetHandler{zkConfig: zkConfig}
}
