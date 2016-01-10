package api

import (
	disc "github.com/jeffjen/go-discovery"

	_ "github.com/Sirupsen/logrus"
	etcd "github.com/coreos/etcd/client"
	ctx "golang.org/x/net/context"

	"encoding/json"
	"net/http"
	"path"
	"strconv"
	"time"
)

const (
	DefaultRootPath = "/"

	DefaultListClusterSize = 10
)

type Clusters struct {
	Root  string
	Size  int
	Nodes []string
}

func checkCluster(c *Clusters, n *etcd.Node) {
	if !n.Dir {
		return // skipping because a cluster root is always a dir
	}
	for _, node := range n.Nodes { // traverse Depth First
		if c.Size <= len(c.Nodes) {
			return // capacity reached for this query
		}
		if node.Dir {
			if path.Base(node.Key) == "docker" {
				c.Nodes = append(c.Nodes, n.Key[1:len(n.Key)])
			} else {
				checkCluster(c, node)
			}
		}
	}
}

func ClusterList(w http.ResponseWriter, r *http.Request) {
	if err := common("GET", r); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// We are sending json data
	w.Header().Add("Content-Type", "application/json")

	var RootPath = DefaultRootPath
	if root := r.Form.Get("root"); root != "" {
		RootPath = root
	}

	var ClusterSize = DefaultListClusterSize
	if projSizeStr := r.Form.Get("size"); projSizeStr != "" {
		projSize, err := strconv.ParseInt(projSizeStr, 10, 0)
		if err == nil {
			ClusterSize = int(projSize)
		}
	}
	var ClusterNodes = &Clusters{
		Root:  RootPath,
		Size:  ClusterSize,
		Nodes: make([]string, 0, ClusterSize),
	}

	kAPI, _ := disc.NewKeysAPI(etcd.Config{
		Endpoints: disc.Endpoints(),
	})

	work, abort := ctx.WithTimeout(ctx.Background(), 3*time.Second)
	defer abort()

	resp, err := kAPI.Get(work, RootPath, &etcd.GetOptions{
		Recursive: true,
	})
	if err != nil {
		json.NewEncoder(w).Encode(ClusterNodes)
		return
	}

	// Go traverse the discovery tree
	for _, node := range resp.Node.Nodes {
		checkCluster(ClusterNodes, node)
	}
	ClusterNodes.Size = len(ClusterNodes.Nodes)

	json.NewEncoder(w).Encode(ClusterNodes)
	return
}

func ClusterCreate(w http.ResponseWriter, r *http.Request) {
	if err := common("POST", r); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	http.Error(w, ErrNotImplemented.Error(), 403)
	return
}
