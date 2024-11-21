package handler

import (
	"fmt"
	"time"

	"github.com/go-zookeeper/zk"
)

type IncHandler struct {
	zkConnection *zk.Conn
	lockPath     string
	getHandler   *GetHandler
}

func (h *IncHandler) Key() string {
	return "INC"
}

func (h *IncHandler) Handle() error {

	// Leader Election: Create a lock node and try to acquire the lock
	var attempt int32 = 0
	var lockPath string = h.lockPath + "/lock-"
	lockNode, err := electLeader(lockPath, h.zkConnection, attempt)
	if err != nil {
		return err
	}

	//We are the leader now, attempt to get the current counter
	counter, zkStat, err := h.getHandler.Handle()
	if err != nil {
		return err
	}

	//Attempt to increase the counter
	var counterPath string = h.getHandler.counterPath
	err = inc(counter, counterPath, h.zkConnection, zkStat)
	if err != nil {
		return err
	}

	// Release the lock
	h.zkConnection.Delete(lockNode, -1)

	return nil
}

func electLeader(lockNode string, zkConnection *zk.Conn, attempt int32) (string, error) {
	path, err := zkConnection.Create(
		lockNode,
		[]byte{},
		zk.FlagEphemeral,
		zk.WorldACL(zk.PermAll),
	)

	//Log data
	fmt.Printf("Created path: %v\n", path)

	//Node already exists => try again later
	if err == zk.ErrNodeExists {

		//Max attempt exceed
		if attempt > 3 {
			return "", fmt.Errorf("max relection attempt exceed")
		}

		fmt.Printf("Leader already exists on attempt %v, waiting to become leader \n", attempt)
		time.Sleep(2 * time.Second)
		attempt++
		return electLeader(lockNode, zkConnection, attempt)
	}

	if err != nil {
		return "", fmt.Errorf("could not create lock: %v", err)
	}

	return path, nil
}

func inc(
	counter int32,
	counterPath string,
	zkConnection *zk.Conn,
	zkStat *zk.Stat,
) error {

	counter++
	fmt.Printf("Attempt to increment counter to %d\n", counter)

	counterAsByte := []byte(fmt.Sprintf("%d", counter))
	_, err := zkConnection.Set(counterPath, counterAsByte, zkStat.Version)
	if err != nil {
		return fmt.Errorf("could not set counter value: %v", err)
	}

	return nil
}

func NewIncHandler(
	zkConnection *zk.Conn,
	lockPath string,
	getHandler *GetHandler,
) *IncHandler {

	return &IncHandler{
		zkConnection: zkConnection,
		lockPath:     lockPath,
		getHandler:   getHandler,
	}

}
