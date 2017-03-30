package kuberang

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"errors"

	"github.com/apprenda/kuberang/pkg/config"
	"github.com/apprenda/kuberang/pkg/util"
)

const (
	runPrefix         = "kuberang-"
	bbDeploymentName  = runPrefix + "busybox"
	ngDeploymentName  = runPrefix + "nginx"
	deploymentTimeout = 300 * time.Second
	httpTimeout       = 1000 * time.Millisecond
)

// CheckKubernetes runs checks against a cluster. It expects to find
// a configured `kubectl` binary in the path.
func CheckKubernetes(skipCleanup bool) error {
	out := os.Stdout
	ngServiceName := nginxServiceName()
	success := true
	registryURL := ""
	if config.RegistryURL != "" {
		registryURL = config.RegistryURL + "/"
	}

	// If kubectl doesn't exist, don't bother doing anything
	if !precheckKubectl() {
		return errors.New("Kubectl must be configured on this machine before running kuberang")
	}
	util.PrettyPrintOk(os.Stdout, "Kubectl configured on this node")

	// Ensure any pre-existing kuberang deployments are cleaned up
	if err := removeExisting(ngServiceName); err != nil {
		return err
	}

	// Make sure we have all we need
	// Quit if we find existing kuberang deployments on the cluster
	if !checkPreconditions(ngServiceName) {
		return errors.New("Pre-conditions failed")
	}

	if !skipCleanup {
		defer powerDown(ngServiceName)
	}

	// Deploy the workloads required for running checks
	if !deployTestWorkloads(registryURL, out, ngServiceName) {
		return errors.New("Failed to deploy test workloads")
	}

	// Get IPs of all nginx pods
	podIPs := []string{}
	var ko KubeOutput
	ok := retry(3, func() bool {
		if ko = RunKubectl("get", "pods", "-l", "run=kuberang-nginx", "-o", "json"); ko.Success {
			podIPs = ko.PodIPs()
			// check for at least one pod IP
			if len(podIPs) == 0 {
				return false
			}
			// make sure no IPs are blank
			for _, podIP := range podIPs {
				if podIP == "" {
					return false
				}
			}
			return true
		}
		return false
	})
	if ok {
		util.PrettyPrintOk(out, "Grab nginx pod ip addresses")
	} else {
		util.PrettyPrintErr(out, "Grab nginx pod ip addresses")
		printFailureDetail(out, ko.CombinedOut)
		success = false
	}

	// Get the service IP of the nginx service
	var serviceIP string
	ok = retry(3, func() bool {
		if ko = RunGetService(ngServiceName); ko.Success {
			serviceIP = ko.ServiceCluserIP()
			if serviceIP != "" {
				return true
			}
		}
		return false
	})
	if ok {
		util.PrettyPrintOk(out, "Grab nginx service ip address")
	} else {
		util.PrettyPrintErr(out, "Grab nginx service ip address")
		printFailureDetail(out, ko.CombinedOut)
		success = false
	}

	// Get the name of the busybox pod
	var busyboxPodName string
	ok = retry(3, func() bool {
		if ko = RunKubectl("get", "pods", "-l", "run=kuberang-busybox", "-o", "json"); ko.Success {
			busyboxPodName = ko.FirstPodName()
			if busyboxPodName != "" {
				return true
			}
		}
		return false
	})
	if ok {
		util.PrettyPrintOk(out, "Grab BusyBox pod name")
	} else {
		util.PrettyPrintErr(out, "Grab BusyBox pod name")
		printFailureDetail(out, ko.CombinedOut)
		success = false
	}

	// Gate on successful acquisition of all the required names / IPs
	if !success {
		return errors.New("Failed to get required information from cluster")
	}

	// The following checks verify the pod network and the ability for
	// pods to talk to each other.
	// 1. Access nginx service via service IP from another pod
	var kubeOut KubeOutput
	ok = retry(3, func() bool {
		kubeOut = RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", serviceIP)
		return kubeOut.Success
	})
	if ok {
		util.PrettyPrintOk(out, "Accessed Nginx service at "+serviceIP+" from BusyBox")
	} else {
		printFailureDetail(out, kubeOut.CombinedOut)
		util.PrettyPrintErr(out, "Accessed Nginx service at "+serviceIP+" from BusyBox")
		success = false
	}

	// 2. Access nginx service via service name (DNS) from another pod
	ok = retry(6, func() bool {
		kubeOut = RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", ngServiceName)
		return kubeOut.Success
	})
	if ok {
		util.PrettyPrintOk(out, "Accessed Nginx service via DNS "+ngServiceName+" from BusyBox")
	} else {
		util.PrettyPrintErr(out, "Accessed Nginx service via DNS "+ngServiceName+" from BusyBox")
		printFailureDetail(out, kubeOut.CombinedOut)
		success = false
	}

	// 3. Access all nginx pods by IP
	for _, podIP := range podIPs {
		ok = retry(3, func() bool {
			kubeOut = RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", podIP)
			return kubeOut.Success
		})
		if ok {
			util.PrettyPrintOk(out, "Accessed Nginx pod at "+podIP+" from BusyBox")
		} else {
			util.PrettyPrintErr(out, "Accessed Nginx pod at "+podIP+" from BusyBox")
			printFailureDetail(out, kubeOut.CombinedOut)
			success = false
		}
	}

	// 4. Check internet connectivity from pod
	if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", "Google.com"); busyboxPodName == "" || ko.Success {
		util.PrettyPrintOk(out, "Accessed Google.com from BusyBox")
	} else {
		util.PrettyPrintErrorIgnored(out, "Accessed Google.com from BusyBox")
	}

	client := http.Client{
		Timeout: httpTimeout,
	}
	// 5. Check connectivity from current machine to all nginx pods
	for _, podIP := range podIPs {
		if _, err := client.Get("http://" + podIP); err == nil {
			util.PrettyPrintOk(out, "Accessed Nginx pod at "+podIP+" from this node")
		} else {
			util.PrettyPrintErrorIgnored(out, "Accessed Nginx pod at "+podIP+" from this node")
		}
	}

	// 6. Check internet connectivity from current machine
	if _, err := client.Get("http://google.com/"); err == nil {
		util.PrettyPrintOk(out, "Accessed Google.com from this node")
	} else {
		util.PrettyPrintErrorIgnored(out, "Accessed Google.com from this node")
	}

	// 7. Verify that the busybox pod is able to ping an API server via the kubernetes service
	ok = retry(3, func() bool {
		kubeOut = RunKubectl("exec", busyboxPodName, "--", "ping", "-c", "5", "kubernetes")
		return kubeOut.Success
	})
	if ok {
		util.PrettyPrintOk(out, "Ping kubernetes service from BusyBox")
	} else {
		util.PrettyPrintErr(out, "Ping kubernetes service from BusyBox")
		printFailureDetail(out, kubeOut.CombinedOut)
		success = false
	}

	if !success {
		return errors.New("One or more required steps failed")
	}
	return nil
}

