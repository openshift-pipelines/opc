package pipelineascode

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/openshift-pipelines/pipelines-as-code/pkg/action"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/apis/pipelinesascode/keys"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/apis/pipelinesascode/v1alpha1"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/formatting"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/opscomments"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params/triggertype"
)

type matchingCond func(pr tektonv1.PipelineRun) bool

var cancelMergePatch = map[string]interface{}{
	"spec": map[string]interface{}{
		"status": tektonv1.PipelineRunSpecStatusCancelledRunFinally,
	},
}

// cancelAllInProgressBelongingToClosedPullRequest cancels all in-progress PipelineRuns
// that belong to a specific pull request in the given repository.
func (p *PacRun) cancelAllInProgressBelongingToClosedPullRequest(ctx context.Context, repo *v1alpha1.Repository) error {
	labelSelector := getLabelSelector(map[string]string{
		keys.URLRepository:    formatting.CleanValueKubernetes(p.event.Repository),
		keys.PullRequest:      strconv.Itoa(p.event.PullRequestNumber),
		keys.CancelInProgress: "true",
	})
	prs, err := p.run.Clients.Tekton.TektonV1().PipelineRuns(repo.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("failed to list pipelineRuns : %w", err)
	}

	if len(prs.Items) == 0 {
		msg := fmt.Sprintf("no pipelinerun found for repository: %v and pullRequest %v",
			p.event.Repository, p.event.PullRequestNumber)
		p.eventEmitter.EmitMessage(repo, zap.InfoLevel, "CancelInProgress", msg)
		return nil
	}

	p.cancelPipelineRuns(ctx, prs, repo, func(_ tektonv1.PipelineRun) bool {
		return true
	})

	return nil
}

// cancelInProgressMatchingPR cancels all PipelineRuns associated with a given repository and pull request,
// except for the one that triggered the cancellation. It first checks if the cancellation is in progress
// and if the repository has a concurrency limit. If a concurrency limit is set, it returns an error as
// cancellation is not supported with concurrency limits. It then retrieves the original pull request name
// from the annotations and lists all PipelineRuns with matching labels. For each PipelineRun that is not
// already done, cancelled, or gracefully stopped, it patches the PipelineRun to cancel it.
func (p *PacRun) cancelInProgressMatchingPR(ctx context.Context, matchPR *tektonv1.PipelineRun, repo *v1alpha1.Repository) error {
	if matchPR == nil {
		return nil
	}
	if key, ok := matchPR.GetAnnotations()[keys.CancelInProgress]; !ok || key != "true" {
		return nil
	}

	// As PipelineRuns are filtered by name, OriginalPRName should be taken from
	// labels instead of annotations because of constraints imposed by kube API.
	prName, ok := matchPR.GetLabels()[keys.OriginalPRName]
	if !ok {
		return nil
	}

	if repo.Spec.ConcurrencyLimit != nil && *repo.Spec.ConcurrencyLimit > 0 {
		return fmt.Errorf("cancel in progress is not supported with concurrency limit")
	}

	labelMap := map[string]string{
		keys.URLRepository:  formatting.CleanValueKubernetes(p.event.Repository),
		keys.OriginalPRName: prName,
	}
	if p.event.TriggerTarget == triggertype.PullRequest {
		labelMap[keys.PullRequest] = strconv.Itoa(p.event.PullRequestNumber)
	}
	labelSelector := getLabelSelector(labelMap)
	if p.event.TriggerTarget == triggertype.PullRequest {
		labelSelector += fmt.Sprintf(",%s in (pull_request, %s)", keys.EventType, opscomments.AnyOpsKubeLabelInSelector())
	}
	p.run.Clients.Log.Infof("cancel-in-progress: selecting pipelineRuns to cancel with labels: %v", labelSelector)
	prs, err := p.run.Clients.Tekton.TektonV1().PipelineRuns(matchPR.GetNamespace()).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("failed to list pipelineRuns : %w", err)
	}

	p.cancelPipelineRuns(ctx, prs, repo, func(pr tektonv1.PipelineRun) bool {
		// skip our own for cancellation
		if sourceBranch, ok := pr.GetAnnotations()[keys.SourceBranch]; ok {
			// NOTE(chmouel): Every PR has their own branch and so is every push to different branch
			// it means we only cancel pipelinerun of the same name that runs to
			// the unique branch. Note: HeadBranch is the branch from where the PR
			// comes from in git jargon.
			if sourceBranch != p.event.HeadBranch {
				p.logger.Infof("cancel-in-progress: skipping pipelinerun %v/%v as it is not from the same branch, annotation source-branch: %s event headbranch: %s", pr.GetNamespace(), pr.GetName(), sourceBranch, p.event.HeadBranch)
				return false
			}
		}

		return pr.GetName() != matchPR.GetName()
	})
	return nil
}

