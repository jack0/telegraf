package kube_inventory

import (
	"context"
	"strings"

	v1 "github.com/ericchiang/k8s/apis/core/v1"

	"github.com/influxdata/telegraf"
)

func collectNodes(ctx context.Context, acc telegraf.Accumulator, ki *KubernetesInventory) {
	list, err := ki.client.getNodes(ctx)
	if err != nil {
		acc.AddError(err)
		return
	}
	for _, n := range list.Items {
		if err = ki.gatherNode(*n, acc); err != nil {
			acc.AddError(err)
			return
		}
	}
}

func (ki *KubernetesInventory) gatherNode(n v1.Node, acc telegraf.Accumulator) error {
	fields := map[string]interface{}{}
	tags := map[string]string{
		"node_name": *n.Metadata.Name,
	}

	for resourceName, val := range n.Status.Capacity {
		switch resourceName {
		case "cpu":
			fields["capacity_cpu_cores"] = atoi(val.GetString_())
		case "memory":
			fields["capacity_memory_bytes"] = convertQuantity(val.GetString_(), 1)
		case "pods":
			fields["capacity_pods"] = atoi(val.GetString_())
		}
	}

	for resourceName, val := range n.Status.Allocatable {
		switch resourceName {
		case "cpu":
			fields["allocatable_cpu_cores"] = atoi(val.GetString_())
		case "memory":
			fields["allocatable_memory_bytes"] = convertQuantity(val.GetString_(), 1)
		case "pods":
			fields["allocatable_pods"] = atoi(val.GetString_())
		}
	}

	acc.AddFields(nodeMeasurement, fields, tags)

	for _, val := range n.Status.Conditions {
		fields := map[string]interface{}{
			"condition": val.GetType(),
		}

		tags["status"] = strings.ToLower(val.GetStatus())

		if info := n.Status.NodeInfo; fields["condition"] == "Ready" && info != nil {
			tags["machine_id"] = info.GetMachineID()
			tags["architecture"] = info.GetArchitecture()
			tags["boot_id"] = info.GetBootID()
			tags["container_runtime_version"] = info.GetContainerRuntimeVersion()
			tags["kernel_version"] = info.GetKernelVersion()
			tags["kubelet_version"] = info.GetKubeletVersion()
			tags["kube_proxy_version"] = info.GetKubeProxyVersion()
			tags["os_image"] = info.GetOsImage()
			tags["os"] = info.GetOperatingSystem()
			tags["system_uuid"] = info.GetSystemUUID()

			if n.Spec != nil {
				tags["pod_cidr"] = n.Spec.GetPodCIDR()
				tags["provider_id"] = n.Spec.GetProviderID()
			}
		}

		acc.AddFields(nodeMeasurement, fields, tags)
	}

	return nil
}