func deployTestWorkloads(registryURL string, out io.Writer, ngServiceName string) bool {
	// Scale out busybox
	busyboxCount := int64(1)
	if ko := RunKubectl("run", bbDeploymentName, fmt.Sprintf("--image=%sbusybox:latest", registryURL), "--image-pull-policy=IfNotPresent", "--", "sleep", "3600"); !ko.Success {
		util.PrettyPrintErr(out, "Issued BusyBox start request")
		printFailureDetail(out, ko.CombinedOut)
		return false
	}
	util.PrettyPrintOk(out, "Issued BusyBox start request")

	// Scale out nginx
	// Try to run a Pod on each Node,
	// This scheduling is not guaranteed but it gets close
	nginxCount := int64(RunGetNodes().NodeCount())
	if ko := RunPod(ngDeploymentName, fmt.Sprintf("%snginx:stable-alpine", registryURL), nginxCount); !ko.Success {
		util.PrettyPrintErr(out, "Issued Nginx start request")
		printFailureDetail(out, ko.CombinedOut)
		return false
	}
	util.PrettyPrintOk(out, "Issued Nginx start request")

	// Add service
	if ko := RunKubectl("expose", "deployment", ngDeploymentName, "--name="+ngServiceName, "--port=80"); !ko.Success {
		util.PrettyPrintErr(out, "Issued expose Nginx service request")
		printFailureDetail(out, ko.CombinedOut)
		return false
	}
	util.PrettyPrintOk(out, "Issued expose Nginx service request")

	// Wait until deployments are ready
	return waitForDeployments(busyboxCount, nginxCount)
}

func checkPreconditions(nginxServiceName string) bool {
	ok := true
	if !precheckNamespace() {
		ok = false
	}
	if !precheckServices(nginxServiceName) {
		ok = false
	}
	if !precheckDeployments() {
		ok = false
	}
	return ok
}

func precheckKubectl() bool {
	if ko := RunKubectl("version"); !ko.Success {
		util.PrettyPrintErr(os.Stdout, "Configured kubectl exists")
		printFailureDetail(os.Stdout, ko.CombinedOut)
		return false
	}
	return true
}

func precheckServices(nginxServiceName string) bool {
	if ko := RunGetService(nginxServiceName); ko.Success {
		util.PrettyPrintErr(os.Stdout, "Nginx service does not already exist")
		printFailureDetail(os.Stdout, ko.CombinedOut)
		return false
	}
	util.PrettyPrintOk(os.Stdout, "Nginx service does not already exist")
	return true
}

