package handler

import (
	"fmt"

	"z-election/zk/config"

	"github.com/go-zookeeper/zk"
)

type IncHandler struct {
	zkConfig   *config.ZkConfig
	getHandler *GetHandler
}

func (h *IncHandler) Key() string {
	return "INC"
}

func (h *IncHandler) Handle(concensusPath string) error {

	//We are the leader now, attempt to get the current value
	consensusVal, zkStat, err := h.getHandler.Handle(concensusPath)
	if err != nil {
		return err
	}

	//Attempt to increase the value
	zkConnection := h.zkConfig.ZkConnection()
	err = h.inc(consensusVal, concensusPath, zkConnection, zkStat)
	if err != nil {
		return err
	}

	return nil
}

func (h *IncHandler) inc(
	consensusVal int32,
	consensusPath string,
	zkConnection *zk.Conn,
	zkStat *zk.Stat,
) error {

	consensusVal++
	nodeId := h.zkConfig.NodeId()
	fmt.Printf("Node with id=%v is attempting to increment consensus value to %d\n", nodeId, consensusVal)

	consensusAsByte := []byte(fmt.Sprintf("%d", consensusVal))
	_, err := zkConnection.Set(consensusPath, consensusAsByte, zkStat.Version)
	if err != nil {
		return fmt.Errorf("node with id=%v could not set consensus value: %v", nodeId, err)
	}

	return nil
}

func NewIncHandler(
	zkConfig *config.ZkConfig,
	getHandler *GetHandler,
) *IncHandler {

	return &IncHandler{
		zkConfig:   zkConfig,
		getHandler: getHandler,
	}

}
