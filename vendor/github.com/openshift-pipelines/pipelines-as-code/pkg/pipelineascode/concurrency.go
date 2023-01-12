package pipelineascode

import (
	"fmt"
	"sync"

	"github.com/openshift-pipelines/pipelines-as-code/pkg/sort"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

const namePath = "{.metadata.name}"

type ConcurrencyManager struct {
	enabled      bool
	pipelineRuns []*v1beta1.PipelineRun
	mutex        *sync.Mutex
}

func NewConcurrencyManager() *ConcurrencyManager {
	return &ConcurrencyManager{
		pipelineRuns: []*v1beta1.PipelineRun{},
		mutex:        &sync.Mutex{},
	}
}

func (c *ConcurrencyManager) AddPipelineRun(pr *v1beta1.PipelineRun) {
	if !c.enabled {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.pipelineRuns = append(c.pipelineRuns, pr)
}

func (c *ConcurrencyManager) Enable() {
	c.enabled = true
}

func (c *ConcurrencyManager) GetExecutionOrder() (string, []*v1beta1.PipelineRun) {
	if !c.enabled {
		return "", nil
	}

	runtimeObjs := []runtime.Object{}
	for _, pr := range c.pipelineRuns {
		runtimeObjs = append(runtimeObjs, pr)
	}

	// sort runs by name
	sort.ByField(namePath, runtimeObjs)

	sortedPipelineRuns := []*v1beta1.PipelineRun{}
	for _, run := range runtimeObjs {
		pr, _ := run.(*v1beta1.PipelineRun)
		sortedPipelineRuns = append(sortedPipelineRuns, pr)
	}
	c.pipelineRuns = sortedPipelineRuns

	return getOrderByName(c.pipelineRuns), c.pipelineRuns
}

func getOrderByName(runs []*v1beta1.PipelineRun) string {
	var order string
	for _, run := range runs {
		if order == "" {
			order = fmt.Sprintf("%s/%s", run.GetNamespace(), run.GetName())
			continue
		}
		order = order + "," + fmt.Sprintf("%s/%s", run.GetNamespace(), run.GetName())
	}
	return order
}
