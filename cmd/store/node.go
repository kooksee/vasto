package store

import (
	"fmt"
	"github.com/chrislusf/vasto/storage/binlog"
	"github.com/chrislusf/vasto/storage/rocks"
	"github.com/chrislusf/vasto/topology/cluster_listener"
	"log"
	"os"
	"os/user"
	"strings"
)

type node struct {
	id                int
	serverId          int
	db                *rocks.Rocks
	lm                *binlog.LogManager
	clusterListener   *cluster_listener.ClusterListener
	replicationFactor int
	// just to avoid repeatedly create these variables
	nextSegmentKey []byte
	nextOffsetKey  []byte
}

func newNodes(option *StoreOption, clusterListener *cluster_listener.ClusterListener) (nodes []*node, err error) {
	for i := 0; i < *option.ReplicationFactor; i++ {
		id := int(*option.Id) - i
		if id < 0 {
			id += clusterListener.ExpectedSize()
		}
		if i != 0 && id == int(*option.Id) {
			break
		}
		dir := fmt.Sprintf("%s/%d", *option.Dir, id)
		if strings.HasPrefix(dir, "~") {
			usr, err := user.Current()
			if err != nil {
				log.Fatal(err)
			}
			dir = usr.HomeDir + dir[1:]
		}
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, fmt.Errorf("mkdir %s: %v", dir, err)
		}
		node := newNode(dir, int(*option.Id), id, clusterListener,
			*option.ReplicationFactor, *option.LogFileSizeMb, *option.LogFileCount)
		nodes = append(nodes, node)
		if i != 0 {
			go node.start()
		}
	}
	return nodes, nil
}

func newNode(dir string, serverId, nodeId int, clusterListener *cluster_listener.ClusterListener,
	replicationFactor int, logFileSizeMb int, logFileCount int) *node {
	n := &node{
		id:                nodeId,
		serverId:          serverId,
		db:                rocks.New(dir),
		clusterListener:   clusterListener,
		replicationFactor: replicationFactor,
	}
	if logFileSizeMb > 0 {
		n.lm = binlog.NewLogManager(dir, nodeId, int64(logFileSizeMb*1024*1024), logFileCount)
		n.lm.Initialze()
	}
	n.nextSegmentKey = []byte(fmt.Sprintf("%d.next.segment", n.id))
	n.nextOffsetKey = []byte(fmt.Sprintf("%d.next.offset", n.id))

	return n
}

func (n *node) start() {
	err := n.bootstrap()
	if err != nil {
		log.Fatalf("bootstrap: %v", err)
	}
	n.follow()
}