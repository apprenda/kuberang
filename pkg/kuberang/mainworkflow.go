package kuberang

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"errors"

	"github.com/apprenda/kuberang/pkg/config"
)

const RunPrefix = "kuberang-"
const BBDeploymentName = RunPrefix + "busybox"
const NGDeploymentName = RunPrefix + "nginx"
const Timeout = 300 //seconds
const HTTP_Timeout = 1000 * time.Millisecond

func CheckKubernetes(outputFormat string) error {
	var report report
	switch outputFormat {
	case "json":
		report = &simpleReport{}
	case "simple":
		report = &echoReport{
			out: os.Stdout,
		}
	default:
		return fmt.Errorf("%s is not supported as an output format", outputFormat)
	}

	ngServiceName := nginxServiceName()

	if !PrecheckKubectl(report) ||
		!PrecheckNamespace(report) ||
		!PrecheckServices(report, ngServiceName) ||
		!PrecheckDeployments(report) {
		PowerDown(report, ngServiceName)
		if outputFormat == "json" {
			printJSONReport(report)
		}
		return errors.New("Pre-conditions failed; must clean up before we can smoke test")
	}

	// Scale out busybox
	busyboxCount := int64(1)
	if ko := RunKubectl("run", BBDeploymentName, "--image=busybox", "--image-pull-policy=IfNotPresent", "--", "sleep", "3600"); !ko.Success {
		report.addError("Issued BusyBox start request")
	} else {
		report.addSuccess("Issued BusyBox start request")
	}
	// Scale out nginx
	// Try to run a Pod on each Node,
	// This scheduling is not guaranteed but it gets close
	nginxCount := int64(RunGetNodes().NodeCount())
	if ko := RunPod(NGDeploymentName, "nginx", nginxCount); !ko.Success {
		report.addError("Issued Nginx start request")
	} else {
		report.addSuccess("Issued Nginx start request")
	}

	// Check for both
	if !WaitForDeployments(report, busyboxCount, nginxCount) {
		return nil
	}

	// Add service
	if ko := RunKubectl("expose", "deployment", NGDeploymentName, "--name="+ngServiceName, "--port=80"); !ko.Success {
		report.addError("Issued expose Nginx service request")
	} else {
		report.addSuccess("Issued expose Nginx service request")
	}

	// Get pod & service IPs
	var podIPs []string
	var serviceIP string
	var busyboxPodName string
	if ko := RunKubectl("get", "pods", "-l", "run=kuberang-nginx", "-o", "json"); ko.Success {
		podIPs = ko.PodIPs()
		report.addSuccess("Grab nginx pod ip addresses")
	} else {
		podIPs = make([]string, 0)
		report.addError("Grab nginx pod ip addresses")
	}

	if ko := RunGetService(ngServiceName); ko.Success {
		serviceIP = ko.ServiceCluserIP()
		report.addSuccess("Grab nginx service ip address")
	} else {
		serviceIP = ""
		report.addError("Grab nginx service ip address")
	}

	if ko := RunKubectl("get", "pods", "-l", "run=kuberang-busybox", "-o", "json"); ko.Success {
		busyboxPodName = ko.FirstPodName()
		report.addSuccess("Grab BusyBox pod name")
	} else {
		busyboxPodName = ""
		report.addError("Grab BusyBox pod name")
	}

	// Check connectivity between pods (using busybox)
	if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", serviceIP); busyboxPodName == "" || ko.Success {
		report.addSuccess("Accessed Nginx service at " + serviceIP + " from BusyBox")
	} else {
		report.addError("Accessed Nginx service at " + serviceIP + " from BusyBox")
	}
	if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", ngServiceName); busyboxPodName == "" || ko.Success {
		report.addSuccess("Accessed Nginx service via DNS " + ngServiceName + " from BusyBox")
	} else {
		report.addError("Accessed Nginx service via DNS " + ngServiceName + " from BusyBox")
	}

	for _, podIP := range podIPs {
		if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", podIP); busyboxPodName == "" || ko.Success {
			report.addSuccess("Accessed Nginx pod at " + podIP + " from BusyBox")
		} else {
			report.addError("Accessed Nginx pod at " + podIP + " from BusyBox")
		}
	}

	// Check connectivity with internet
	if ko := RunKubectl("exec", busyboxPodName, "--", "wget", "-qO-", "Google.com"); busyboxPodName == "" || ko.Success {
		report.addSuccess("Accessed Google.com from BusyBox")
	} else {
		report.addIgnored("Accessed Google.com from BusyBox")
	}

	// Check connectivity from current machine (using curl or wget)
	// Set Timeout or it could wait forever
	client := http.Client{
		Timeout: HTTP_Timeout,
	}
	if _, err := client.Get("http://" + ngServiceName); err == nil {
		report.addSuccess("Accessed Nginx service via DNS " + ngServiceName + " from this node")
	} else {
		report.addIgnored("Accessed Nginx service via DNS " + ngServiceName + " from this node")
	}
	for _, podIP := range podIPs {
		if _, err := client.Get("http://" + podIP); err == nil {
			report.addSuccess("Accessed Nginx pod at " + podIP + " from this node")
		} else {
			report.addIgnored("Accessed Nginx pod at " + podIP + " from this node")
		}
	}
	if _, err := client.Get("http://google.com/"); err == nil {
		report.addSuccess("Accessed Google.com from this node")
	} else {
		report.addIgnored("Accessed Google.com from this node")
	}

	PowerDown(report, ngServiceName)

	if outputFormat == "json" {
		if err := printJSONReport(report); err != nil {
			return err
		}
	}

	if report.isSuccess() {
		return nil
	}
	return errors.New("One or more required steps failed")
}

