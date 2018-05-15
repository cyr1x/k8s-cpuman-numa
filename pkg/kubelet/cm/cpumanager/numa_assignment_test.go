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
	"reflect"
	"testing"

	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/topology"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)


func TestCPUAccumulatorFreeSocketsTelco(t *testing.T) {
	testCases := []struct {
		description   string
		topo          *topology.CPUTopology
		availableCPUs cpuset.CPUSet
		socketIdPref  int
		expect        []int
	}{
		{
			"single socket HT, 1 socket free, no pref socket",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			-1,
			[]int{0},
		},
		{
			"single socket HT, 0 sockets free, no pref socket",
			topoSingleSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 6, 7),
			-1,
			[]int{},
		},
		{
			"single socket HT, 1 socket free, pref socket 0",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			0,
			[]int{0},
		},
		{
			"single socket HT, 0 sockets free, pref socket 0",
			topoSingleSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 6, 7),
			0,
			[]int{},
		},
		{
			"dual socket HT, 2 sockets free, no pref socket",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			-1,
			[]int{0, 1},
		},
		{
			"dual socket HT, 1 socket free, no pref socket",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11),
			-1,
			[]int{1},
		},
		{
			"dual socket HT, 0 sockets free, no pref socket",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 2, 3, 4, 5, 6, 7, 8, 9, 11),
			-1,
			[]int{},
		},
		{
			"dual socket HT, 2 sockets free, pref socket 0",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			0,
			[]int{0, 1},
		},
		{
			"dual socket HT, 1 socket free, pref socket 0",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11),
			0,
			[]int{1},
		},
		{
			"dual socket HT, 0 sockets free, pref socket 0",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 2, 3, 4, 5, 6, 7, 8, 9, 11),
			0,
			[]int{},
		},
		{
			"dual socket HT, 2 sockets free, pref socket 1",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			1,
			[]int{1, 0},
		},
		{
			"dual socket HT, 1 socket free, pref socket 1",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11),
			1,
			[]int{1},
		},
		{
			"dual socket HT, 0 sockets free, pref socket 1",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 2, 3, 4, 5, 6, 7, 8, 9, 11),
			1,
			[]int{},
		},
	}

	for _, tc := range testCases {
		acc := newCPUAccumulator(tc.topo, tc.availableCPUs, 0)
		result := acc.freeSocketsTelco(tc.socketIdPref)
		if !reflect.DeepEqual(result, tc.expect) {
			t.Errorf("[%s] expected %v to equal %v", tc.description, result, tc.expect)
		}
	}
}



func TestCPUAccumulatorFreeCPUsTelco(t *testing.T) {
	testCases := []struct {
		description   string
		topo          *topology.CPUTopology
		availableCPUs cpuset.CPUSet
		socketIdPref  int
		expect        []int
	}{
		{
			"single socket HT, 8 cpus free, no pref socket",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			-1,
			[]int{0, 4, 1, 5, 2, 6, 3, 7},
		},
		{
			"single socket HT, 5 cpus free, no pref socket",
			topoSingleSocketHT,
			cpuset.NewCPUSet(3, 4, 5, 6, 7),
			-1,
			[]int{4, 5, 6, 3, 7},
		},
		{
			"single socket HT, 8 cpus free, pref socket 0",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			0,
			[]int{0, 4, 1, 5, 2, 6, 3, 7},
		},
		{
			"single socket HT, 5 cpus free, pref socket 0",
			topoSingleSocketHT,
			cpuset.NewCPUSet(3, 4, 5, 6, 7),
			0,
			[]int{4, 5, 6, 3, 7},
		},
		{
			"dual socket HT, 12 cpus free, no pref socket",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			-1,
			[]int{0, 6, 2, 8, 4, 10, 1, 7, 3, 9, 5, 11},
		},
		{
			"dual socket HT, 11 cpus free, no pref socket",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			-1,
			[]int{6, 2, 8, 4, 10, 1, 7, 3, 9, 5, 11},
		},
		{
			"dual socket HT, 10 cpus free, no pref socket",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 7, 8, 9, 10, 11),
			-1,
			[]int{2, 8, 4, 10, 1, 7, 3, 9, 5, 11},
		},
		{
			"dual socket HT, 12 cpus free, pref socket 0",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			0,
			[]int{0, 6, 2, 8, 4, 10, 1, 7, 3, 9, 5, 11},
		},
		{
			"dual socket HT, 11 cpus free, no pref socket 0",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			0,
			[]int{6, 2, 8, 4, 10, 1, 7, 3, 9, 5, 11},
		},
		{
			"dual socket HT, 10 cpus free, pref socket 0",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 7, 8, 9, 10, 11),
			0,
			[]int{2, 8, 4, 10, 1, 7, 3, 9, 5, 11},
		},
		{
			"dual socket HT, 12 cpus free, pref socket 1",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			1,
			[]int{1, 7, 3, 9, 5, 11, 0, 6, 2, 8, 4, 10},
		},
		{
			"dual socket HT, 11 cpus free, no pref socket 1",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			1,
			[]int{1, 7, 3, 9, 5, 11, 6, 2, 8, 4, 10},
		},
		{
			"dual socket HT, 10 cpus free, pref socket 1",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 7, 8, 9, 10, 11),
			1,
			[]int{1, 7, 3, 9, 5, 11, 2, 8, 4, 10},
		},
		{
			"dual socket HT, 10 cpus free, pref socket 1",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 2, 3, 4, 5, 6, 8, 9, 10, 11),
			1,
			[]int{3, 9, 5, 11, 0, 6, 2, 8, 4, 10},
		},
	}

	for _, tc := range testCases {
		acc := newCPUAccumulator(tc.topo, tc.availableCPUs, 0)
		result := acc.freeCPUsTelco(tc.socketIdPref)
		if !reflect.DeepEqual(result, tc.expect) {
			t.Errorf("[%s] expected %v to equal %v", tc.description, result, tc.expect)
		}
	}
}

