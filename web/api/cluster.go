package api

import (
	prov "github.com/jeffjen/podd/provider"

	_ "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"

	"net/http"
)

func ClusterList(c *gin.Context) {
	type clusterOutput struct {
		Name   string `json:"name"`
		Min    int64  `json:"node_min"`
		Max    int64  `json:"node_max"`
		Count  int64  `json:"node_count"`
		Online int64  `json:"node_online"`
	}

	clusters, stop := AutoScaling.ListCluster()
	defer close(stop)

	output := struct {
		Size  int64
		Nodes []*clusterOutput
	}{
		Size:  0,
		Nodes: make([]*clusterOutput, 0),
	}

	for one := range clusters {
		name, min, max, count := one.Stats()
		output.Nodes = append(output.Nodes, &clusterOutput{
			Name:   name,
			Min:    min,
			Max:    max,
			Count:  count,
			Online: one.Online(),
		})
		output.Size += 1
	}

	c.JSON(http.StatusOK, output)
}

func ClusterUpdate(c *gin.Context) {
	var opts prov.ScalePolicy
	if err := c.Bind(&opts); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if !opts.VerfiyScaleConstraint() {
		c.AbortWithStatus(http.StatusExpectationFailed)
		return
	}

	// Check that the cluster is registered
	cluster := AutoScaling.GetCluster(c.Param("name"))
	if cluster == nil {
		c.AbortWithStatus(http.StatusExpectationFailed)
		return
	}

	// Request provider to update cluster size
	if err := cluster.Configure(opts.Min, opts.Max, opts.Count); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.String(http.StatusOK, "done")
}

func ClusterCreate(c *gin.Context) {
	var cOpts prov.ClusterOptions
	if err := c.Bind(&cOpts); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Check that the cluster is not already registered
	cOpts.Name = c.Param("name")
	if AutoScaling.GetCluster(cOpts.Name) != nil {
		c.AbortWithStatus(http.StatusExpectationFailed)
		return
	}

	// Verify that the request for node scaling is valid
	if !cOpts.VerfiyScaleConstraint() {
		c.AbortWithStatus(http.StatusExpectationFailed)
		return
	}

	// If Root (being the cluster root identifier) is not given, use Default
	if cOpts.Root == "" {
		cOpts.Root = prov.ClusterGroup
	}

	// Request provider to register a new cluster
	if err := AutoScaling.Register(cOpts); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.String(http.StatusOK, "done")
}
