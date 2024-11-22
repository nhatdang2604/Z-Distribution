package handler

import (
	"fmt"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/nhatdang2604/z-distribution/engine/zk/config"
)

type IncWithSleepHandler struct {
	zkConfig   *config.ZkConfig
	getHandler *GetHandler
}

func (h *IncWithSleepHandler) Key() string {
	return "SLEEPINC"
}

func (h *IncWithSleepHandler) Handle() error {

	// Leader Election: Create a lock node and try to acquire the lock
	var zkConnection *zk.Conn = h.zkConfig.ZkConnection()
	var parentLockPath string = h.zkConfig.LockPath()
	var attempt int32 = 0
	var lockPath string = parentLockPath + "/lock-"
	lockNode, err := h.electLeader(lockPath, zkConnection, attempt)
	if err != nil {
		return err
	}

	// Claim the lock
	_, lockZkStat, err := zkConnection.Get(lockNode)
	if err != nil {
		return fmt.Errorf("error on claiming lock as %v with err: %v", lockNode, err)
	}

	// Release the lock
	defer func() {
		err := zkConnection.Delete(lockNode, lockZkStat.Version)
		if err != nil {
			fmt.Printf("Error on claiming lock as %v with err: %v\n", lockNode, err)
		}
	}()

	//We are the leader now, attempt to get the current counter
	counter, counterZkStat, err := h.getHandler.Handle()
	if err != nil {
		return err
	}

	time.Sleep(10 * time.Second) // Force race condition occurred

	//Attempt to increase the counter
	var counterPath string = h.zkConfig.CounterPath()
	err = h.inc(counter, counterPath, zkConnection, counterZkStat)
	if err != nil {
		return err
	}

	return nil
}

func (h *IncWithSleepHandler) electLeader(lockNode string, zkConnection *zk.Conn, attempt int32) (string, error) {
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
		return h.electLeader(lockNode, zkConnection, attempt)
	}

	if err != nil {
		return "", fmt.Errorf("could not create lock %v with error: %v", lockNode, err)
	}

	return path, nil
}

func (h *IncWithSleepHandler) inc(
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

func NewIncWithSleepHandler(
	zkConfig *config.ZkConfig,
	getHandler *GetHandler,
) *IncWithSleepHandler {

	return &IncWithSleepHandler{
		zkConfig:   zkConfig,
		getHandler: getHandler,
	}

}