func TestTakeByTopologyTelco(t *testing.T) {
	testCases := []struct {
		description   string
		topo          *topology.CPUTopology
		availableCPUs cpuset.CPUSet
		numCPUs       int
		socketIdPref  int
		expErr        string
		expResult     cpuset.CPUSet
	}{
		{
			"take more cpus than are available from single socket with HT",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 2, 4, 6),
			5,
			-1,
			"not enough cpus available to satisfy request",
			cpuset.NewCPUSet(),
		},
		{
			"take zero cpus from single socket with HT",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			0,
			-1,
			"",
			cpuset.NewCPUSet(),
		},
		{
			"take one cpu from single socket with HT",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			1,
			-1,
			"",
			cpuset.NewCPUSet(0),
		},
		{
			"take one cpu from single socket with HT, some cpus are taken",
			topoSingleSocketHT,
			cpuset.NewCPUSet(1, 3, 5, 6, 7),
			1,
			-1,
			"",
			cpuset.NewCPUSet(6),
		},
		{
			"take two cpus from single socket with HT",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			2,
			-1,
			"",
			cpuset.NewCPUSet(0, 4),
		},
		{
			"take all cpus from single socket with HT",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			8,
			-1,
			"",
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
		},
		{
			"take two cpus from single socket with HT, only one core totally free",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 6),
			2,
			-1,
			"",
			cpuset.NewCPUSet(2, 6),
		},
		{
			"take 1 cpu from dual socket with HT - core from Socket 0",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 7, 8, 9, 10, 11),
			1,
			-1,
			"",
			cpuset.NewCPUSet(2),
		},
		{
			"take a socket of cpus from dual socket with HT",
			topoDualSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			6,
			-1,
			"",
			cpuset.NewCPUSet(0, 2, 4, 6, 8, 10),
		},
		{
			"take one cpu from single socket with HT, with prefered socket 0",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			1,
			0,
			"",
			cpuset.NewCPUSet(0),
		},
		{
			"take one cpu from single socket with HT, some cpus are taken, with prefered socket 0",
			topoSingleSocketHT,
			cpuset.NewCPUSet(1, 3, 5, 6, 7),
			1,
			0,
			"",
			cpuset.NewCPUSet(6),
		},
		{
			"take two cpus from single socket with HT, with prefered socket 0",
			topoSingleSocketHT,
			cpuset.NewCPUSet(0, 1, 2, 3, 4, 5, 6, 7),
			2,
			0,
			"",
			cpuset.NewCPUSet(0, 4),
		},
		{
			"take 1 cpu from dual socket with HT - priority to Socket 0",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 7, 8, 9, 10, 11),
			1,
			0,
			"",
			cpuset.NewCPUSet(2),
		},
		{
			"take 1 cpu from dual socket with HT - priority to Socket 1",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 7, 8, 9, 10, 11),
			1,
			0,
			"",
			cpuset.NewCPUSet(1),
		},
		{
			"take 2 cpus from dual socket with HT - priority to Socket 0",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 7, 8, 9, 10, 11),
			2,
			0,
			"",
			cpuset.NewCPUSet(2, 8),    //first free core on socket 0
		},
		{
			"take 2 cpus from dual socket with HT - priority to Socket 1",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 7, 8, 9, 10, 11),
			2,
			1,
			"",
			cpuset.NewCPUSet(1, 7),    //first free core on socket 1
		},
		{
			"take 2 cpus from dual socket with HT, some cpus are taken - priority to Socket 0",
			topoDualSocketHT,
			cpuset.NewCPUSet(3, 4, 5, 7, 8, 9, 10, 11),
			2,
			0,
			"",
			cpuset.NewCPUSet(4, 10),    //first free core on socket 0
		},
		{
			"take 2 cpus from dual socket with HT, some cpus are taken - priority to Socket 1",
			topoDualSocketHT,
			cpuset.NewCPUSet(3, 4, 5, 7, 8, 9, 10, 11),
			2,
			1,
			"",
			cpuset.NewCPUSet(3, 9),    //first free core on socket 1
		},
		{
			"take 3 cpus from dual socket with HT - priority to Socket 0",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			3,
			0,
			"",
			cpuset.NewCPUSet(2, 8, 4),    
		},
		{
			"take 3 cpus from dual socket with HT - priority to Socket 1",
			topoDualSocketHT,
			cpuset.NewCPUSet(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			3,
			0,
			"",
			cpuset.NewCPUSet(1, 7, 3),    
		},

	}

	for _, tc := range testCases {
		result, err := takeByTopologyTelco(tc.topo, tc.availableCPUs, tc.numCPUs, tc.socketIdPref)
		if tc.expErr != "" && err.Error() != tc.expErr {
			t.Errorf("expected error to be [%v] but it was [%v] in test \"%s\"", tc.expErr, err, tc.description)
		}
		if !result.Equals(tc.expResult) {
			t.Errorf("expected result [%s] to equal [%s] in test \"%s\"", result, tc.expResult, tc.description)
		}
	}
}