func PrecheckKubectl(report report) bool {
	if ko := RunKubectl("version"); !ko.Success {
		report.addError("Configured kubectl exists")
		return false
	}
	report.addSuccess("Configured kubectl exists")
	return true
}

func PrecheckServices(report report, nginxServiceName string) bool {
	if ko := RunGetService(nginxServiceName); ko.Success {
		report.addError("Nginx service does not already exist")
		return false
	}
	report.addSuccess("Nginx service does not already exist")
	return true
}

func PrecheckDeployments(report report) bool {
	ret := true
	if ko := RunGetDeployment(BBDeploymentName); ko.Success {
		report.addError("BusyBox service does not already exist")
		ret = false
	} else {
		report.addSuccess("BusyBox service does not already exist")
	}
	if ko := RunGetDeployment(NGDeploymentName); ko.Success {
		report.addError("Nginx service does not already exist")
		ret = false
	} else {
		report.addSuccess("Nginx service does not already exist")
	}
	return ret
}

func PrecheckNamespace(report report) bool {
	ret := true
	if config.Namespace != "" {
		if ko := RunGetNamespace(config.Namespace); !ko.Success || ko.NamespaceStatus() != "Active" {
			report.addError("Configured kubernetes namespace `" + config.Namespace + "` exists")
			ret = false
		} else {
			report.addError("Configured kubernetes namespace `" + config.Namespace + "` exists")
		}
	}
	return ret
}

func CheckDeployments(busbyboxCount, nginxCount int64) bool {
	ret := true
	if ko := RunGetDeployment(BBDeploymentName); !ko.Success || ko.ObservedReplicaCount() != busbyboxCount {
		ret = false
	}
	if ko := RunGetDeployment(NGDeploymentName); !ko.Success || ko.ObservedReplicaCount() != nginxCount {
		ret = false
	}
	return ret
}

func WaitForDeployments(report report, busbyboxCount, nginxCount int64) bool {
	for i := 0; i < Timeout; i++ {
		if CheckDeployments(busbyboxCount, nginxCount) {
			report.addSuccess("Both deployments completed successfully within timeout")
			return true
		}
		time.Sleep(1 * time.Second)
	}
	report.addError("Both deployments completed successfully within timeout")
	return false
}

func PowerDown(report report, nginxServiceName string) {
	// Power down service
	if ko := RunKubectl("delete", "service", nginxServiceName); ko.Success {
		report.addSuccess("Powered down Nginx service")
	} else {
		report.addError("Powered down Nginx service")
	}
	// Power down bb
	if ko := RunKubectl("delete", "deployments", BBDeploymentName); ko.Success {
		report.addSuccess("Powered down Busybox deployment")
	} else {
		report.addError("Powered down Busybox deployment")
	}
	// Power down nginx
	if ko := RunKubectl("delete", "deployments", NGDeploymentName); ko.Success {
		report.addSuccess("Powered down Nginx deployment")
	} else {
		report.addError("Powered down Nginx deployment")
	}
}

func nginxServiceName() string {
	return fmt.Sprintf("%s-%d", RunPrefix+"nginx", time.Now().UnixNano())
}

func printJSONReport(report report) error {
	r := report.(*simpleReport)
	b, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		return fmt.Errorf("error printing JSON report: %v", err)
	}
	fmt.Printf("%s\n", string(b))
	return nil
}
