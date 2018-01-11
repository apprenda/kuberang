#Kuberang Helm Charts

kuberang is a command-line utility for smoke testing a Kubernetes install.

It scales out a pair of services, checks their connectivity, and then scales back in again.

Here, kuberang will be run as CronJob, set to run every 30 minutes.

To retrieve the logs, we need to determine the pod the job was executed on:

```
KUBERANG_POD=$(kubectl get pods -a --selector=type=kuberang-logging --output=jsonpath={.items..metadata.name})
```
Note: add the `-n <namespace>` option if you are running kuberang in a namespace.

(The 'type' label was added to the job definition in order to allow this filtering).

Then ```kubectl logs $KUBERANG_POD -n <namespace>``` will produce the logs, e.g:

```
Kubectl configured on this node                                                 [OK]
Delete existing deployments if they exist                                       [OK]
Nginx service does not already exist                                            [OK]
BusyBox service does not already exist                                          [OK]
Nginx service does not already exist                                            [OK]
Issued BusyBox start request                                                    [OK]
Issued Nginx start request                                                      [OK]
Issued expose Nginx service request                                             [OK]
Both deployments completed successfully within timeout                          [OK]
Grab nginx pod ip addresses                                                     [OK]
Grab nginx service ip address                                                   [OK]
Grab BusyBox pod name                                                           [OK]
Accessed Nginx service at 192.168.139.151 from BusyBox                          [OK]
Accessed Nginx service via DNS kuberang-nginx-1515421803094247236 from BusyBox  [OK]
Accessed Nginx pod at 192.168.103.241 from BusyBox                              [OK]
Accessed Nginx pod at 192.168.53.185 from BusyBox                               [OK]
Accessed Nginx pod at 192.168.56.170 from BusyBox                               [OK]
Accessed Nginx pod at 192.168.103.240 from BusyBox                              [OK]
Accessed Nginx pod at 192.168.53.184 from BusyBox                               [OK]
Accessed Nginx pod at 192.168.56.169 from BusyBox                               [OK]
Accessed Google.com from BusyBox                                                [ERROR IGNORED]
Accessed Nginx pod at 192.168.103.241 from this node                            [OK]
Accessed Nginx pod at 192.168.53.185 from this node                             [OK]
Accessed Nginx pod at 192.168.56.170 from this node                             [OK]
Accessed Nginx pod at 192.168.103.240 from this node                            [OK]
Accessed Nginx pod at 192.168.53.184 from this node                             [OK]
Accessed Nginx pod at 192.168.56.169 from this node                             [OK]
Accessed Google.com from this node                                              [ERROR IGNORED]
Powered down Nginx service                                                      [OK]
Powered down Busybox deployment                                                 [OK]
Powered down Nginx deployment                                                   [OK]

```