func precheckDeployments() bool {
	ret := true
	if ko := RunGetDeployment(bbDeploymentName); ko.Success {
		util.PrettyPrintErr(os.Stdout, "BusyBox service does not already exist")
		printFailureDetail(os.Stdout, ko.CombinedOut)
		ret = false
	} else {
		util.PrettyPrintOk(os.Stdout, "BusyBox service does not already exist")
	}
	if ko := RunGetDeployment(ngDeploymentName); ko.Success {
		util.PrettyPrintErr(os.Stdout, "Nginx service does not already exist")
		printFailureDetail(os.Stdout, ko.CombinedOut)
		ret = false
	} else {
		util.PrettyPrintOk(os.Stdout, "Nginx service does not already exist")
	}
	return ret
}

func precheckNamespace() bool {
	ret := true
	if config.Namespace != "" {
		ko := RunGetNamespace(config.Namespace)
		if !ko.Success {
			util.PrettyPrintErr(os.Stdout, "Configured kubernetes namespace `"+config.Namespace+"` exists")
			printFailureDetail(os.Stdout, ko.CombinedOut)
			ret = false
		} else if ko.NamespaceStatus() != "Active" {
			util.PrettyPrintErr(os.Stdout, "Configured kubernetes namespace `"+config.Namespace+"` exists")
			ret = false
		} else {
			util.PrettyPrintOk(os.Stdout, "Configured kubernetes namespace `"+config.Namespace+"` exists")
		}
	}
	return ret
}

func checkDeployments(busyboxCount, nginxCount int64) bool {
	ret := true
	ko := RunGetDeployment(bbDeploymentName)
	if !ko.Success {
		ret = false
	} else if ko.ObservedReplicaCount() != busyboxCount {
		ret = false
	}
	ko = RunGetDeployment(ngDeploymentName)
	if !ko.Success {
		ret = false
	} else if ko.ObservedReplicaCount() != nginxCount {
		ret = false
	}
	return ret
}

func waitForDeployments(busyboxCount, nginxCount int64) bool {
	start := time.Now()
	for time.Since(start) < deploymentTimeout {
		if checkDeployments(busyboxCount, nginxCount) {
			util.PrettyPrintOk(os.Stdout, "Both deployments completed successfully within timeout")
			return true
		}
		time.Sleep(1 * time.Second)
	}
	util.PrettyPrintErr(os.Stdout, "Both deployments completed successfully within timeout")
	return false
}

func powerDown(nginxServiceName string) {
	// Power down service
	if ko := RunKubectl("delete", "service", nginxServiceName); ko.Success {
		util.PrettyPrintOk(os.Stdout, "Powered down Nginx service")
	} else {
		util.PrettyPrintErr(os.Stdout, "Powered down Nginx service")
		printFailureDetail(os.Stdout, ko.CombinedOut)
	}
	// Power down bb
	if ko := RunKubectl("delete", "deployments", bbDeploymentName); ko.Success {
		util.PrettyPrintOk(os.Stdout, "Powered down Busybox deployment")
	} else {
		util.PrettyPrintErr(os.Stdout, "Powered down Busybox deployment")
		printFailureDetail(os.Stdout, ko.CombinedOut)
	}
	// Power down nginx
	if ko := RunKubectl("delete", "deployments", ngDeploymentName); ko.Success {
		util.PrettyPrintOk(os.Stdout, "Powered down Nginx deployment")
	} else {
		util.PrettyPrintErr(os.Stdout, "Powered down Nginx deployment")
		printFailureDetail(os.Stdout, ko.CombinedOut)
	}
}

func removeExisting(nginxServiceName string) error {
	ko := RunKubectl("delete", "--ignore-not-found=true",
		fmt.Sprintf("deployment/%s", bbDeploymentName),
		fmt.Sprintf("deployment/%s", ngDeploymentName),
		fmt.Sprintf("service/%s", nginxServiceName),
	)
	if !ko.Success {
		util.PrettyPrintErr(os.Stdout, "Delete existing deployments if they exist")
		printFailureDetail(os.Stdout, ko.CombinedOut)
		return errors.New("Failure removing existing kuberang deployments")
	}
	util.PrettyPrintOk(os.Stdout, "Delete existing deployments if they exist")
	return nil
}

func nginxServiceName() string {
	return fmt.Sprintf("%s-%d", runPrefix+"nginx", time.Now().UnixNano())
}

func printFailureDetail(out io.Writer, detail string) {
	fmt.Fprintln(out, "-------- OUTPUT --------")
	fmt.Fprintf(out, detail)
	fmt.Fprintln(out, "------------------------")
	fmt.Fprintln(out)
}
