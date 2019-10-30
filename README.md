# Kubernetes

[![GoDoc Widget]][GoDoc] [![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/569/badge)](https://bestpractices.coreinfrastructure.org/projects/569)

<img src="https://github.com/kubernetes/kubernetes/raw/master/logo/logo.png" width="100">

----

Kubernetes is an open source system for managing [containerized applications]
across multiple hosts; providing basic mechanisms for deployment, maintenance,
and scaling of applications.

Kubernetes builds upon a decade and a half of experience at Google running
production workloads at scale using a system called [Borg],
combined with best-of-breed ideas and practices from the community.

Kubernetes is hosted by the Cloud Native Computing Foundation ([CNCF]).
If you are a company that wants to help shape the evolution of
technologies that are container-packaged, dynamically-scheduled
and microservices-oriented, consider joining the CNCF.
For details about who's involved and how Kubernetes plays a role,
read the CNCF [announcement].

----

## To start using Kubernetes

See our documentation on [kubernetes.io].

Try our [interactive tutorial].

Take a free course on [Scalable Microservices with Kubernetes].

## To start developing Kubernetes

The [community repository] hosts all information about
building Kubernetes from source, how to contribute code
and documentation, who to contact about what, etc.

If you want to build Kubernetes right away there are two options:

##### You have a working [Go environment].

```
mkdir -p $GOPATH/src/k8s.io
cd $GOPATH/src/k8s.io
git clone https://github.com/kubernetes/kubernetes
cd kubernetes
make
```

##### You have a working [Docker environment].

```
git clone https://github.com/kubernetes/kubernetes
cd kubernetes
make quick-release
```

For the full story, head over to the [developer's documentation].

## Support

If you need support, start with the [troubleshooting guide],
and work your way through the process that we've outlined.

That said, if you have questions, reach out to us
[one way or another][communication].

[announcement]: https://cncf.io/news/announcement/2015/07/new-cloud-native-computing-foundation-drive-alignment-among-container
[Borg]: https://research.google.com/pubs/pub43438.html
[CNCF]: https://www.cncf.io/about
[communication]: https://git.k8s.io/community/communication
[community repository]: https://git.k8s.io/community
[containerized applications]: https://kubernetes.io/docs/concepts/overview/what-is-kubernetes/
[developer's documentation]: https://git.k8s.io/community/contributors/devel#readme
[Docker environment]: https://docs.docker.com/engine
[Go environment]: https://golang.org/doc/install
[GoDoc]: https://godoc.org/k8s.io/kubernetes
[GoDoc Widget]: https://godoc.org/k8s.io/kubernetes?status.svg
[interactive tutorial]: https://kubernetes.io/docs/tutorials/kubernetes-basics
[kubernetes.io]: https://kubernetes.io
[Scalable Microservices with Kubernetes]: https://www.udacity.com/course/scalable-microservices-with-kubernetes--ud615
[troubleshooting guide]: https://kubernetes.io/docs/tasks/debug-application-cluster/troubleshooting/

[![Analytics](https://kubernetes-site.appspot.com/UA-36037335-10/GitHub/README.md?pixel)]()

----

## Add CPU Pinning feature in kubernetes v1.16

This feature works with Topology Manager component in order to prioritize a NUMA node for CPU pinning.

### Files modified

Modified files:
 * README.md   (this file)
 * pkg/kubelet/cm/cpumanager/cpu_manager.go
 * pkg/kubelet/cm/cpumanager/topology_hints.go

### Get the sources

Sources (complete Kubernetes and modifications) are a available in the Orange Forge repository:
 * URL: https://gitlab.forge.orange-labs.fr/telco-k8s
 * project: k8s-cpuman
 * Tags for the numa cpu pinning:
    * v1.9.3-6_numa
    * v1.9.5-6_numa
    * v1.12.3-6_numa
 * Branches for the numa cpu pinning:
    * telco_policy   (for 1.9.x)
    * telco_policy_1.12   (for 1.12.3)
    * telco_policy_1.16   (for 1.16)

### Execute unit tests

Go to your kubernetes directory and use this command to execute all the test files in the specified directory (here only the cpu manager unit tests):
```
$ make check WHAT=./pkg/kubelet/cm/cpumanager GOFLAGS=-v
```


### Compile and build Kubernetes (and the new kubelet service)

Tested on a VMWare ESXi VM with CentOS 7.7 (64 bits).

To build kubernetes binary with the same characteristics as your building environment, go to your kubernetes directory and use this command :

```sh
$ make quick-release
```

To build only one kubernetes binary component with the same characteristics as your building environment, go to your kubernetes directory and use this command :

```sh
$ make WHAT=cmd/kubelet
```

Archives and binaries are stored in kubernetes's the _output directory. Thus, binaries are in kubernetes/_output/local/bin/linux/amd64


If you need to build binaries for more OS and processors, use this command:
```sh
$ make release
```
It's a very long process. You need more than 12 Gb of RAM for parallel compilation (faster).

Binaries are in _output directory.


### Deploy new kubelet service on an existing cluster (cluster deployed with kubeadm)

#### Manual procedure

On each node (minion):
 1. Backup kubelet binary /usr/bin/kubelet
 2. Exclude the node from the cluster: from client side with kubectl, use:
``` sh
$ kubectl drain node_name --ignore-daemonsets      (other option: --delete-local-data)
```
 3. Stop kubelet service:
``` sh
$ sudo systemctl stop kubelet
```
 4. Remove cpu_manager_state file
 5. Replace /usr/bin/kubelet file with the numa version generated
 6. In /var/lib/kubelet/config.yaml file, change cpuManagerPolicy parameter to "static", active topologymanager and reserve resources (for cpumanager) by appending lines :
 ``` yaml
 featureGates:
  TopologyManager: true
 kubeReserved:
  cpu: 500m
 ```
 7. Start kubelet service:
``` sh
$ sudo systemctl start kubelet
```
 8. Reactivate the node in the cluster and check node status: from client side with kubectl, use:
``` sh
$ kubectl uncordon node_name
$ kubectl get nodes
```
The node version must be "v1.16" and his status "Ready".

#### Check NUMA Node works

 1. Getting Container ID (first column) on the node by running :
``` sh
$ docker ps
```
 2. Getting in which processor the process is assigned to :
 ``` sh
 $ ps -o pid,psr,cmd $(docker inspect --format '{{.State.Pid}}' ContainerID)
 ```

#### Ansible procedure

A "kubeletdeploy" tool is available in the Orange Forge git:
 * Project: **k8s-kubespray**
 * Directory: **tools/kubeletdeploy**
 * URL: https://gitlab.forge.orange-labs.fr/telco-k8s/k8s-kubespray/tree/master/tools/kubeletdeploy

A README is available.

This tool use Ansible to deploy the kubelet binary and config files to one or more nodes (without draining the nodes).
**Warning**: the kubelet config file is defined in a jinja template **./templates/kubelet.env.node.j2**

Actually this template is a "standard" template for K8s kubelet v1.9.5. You need to change this template for your need and version used on your cluster.



### How to use numa in a pod?

You must add an annotation "PreferredNUMANodeId" specifying the CPU index (or NUMA node) on which you want the pod to run and use class of service Guaranteed. Index starts at 0.

For example:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: exclusive-2-s1
  annotations:
    PreferredNUMANodeId: "1"
spec:
  nodeName: node3
  containers:
  - image: quay.io/connordoyle/cpuset-visualizer
    name: exclusive-2-s1
    resources:
      limits:
        cpu: 2
        memory: "256M"
      requests:
        cpu: 2
        memory: "256M"
```
This pod try to execute the container on the 2nd CPU (N° 1) on the node3, on 2 reserved "logical CPUs" (ie one physical core with hyper threading)
