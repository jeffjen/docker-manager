package provider

type ServiceExpectation struct {
	// Name of the service
	Name string

	// Min, Max of cluster size and current size
	Min   int64
	Max   int64
	Count int64

	// Service scaling strategy
	// @See SpreadService, BinpackService
	Strategy string

	// Constraints applied to physical instance carrying service
	ClusterOptions
}

func (s ServiceExpectation) VerfiyScheduleConstraint() bool {
	return s.Min <= s.Max && s.Count <= s.Max && s.Count >= s.Min
}

type ClusterOptions struct {
	// Cluster root identifier
	Root string `form:"group" json:"group"`

	// Cluster discovery provider
	Discovery string `form:"discovery" json:"discovery"`

	// Name of the cluster
	Name string `form:"name" json:"name"`

	// Min, Max of cluster size and current size
	Min   int64 `form:"node_min" json:"node_min"`
	Max   int64 `form:"node_max" json:"node_max"`
	Count int64 `form:"node_count" json:"node_count"`

	// Instance type by provider
	Type string `form:"type" json:"type" binding:"required"`

	// Access permission and role from provider
	Role string `form:"role" json:"role" binding:"required"`

	// Swap size for launched instances
	Swapsize string `form:"swapsize" json:"swapsize" binding:"required"`

	// Request to obtain public IP
	PublicIP bool `form:"public" json:"public"`

	// URI to post our message
	WebHook string `form:"web_hook" json:"web_hook"`

	// Channel a message to be post on
	Channel string `form:"channel" json:"channel"`
}

func (c ClusterOptions) VerfiyScaleConstraint() bool {
	return c.Min <= c.Max && c.Count <= c.Max && c.Count >= c.Min
}

type ScalePolicy struct {
	Min   int64 `form:"node_min" json:"node_min"`
	Max   int64 `form:"node_max" json:"node_max"`
	Count int64 `form:"node_count" json:"node_count"`
}

func (p ScalePolicy) VerfiyScaleConstraint() bool {
	return p.Min <= p.Max && p.Count <= p.Max && p.Count >= p.Min
}

type AutoScaling interface {
	// Provision a new Cluster through provider
	Register(opts ClusterOptions) error

	// Retrieve a cluster by identifier
	GetCluster(name string) Cluster

	// Iterate through registerd clusters
	ListCluster() (cluster <-chan Cluster, stop chan<- struct{})
}

type Cluster interface {
	// Change the physical size of cluster
	Configure(min, max, count int64) error

	// Report configured nodes in cluster
	Online() int64

	// Report cluster information
	Stats() (name string, min, max, count int64)

	// Obtain a copy of the options used to register/create this cluster
	Describe() ClusterOptions
}
