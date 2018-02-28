package kuberang

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/apprenda/kuberang/pkg/config"
)

type KubeOutput struct {
	Success     bool
	CombinedOut string
	RawOut      []byte
}

func RunKubectl(args ...string) KubeOutput {
	if config.Kubeconfig != "" {
		args = append([]string{"--kubeconfig=" + config.Kubeconfig}, args...)
	}

	if config.Namespace != "" {
		args = append([]string{"--namespace=" + config.Namespace}, args...)
	}

	kubeCmd := exec.Command("kubectl", args...)
	bytes, err := kubeCmd.CombinedOutput()
	if err != nil {
		return KubeOutput{
			Success:     false,
			CombinedOut: string(bytes),
			RawOut:      bytes,
		}
	}
	return KubeOutput{
		Success:     true,
		CombinedOut: string(bytes),
		RawOut:      bytes,
	}
}

func RunGetService(svcName string) KubeOutput {
	return RunKubectl("get", "service", svcName, "-o", "json")
}

func RunGetDeployment(name string) KubeOutput {
	return RunKubectl("get", "deployment", name, "-o", "json")
}

func RunGetNamespace(name string) KubeOutput {
	return RunKubectl("get", "namespace", name, "-o", "json")
}

func RunGetNodes() KubeOutput {
	return RunKubectl("get", "nodes", "-o", "json")
}

func (ko KubeOutput) ObservedReplicaCount() int64 {
	resp := DeploymentResponse{}
	json.Unmarshal(ko.RawOut, &resp)
	return resp.Status.AvaiableReplicas
}

type DeploymentResponse struct {
	Status struct {
		AvaiableReplicas int64 `json:"availableReplicas"`
	} `json:"status"`
}

func (ko KubeOutput) ServiceCluserIP() string {
	resp := ServiceResponse{}
	json.Unmarshal(ko.RawOut, &resp)
	return resp.Spec.ClusterIP
}

type ServiceResponse struct {
	Spec struct {
		ClusterIP string `json:"clusterIP"`
	} `json:"spec"`
}

func (ko KubeOutput) PodIPs() []string {
	//In Scala, this code would be gorgeous. In Golang, it's a blood blister
	resp := PodsResponse{}
	if err := json.Unmarshal(ko.RawOut, &resp); err != nil {
		fmt.Println(err)
	}
	podIPs := make([]string, len(resp.Items))
	for i, item := range resp.Items {
		podIPs[i] = item.Status.PodIP
	}
	return podIPs
}

func (ko KubeOutput) FirstPodName() string {
	resp := PodsResponse{}
	if err := json.Unmarshal(ko.RawOut, &resp); err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(ko.RawOut, &resp)
	if len(resp.Items) < 1 {
		return ""
	}
	return resp.Items[0].Metadata.Name
}

type PodsResponse struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Status struct {
			PodIP string `json:"podIP"`
		} `json:"status"`
	} `json:"items"`
}

type NodeResponse struct {
	Items []struct {
		Spec struct {
			Unschedulable bool `json:"unschedulable,omitempty"`
		} `json:"spec"`
	} `json:"items"`
}

func (ko KubeOutput) NodeCount() int {
	resp := NodeResponse{}
	json.Unmarshal(ko.RawOut, &resp)
	count := 0
	for _, item := range resp.Items {
		if item.Spec.Unschedulable == false {
			count++
		}
	}
	return count
}

func (ko KubeOutput) NamespaceStatus() string {
	resp := NamespaceResponse{}
	json.Unmarshal(ko.RawOut, &resp)
	return resp.Status.Phase
}

type NamespaceResponse struct {
	Status struct {
		Phase string `json:"phase"`
	} `json:"status"`
}
