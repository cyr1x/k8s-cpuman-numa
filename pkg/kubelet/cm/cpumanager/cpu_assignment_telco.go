/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cpumanager

import (
	"fmt"
	"sort"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/topology"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

// Returns free socket IDs as a slice sorted by:
// - socket ID in parameter, if he's free.
// - other socket ID, ascending.
func (a *cpuAccumulator) freeSocketsTelco( socketIdPref int ) []int {

	sockets := a.details.Sockets().Filter(a.isSocketFree).ToSlice()

	if socketIdPref == -1 {
		//identical to the static policy: free socket id by ascending order
		return sockets
	}

	if !a.isSocketFree( socketIdPref ) {
		//preferential socket isn't free, we can return the details table
		return sockets
	}

	// socketIdPref must be the first id, followed by other free socket id
	var freeSock []int
	freeSock = append( freeSock, socketIdPref )   //first id in the result
	for _,s := range sockets {
		if s != socketIdPref {
			freeSock = append( freeSock, s )
		}
	}
	return freeSock
}

// Returns CPU IDs of prefered socket as a slice sorted by:
// - number of CPUs available on this socket
// - number of CPUs available on the same core
// - socket ID.
// - core ID.
// socketIdPref must be >= 0
func (a *cpuAccumulator) freeCPUsTelco( socketIdPref int ) []int {
	result := []int{}

	if socketIdPref >= 0 {
		//get only cores in the prefered socket
		cores := a.details.CoresInSocket(socketIdPref).ToSlice()

		//keep the same sort algorithm as static policy, but only with the cores of the prefered socket
		sort.Slice(
			cores,
			func(i, j int) bool {
				iCore := cores[i]
				jCore := cores[j]

				iCPUs := a.topo.CPUDetails.CPUsInCore(iCore).ToSlice()
				jCPUs := a.topo.CPUDetails.CPUsInCore(jCore).ToSlice()

				iSocket := a.topo.CPUDetails[iCPUs[0]].SocketID
				jSocket := a.topo.CPUDetails[jCPUs[0]].SocketID

				// Compute the number of CPUs in the result reside on the same socket
				// as each core.
				iSocketColoScore := a.topo.CPUDetails.CPUsInSocket(iSocket).Intersection(a.result).Size()
				jSocketColoScore := a.topo.CPUDetails.CPUsInSocket(jSocket).Intersection(a.result).Size()

				// Compute the number of available CPUs available on the same socket
				// as each core.
				iSocketFreeScore := a.details.CPUsInSocket(iSocket).Size()
				jSocketFreeScore := a.details.CPUsInSocket(jSocket).Size()

				// Compute the number of available CPUs on each core.
				iCoreFreeScore := a.details.CPUsInCore(iCore).Size()
				jCoreFreeScore := a.details.CPUsInCore(jCore).Size()

				return iSocketColoScore > jSocketColoScore ||
					iSocketFreeScore < jSocketFreeScore ||
					iCoreFreeScore < jCoreFreeScore ||
					iSocket < jSocket ||
					iCore < jCore
			})

		// For each core, append sorted CPU IDs to result.
		for _, core := range cores {
			result = append(result, a.details.CPUsInCore(core).ToSlice()...)
		}
	}
	return result
}


//---------------------------------------------------------------------------
// Choose logical cpu(s) for a "guaranteed" pod
//
// cpuPref: socket ID for the preferred cpu (-1 if there is no preference)
//
func takeByTopologyTelco(topo *topology.CPUTopology, availableCPUs cpuset.CPUSet, numCPUs int, socketIdPref int) (cpuset.CPUSet, error) {
	acc := newCPUAccumulator(topo, availableCPUs, numCPUs)
	if acc.isSatisfied() {
		return acc.result, nil
	}
	if acc.isFailed() {
		return cpuset.NewCPUSet(), fmt.Errorf("not enough cpus available to satisfy request")
	}

	// Algorithm: topology-aware best-fit but with socket priority if socketIdPref >= 0

	// 1. Acquire whole sockets, if available and the container requires at
	//    least a socket's-worth of CPUs.
	for _, s := range acc.freeSocketsTelco( socketIdPref ) {
		if acc.needs(acc.topo.CPUsPerSocket()) {
			//glog.V(4).Infof("[cpumanager] takeByTopologyTelco: claiming socket [%d]", s)
			var msgpref = "non prefered"
			if s == socketIdPref {
				msgpref = "prefered"
			}
			glog.Infof("[cpumanager] takeByTopologyTelco: claiming %s socket [%d]", msgpref, s)
			acc.take(acc.details.CPUsInSocket(s))
			if acc.isSatisfied() {
				return acc.result, nil
			}
		}
	}

	// 2. Acquire whole cores on preferential socket
	if socketIdPref >= 0 {

		//TODO: remove this debug messages
		glog.Infof("[cpumanager] takeByTopologyTelco: [%d] free cores on socket [%d]", acc.details.CoresInSocket(socketIdPref).Filter(acc.isCoreFree).Size(), socketIdPref)
		glog.Infof("[cpumanager] takeByTopologyTelco - Debug: cores in socket [%d] = %+v", socketIdPref, acc.details.CoresInSocket(socketIdPref) );
		glog.Infof("[cpumanager] takeByTopologyTelco - Debug: free cores in socket [%d] = %+v", socketIdPref, acc.details.CoresInSocket(socketIdPref).Filter(acc.isCoreFree) );

		//Range free cores on this socket
		for _, c := range acc.details.CoresInSocket(socketIdPref).Filter(acc.isCoreFree).ToSlice() {
			if acc.needs(acc.topo.CPUsPerCore()) {
				//glog.V(4).Infof("[cpumanager] takeByTopologyTelco: claiming core [%d] on socket [%d]", c, socketIdPref)
				glog.Infof("[cpumanager] takeByTopologyTelco: claiming core [%d] on prefered socket [%d]", c, socketIdPref)
				acc.take(acc.details.CPUsInCore(c))
				if acc.isSatisfied() {
					return acc.result, nil
				}
			}
		}
	}

	// 3. Acquire whole cores, if no more core available on prefered socket and
	//    if available and the container requires at least a core's-worth of CPUs.
	for _, c := range acc.freeCores() {
		if acc.needs(acc.topo.CPUsPerCore()) {
			//glog.V(4).Infof("[cpumanager] takeByTopologyTelco: claiming core [%d]", c)
			glog.Infof("[cpumanager] takeByTopologyTelco: claiming core [%d] on non prefered socket", c)
			acc.take(acc.details.CPUsInCore(c))
			if acc.isSatisfied() {
				return acc.result, nil
			}
		}
	}

	// 4. Acquire single threads, preferring to fill partially-allocated cores
	//    on the same sockets as the whole cores we have already taken in this
	//    allocation. Priority to the prefered socket
	//Range free cpu (process unit) on this socket
	for _, c := range acc.freeCPUsTelco(socketIdPref) {
		if acc.needs(1) {
			//glog.V(4).Infof("[cpumanager] takeByTopologyTelco: claiming CPU [%d] on socket [%d]", c, socketIdPref)
			glog.V(4).Infof("[cpumanager] takeByTopologyTelco: claiming CPU [%d] on prefered socket [%d]", c, socketIdPref)
			acc.take(cpuset.NewCPUSet(c))
		}
		if acc.isSatisfied() {
			return acc.result, nil
		}
	}

	// 5. Acquire single threads on another sockets, preferring to fill partially-allocated cores
	//    on the same sockets as the whole cores we have already taken in this
	//    allocation.
	for _, c := range acc.freeCPUs() {
		if acc.needs(1) {
			//glog.V(4).Infof("[cpumanager] takeByTopologyTelco: claiming CPU [%d]", c)
			glog.Infof("[cpumanager] takeByTopologyTelco: claiming CPU [%d] on non prefered socket", c)
			acc.take(cpuset.NewCPUSet(c))
		}
		if acc.isSatisfied() {
			return acc.result, nil
		}
	}

	return cpuset.NewCPUSet(), fmt.Errorf("failed to allocate cpus")
}
