package provider

import (
	disc "github.com/jeffjen/go-discovery"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	etcd "github.com/coreos/etcd/client"
	ctx "golang.org/x/net/context"

	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path"
)

var (
	AWS_EC2_AMI_ID = os.Getenv("AWS_EC2_AMI_ID")

	AWS_EC2_INSTANCE_SSH_KEY = os.Getenv("AWS_EC2_INSTANCE_SSH_KEY")

	AWS_AUTOSCALING_GROUP_DISCOVERY = os.Getenv("AWS_AUTOSCALING_GROUP_DISCOVERY")

	AWS_AUTOSCALING_GROUP_VPC_SECURITY_GROUPS = os.Getenv("AWS_AUTOSCALING_GROUP_VPC_SECURITY_GROUPS")

	AWS_AUTOSCALING_GROUP_VPC_FUNCTION = os.Getenv("AWS_AUTOSCALING_GROUP_VPC_FUNCTION")

	AWS_AUTOSCALING_GROUP_VPC_SUBNETS = os.Getenv("AWS_AUTOSCALING_GROUP_VPC_SUBNETS")
)

func prepareLaunchConfig(cOpts *ClusterOptions) *autoscaling.CreateLaunchConfigurationInput {
	var buf = new(bytes.Buffer)

	cloudInitTmpl.Execute(buf, cOpts)

	return &autoscaling.CreateLaunchConfigurationInput{
		LaunchConfigurationName:  aws.String(cOpts.Name),
		AssociatePublicIpAddress: aws.Bool(cOpts.PublicIP),
		BlockDeviceMappings: []*autoscaling.BlockDeviceMapping{
			&autoscaling.BlockDeviceMapping{ // Storage device for Docker metadata and thinpool
				DeviceName: aws.String("xvdb"),
				Ebs: &autoscaling.Ebs{
					VolumeSize:          aws.Int64(16),
					VolumeType:          aws.String("gp2"),
					Iops:                nil,
					DeleteOnTermination: aws.Bool(true),
					Encrypted:           aws.Bool(false),
				},
			},
		},
		EbsOptimized:       aws.Bool(false),
		IamInstanceProfile: aws.String(cOpts.Role),
		ImageId:            aws.String(AWS_EC2_AMI_ID),
		InstanceMonitoring: &autoscaling.InstanceMonitoring{Enabled: aws.Bool(false)},
		InstanceType:       aws.String(cOpts.Type),
		KeyName:            aws.String(AWS_EC2_INSTANCE_SSH_KEY),
		SecurityGroups: []*string{
			aws.String(AWS_AUTOSCALING_GROUP_VPC_SECURITY_GROUPS),
		},
		UserData: aws.String(base64.StdEncoding.EncodeToString(buf.Bytes())),
	}
}

func prepareAutoScaling(cOpts *ClusterOptions) *autoscaling.CreateAutoScalingGroupInput {
	return &autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName:    aws.String(cOpts.Name),
		MaxSize:                 aws.Int64(cOpts.Max),
		MinSize:                 aws.Int64(cOpts.Min),
		DesiredCapacity:         aws.Int64(cOpts.Count),
		LaunchConfigurationName: aws.String(cOpts.Name),
		LoadBalancerNames:       []*string{}, // TODO: skip setting up LoadBalancerNames setup
		Tags: []*autoscaling.Tag{
			{
				Key:               aws.String("engine"),
				PropagateAtLaunch: aws.Bool(true),
				Value:             aws.String("docker"),
			},
			{
				Key:               aws.String("management"),
				PropagateAtLaunch: aws.Bool(true),
				Value:             aws.String("swarm"),
			},
			{
				Key:               aws.String("cluster"),
				PropagateAtLaunch: aws.Bool(true),
				Value:             aws.String(cOpts.Name),
			},
			{
				Key:               aws.String("type"),
				PropagateAtLaunch: aws.Bool(true),
				Value:             aws.String(AWS_AUTOSCALING_GROUP_VPC_FUNCTION),
			},
		},
		VPCZoneIdentifier: aws.String(AWS_AUTOSCALING_GROUP_VPC_SUBNETS),
	}
}