// cancelPipelineRunsOpsComment cancels all PipelineRuns associated with a given repository and pull request.
// when the user issue a cancel comment.
func (p *PacRun) cancelPipelineRunsOpsComment(ctx context.Context, repo *v1alpha1.Repository) error {
	labelSelector := getLabelSelector(map[string]string{
		keys.URLRepository: formatting.CleanValueKubernetes(p.event.Repository),
		keys.SHA:           formatting.CleanValueKubernetes(p.event.SHA),
	})

	if p.event.TriggerTarget == triggertype.PullRequest {
		labelSelector = getLabelSelector(map[string]string{
			keys.PullRequest: strconv.Itoa(p.event.PullRequestNumber),
		})
	}

	prs, err := p.run.Clients.Tekton.TektonV1().PipelineRuns(repo.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return fmt.Errorf("failed to list pipelineRuns : %w", err)
	}

	if len(prs.Items) == 0 {
		msg := fmt.Sprintf("no pipelinerun found for repository: %v , sha: %v and pulRequest %v",
			p.event.Repository, p.event.SHA, p.event.PullRequestNumber)
		p.eventEmitter.EmitMessage(repo, zap.InfoLevel, "CancelInProgress", msg)
		return nil
	}

	p.cancelPipelineRuns(ctx, prs, repo, func(pr tektonv1.PipelineRun) bool {
		if p.event.TargetCancelPipelineRun != "" {
			if prName, ok := pr.GetAnnotations()[keys.OriginalPRName]; !ok || prName != p.event.TargetCancelPipelineRun {
				return false
			}
		}
		return true
	})

	return nil
}

func (p *PacRun) cancelPipelineRuns(ctx context.Context, prs *tektonv1.PipelineRunList, repo *v1alpha1.Repository, condition matchingCond) {
	var wg sync.WaitGroup
	for _, pr := range prs.Items {
		if !condition(pr) {
			continue
		}

		if pr.IsCancelled() || pr.IsGracefullyCancelled() || pr.IsGracefullyStopped() {
			p.logger.Infof("cancel-in-progress: skipping cancelling pipelinerun %v/%v, already in %v state", pr.GetNamespace(), pr.GetName(), pr.Spec.Status)
			continue
		}

		if pr.IsDone() {
			p.logger.Infof("cancel-in-progress: skipping cancelling pipelinerun %v/%v, already done", pr.GetNamespace(), pr.GetName())
			continue
		}

		p.logger.Infof("cancel-in-progress: cancelling pipelinerun %v/%v", pr.GetNamespace(), pr.GetName())
		wg.Add(1)
		go func(ctx context.Context, pr tektonv1.PipelineRun) {
			defer wg.Done()
			if _, err := action.PatchPipelineRun(ctx, p.logger, "cancel patch", p.run.Clients.Tekton, &pr, cancelMergePatch); err != nil {
				errMsg := fmt.Sprintf("failed to cancel pipelineRun %s/%s: %s", pr.GetNamespace(), pr.GetName(), err.Error())
				p.eventEmitter.EmitMessage(repo, zap.ErrorLevel, "CancelInProgress", errMsg)
			}
		}(ctx, pr)
	}
	wg.Wait()
}

func getLabelSelector(labelsMap map[string]string) string {
	labelSelector := labels.NewSelector()
	for k, v := range labelsMap {
		req, _ := labels.NewRequirement(k, selection.Equals, []string{v})
		if req != nil {
			labelSelector = labelSelector.Add(*req)
		}
	}
	return labelSelector.String()
}
