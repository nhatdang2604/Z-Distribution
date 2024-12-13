package zk

import (
	"fmt"
	"strings"
	"z-election/zk/config"
	"z-election/zk/handler"

	"github.com/go-zookeeper/zk"
)

type ElectionCandidate struct {
	config     *config.ZkConfig
	getHandler *handler.GetHandler
	incHandler *handler.IncHandler
}

func (ec *ElectionCandidate) Start() error {

	// Init the Zookeeper config
	err := ec.config.Start()
	if err != nil {
		return err
	}

	return nil
}

func (ec *ElectionCandidate) Stop() {
	ec.config.Stop()
}

// Handle commands from the user
func (ec *ElectionCandidate) Handle(cmd string) (interface{}, error) {
	nodeId := ec.config.NodeId()
	initAttempt := 0
	candidateNode, err := ec.elect(initAttempt)
	if err != nil {
		return nil, err
	}

	//Reclaim the candidate node, to release later
	_, candidateZkStat, err := ec.config.ZkConnection().Get(candidateNode)
	if err != nil {
		return nil, fmt.Errorf("node with id=%v in attempt to get candidate node=%v had error: %v", nodeId, candidateNode, err)
	}

	//Release the candidate node later
	defer func() {
		err := ec.config.ZkConnection().Delete(candidateNode, candidateZkStat.Version)
		if err != nil {
			fmt.Printf("Node with id=%v had error on releasing node=%v with err: %v\n", nodeId, candidateNode, err)
		}
	}()

	isLeader, err := ec.isBecomeLeader(candidateNode)
	if err != nil {
		return nil, err
	}

	if isLeader {
		return ec.execute(cmd)
	} else {
		fmt.Printf("Node with id=%v is not the leader\n", ec.config.NodeId())
	}

	return nil, nil
}

func (ec *ElectionCandidate) elect(attempt int) (string, error) {

	//Max attempt to elect leader
	if attempt > 3 {
		return "", fmt.Errorf("node with id=%v in attempt to elect leader had error: too many attempts", ec.config.NodeId())
	}

	//Create candidate node
	nodeId := ec.config.NodeId()
	leaderPath := ec.config.LeaderPath()
	candidateNode := leaderPath + "/candidate-"
	actualCandidateNode, err := ec.config.ZkConnection().Create(candidateNode, []byte(fmt.Sprintf("%v", nodeId)), zk.FlagEphemeralSequential, zk.WorldACL(zk.PermAll))

	//Retry to elect if error
	if err != nil {
		fmt.Printf("Node with id=%v in attempt %v to create candidate path at %v had error: %v\n", nodeId, attempt, actualCandidateNode, err)
		return ec.elect(attempt + 1)
	}

	fmt.Printf("Node with id=%v create candidate path at %v successfully\n", nodeId, actualCandidateNode)

	return actualCandidateNode, nil
}

func (ec *ElectionCandidate) isBecomeLeader(candidateNode string) (bool, error) {
	nodeId := ec.config.NodeId()
	leaderPath := ec.config.LeaderPath()
	childrenNodes, _, err := ec.config.ZkConnection().Children(leaderPath)
	if err != nil {
		return false, fmt.Errorf("node with id=%v with candidate node=%v in attempt to get children nodes at %v had error: %v", nodeId, candidateNode, leaderPath, err)
	}

	if len(childrenNodes) == 0 {
		return false, fmt.Errorf("node with id=%v with candidate node=%v in attempt to get children nodes at %v had error: no children nodes", nodeId, candidateNode, leaderPath)
	}

	// Check if the candidate node is the leader with the smallest sequence number
	leaderNode := ""
	for _, childNode := range childrenNodes {
		toCompareChildNode := leaderPath + "/" + childNode
		if (leaderNode == "") || (leaderNode < toCompareChildNode) {
			leaderNode = toCompareChildNode
		}
	}

	return (leaderNode == candidateNode), nil
}

func (ec *ElectionCandidate) execute(cmd string) (interface{}, error) {
	switch cmd {
	case strings.ToLower(ec.incHandler.Key()):
		{

			//Increment the counter
			err := ec.incHandler.Handle(ec.config.ConsensusPath())
			if err != nil {
				return nil, err
			}

			//Get and print the concensusVal value
			cmd := strings.ToLower(ec.getHandler.Key())
			return ec.execute(cmd)
		}

	case strings.ToLower(ec.getHandler.Key()):
		{

			//Get the concensusVal value
			concensusVal, _, err := ec.getHandler.Handle(ec.config.ConsensusPath())
			if err != nil {
				return nil, err
			}

			fmt.Printf("Node with id=%v has consensus value=%v\n", ec.config.NodeId(), concensusVal)

			return concensusVal, nil
		}

	default:
		{
			return nil, fmt.Errorf("node with id=%v had error on executing command: unknown command", ec.config.NodeId())
		}
	}
}

func NewElectionCandidate(config *config.ZkConfig) *ElectionCandidate {
	getHandler := handler.NewGetHandler(config)
	incHandler := handler.NewIncHandler(config, getHandler)
	return &ElectionCandidate{
		config:     config,
		getHandler: getHandler,
		incHandler: incHandler,
	}
}
