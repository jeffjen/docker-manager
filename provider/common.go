package provider

import (
	"text/template"
)

const (
	UserData = `#!/bin/bash

curl -sSL -O https://raw.githubusercontent.com/jeffjen/aws-devops/master/bootstrap.sh
chmod +x bootstrap.sh

INSTANCE_OPTS="--reboot --dockeruser ubuntu {{if .Swapsize}}--swap {{.Swapsize}}{{end}}"

NOTIFICATION_OPTS="{{if .WebHook}}--agent-notify-uri {{.WebHook}} --agent-notify-channel {{.Channel}}{{end}}"

AGENT_OPTS="--discovery etcd://172.99.0.154:2379 --cluster {{.Root}}/{{.Name}} ${NOTIFICATION_OPTS}"

./bootstrap.sh ${INSTANCE_OPTS} ${AGENT_OPTS}
`
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

	cloudInitTmpl = template.Must(template.New("cloud-init").Parse(UserData))
)

func New(name string) AutoScaling {
	handle, ok := autoScalingByType[name]
	if ok {
		return handle()
	} else {
		return nil
	}
}