type cluster struct {
	auto *autoscaling.AutoScaling
	kAPI etcd.KeysAPI

	*ClusterOptions
}

func (c *cluster) Configure(min, max, count int64) error {
	params := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName:             aws.String(c.Name),
		MaxSize:                          aws.Int64(max),
		MinSize:                          aws.Int64(min),
		DesiredCapacity:                  aws.Int64(count),
		NewInstancesProtectedFromScaleIn: aws.Bool(false),
		TerminationPolicies: []*string{
			aws.String("Default"),
		},
		VPCZoneIdentifier: aws.String(AWS_AUTOSCALING_GROUP_VPC_SUBNETS),
	}

	_, err := c.auto.UpdateAutoScalingGroup(params)

	if err == nil {
		c.Max = *params.MaxSize
		c.Min = *params.MinSize
		c.Count = *params.DesiredCapacity
	}

	return err
}

func (c *cluster) Online() int64 {
	root := path.Join(c.Root, c.Name, "docker/swarm/nodes")
	active, err := c.kAPI.Get(ctx.Background(), root, nil)
	if err != nil {
		return 0
	} else {
		return int64(len(active.Node.Nodes))
	}
}

func (c *cluster) Stats() (name string, min, max, count int64) {
	name, min, max, count = c.Name, c.Min, c.Max, c.Count
	return
}

func (c *cluster) Describe() ClusterOptions {
	return *c.ClusterOptions
}

type service struct {
	cluster map[string]Cluster

	auto *autoscaling.AutoScaling
	kAPI etcd.KeysAPI
}

func (svc *service) Register(opts ClusterOptions) error {
	var cOpts = &opts
	if cOpts.Discovery == "" {
		cOpts.Discovery = AWS_AUTOSCALING_GROUP_DISCOVERY
	}
	_, err := svc.auto.CreateLaunchConfiguration(prepareLaunchConfig(cOpts))
	if err != nil {
		return err
	}
	_, err = svc.auto.CreateAutoScalingGroup(prepareAutoScaling(cOpts))
	if err != nil {
		return err
	}
	svc.cluster[cOpts.Name] = &cluster{svc.auto, svc.kAPI, cOpts}
	return nil
}

func (svc *service) GetCluster(name string) Cluster {
	return svc.cluster[name]
}

func (svc *service) ListCluster() (<-chan Cluster, chan<- struct{}) {
	out, stop := make(chan Cluster), make(chan struct{})
	go func() {
		defer close(out)
		for _, cluster := range svc.cluster {
			select {
			case out <- cluster:
				break
			case <-stop:
				return
			}
		}
	}()
	return out, stop
}

func newAWS() AutoScaling {
	dsc := etcd.NewKeysAPI(disc.NewDiscovery())
	auto := autoscaling.New(session.New())
	svc := &service{make(map[string]Cluster), auto, dsc}
	go func() {
		root, err := svc.kAPI.Get(ctx.Background(), ClusterGroup, nil)
		if err != nil {
			return // TODO: I should continue to track nodes joining
		}
		for _, node := range root.Node.Nodes {
			cOpts := &ClusterOptions{
				Root: ClusterGroup,
				Name: path.Base(node.Key),
			}
			params := &autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: []*string{aws.String(cOpts.Name)},
				MaxRecords:            aws.Int64(1),
			}
			resp, err := svc.auto.DescribeAutoScalingGroups(params)
			if err == nil {
				group := resp.AutoScalingGroups[0]
				cOpts.Min = *group.MinSize
				cOpts.Max = *group.MaxSize
				cOpts.Count = int64(len(group.Instances))
				svc.cluster[cOpts.Name] = &cluster{svc.auto, svc.kAPI, cOpts}
			} else {
				fmt.Println(err.Error())
			}
		}
	}()
	return svc
}
