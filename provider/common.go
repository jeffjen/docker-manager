package provider

import (
	"text/template"
)

const (
	ClusterGroup = "/cluster"

	// spread the srevice to as much instances as possible
	SpreadService = "spread"

	// pack service to a node as much as posible
	BinpackService = "binpack"
)

var (
	autoScalingByType = map[string]func() AutoScaling{
		"AWS": newAWS,
	}

	cloudInitTmpl = template.Must(template.ParseFiles("cloud-init/init"))
)

func New(name string) AutoScaling {
	handle, ok := autoScalingByType[name]
	if ok {
		return handle()
	} else {
		return nil
	}
}
