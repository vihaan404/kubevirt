package topology

import (
	"fmt"

	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/pointer"

	k6tv1 "kubevirt.io/client-go/api/v1"
)

type Hinter interface {
	TopologyHintsForVMI(vmi *k6tv1.VirtualMachineInstance) (hints *k6tv1.TopologyHints, err error)
}

type topologyHinter struct {
	nodeStore cache.Store
	arch      string
}

func (t *topologyHinter) TopologyHintsForVMI(vmi *k6tv1.VirtualMachineInstance) (hints *k6tv1.TopologyHints, err error) {
	if VMIHasInvTSCFeature(vmi) && t.arch == "amd64" {
		freq, err := t.LowestTSCFrequencyOnCluster()
		if err != nil {
			return nil, fmt.Errorf("failed to determine the lowest tsc frequency on the cluster: %v", err)
		}
		return &k6tv1.TopologyHints{
			TSCFrequency: pointer.Int64Ptr(freq),
		}, nil
	}
	return nil, nil
}

func (t *topologyHinter) LowestTSCFrequencyOnCluster() (int64, error) {
	nodes := FilterNodesFromCache(t.nodeStore.List(),
		HasInvTSCFrequency,
	)
	freq := LowestTSCFrequency(nodes)
	if freq == 0 {
		return 0, fmt.Errorf("no schedulable node exposes a tsc-frequency")
	}
	return freq, nil
}

func NewTopologyHinter(nodeStore cache.Store, arch string) *topologyHinter {
	return &topologyHinter{nodeStore: nodeStore, arch: arch}
}
