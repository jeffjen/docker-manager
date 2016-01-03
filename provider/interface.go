package provider

const (
	// spread the srevice to as much instances as possible
	SpreadService = "spread"

	// pack service to a node as much as posible
	BinpackService = "binpack"
)

type ServiceExpectation struct {
	// Name of the service
	Name string

	// Min, Max of cluster size and current size
	Min   uint
	Max   uint
	Count uint

	// Service scaling strategy
	// @See SpreadService, BinpackService
	Strategy string

	// Constraints applied to physical instance carrying service
	cOpts ClusterOptions

	// Relief mechanism should expectations not satisfied
	rOpts ReliefOptions
}

type ClusterOptions struct {
	// Name of the cluster
	Name string

	// Min, Max of cluster size and current size
	Min   uint
	Max   uint
	Count uint

	// Launch Configuration resource
	LaunchConfig string
}

type TerminatePolicy struct {
}

type AutoScaling interface {
	// Set expectations of the service
	SetExpectation(exp ServiceExpectation)

	// Scale the cluster size UP
	Scaleup(service string, n uint) bool

	// Scale the cluster size DOWN
	Scalein(service string, n uint, policy TerminatePolicy) bool
}
