package config

var (
	// Kubeconfig is the path to the kubeconfig file
	Kubeconfig string
	// Namespace where the kuberang tests will be executed
	Namespace string
	// RegistryURL to be used for downloading the container images used in the smoke test
	RegistryURL string
	// SkipCleanup determines whether the workloads should be cleaned up after the test
	SkipCleanup bool
	// SkipDNSTests determines whether the DNS tests should be performed
	SkipDNSTests bool
	// IgnorePodIPAccessibilityCheck determines whether a failed pod IP accessibility check
	// should fail the smoke test as a whole
	IgnorePodIPAccessibilityCheck bool
)
