package actions

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/openshift-pipelines/manual-approval-gate/pkg/apis/approvaltask/v1alpha1"
	"github.com/openshift-pipelines/manual-approval-gate/pkg/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

var (
	doOnce      sync.Once
	apiGroupRes []*restmapper.APIGroupResources
)

// List fetches the resource and convert it to respective object
func List(gr schema.GroupVersionResource, c *cli.Clients, opts metav1.ListOptions, ns string, obj interface{}) error {
	unstructuredObj, err := list(gr, c.Dynamic, c.ApprovalTask.Discovery(), ns, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list objects from %s namespace \n", ns)
		return err
	}

	return runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.UnstructuredContent(), obj)
}

// list takes a partial resource and fetches a list of that resource's objects in the cluster using the dynamic client.
func list(gr schema.GroupVersionResource, dynamic dynamic.Interface, discovery discovery.DiscoveryInterface, ns string, op metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	gvr, err := GetGroupVersionResource(gr, discovery)
	if err != nil {
		return nil, err
	}

	allRes, err := dynamic.Resource(*gvr).Namespace(ns).List(context.Background(), op)
	if err != nil {
		return nil, err
	}

	return allRes, nil
}

func Get(gr schema.GroupVersionResource, c *cli.Clients, opts *cli.Options) (*v1alpha1.ApprovalTask, error) {
	gvr, err := GetGroupVersionResource(gr, c.ApprovalTask.Discovery())
	if err != nil {
		return nil, err
	}

	at, err := get(gvr, c, opts)
	if err != nil {
		return &v1alpha1.ApprovalTask{}, err
	}

	return at, nil
}

func get(gvr *schema.GroupVersionResource, c *cli.Clients, opts *cli.Options) (*v1alpha1.ApprovalTask, error) {
	result, err := c.Dynamic.Resource(*gvr).Namespace(opts.Namespace).Get(context.Background(), opts.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	at := &v1alpha1.ApprovalTask{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(result.Object, at)
	if err != nil {
		return nil, err
	}

	return at, nil
}

func Update(gr schema.GroupVersionResource, c *cli.Clients, opts *cli.Options) error {
	gvr, err := GetGroupVersionResource(gr, c.ApprovalTask.Discovery())
	if err != nil {
		return err
	}

	at, err := get(gvr, c, opts)
	if err != nil {
		return err
	}

	if !containsUsername(at.Spec.Approvers, opts) {
		return fmt.Errorf("Approver: %s, is not present in the approvers list", opts.Username)
	}

	if err := update(gvr, c.Dynamic, at, opts); err != nil {
		return err
	}

	return nil
}

func update(gvr *schema.GroupVersionResource, dynamic dynamic.Interface, at *v1alpha1.ApprovalTask, opts *cli.Options) error {
	for i, approver := range at.Spec.Approvers {
		switch v1alpha1.DefaultedApproverType(approver.Type) {
		case "User":
			if approver.Name == opts.Username {
				// return true
				at.Spec.Approvers[i].Input = opts.Input
				if opts.Message != "" {
					at.Spec.Approvers[i].Message = opts.Message
				}
			}
		case "Group":
			for _, groupName := range opts.Groups {
				if approver.Name == groupName {
					// return true
					at.Spec.Approvers[i].Input = opts.Input
					if opts.Message != "" {
						at.Spec.Approvers[i].Message = opts.Message
					}

					userExists := false

					for j, existing := range at.Spec.Approvers[i].Users {
						if existing.Name == opts.Username {
							userExists = true
							if existing.Input != opts.Input {
								at.Spec.Approvers[i].Users[j].Input = opts.Input
							}
							break
						}
					}
					if !userExists {
						newUser := v1alpha1.UserDetails{
							Name:  opts.Username,
							Input: opts.Input,
						}
						at.Spec.Approvers[i].Users = append(at.Spec.Approvers[i].Users, newUser)
					}
				}
			}
		}

		if approver.Name == opts.Username {
			at.Spec.Approvers[i].Input = opts.Input
			if opts.Message != "" {
				at.Spec.Approvers[i].Message = opts.Message
			}
		}
	}

	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&at)
	if err != nil {
		fmt.Printf("Error converting to unstructured: %v\n", err)
		return err
	}

	unstrObj := &unstructured.Unstructured{Object: unstructuredMap}
	_, err = dynamic.Resource(*gvr).Namespace(opts.Namespace).Update(context.TODO(), unstrObj, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func GetGroupVersionResource(gr schema.GroupVersionResource, discovery discovery.DiscoveryInterface) (*schema.GroupVersionResource, error) {
	var err error
	doOnce.Do(func() {
		err = InitializeAPIGroupRes(discovery)
	})
	if err != nil {
		return nil, err
	}

	rm := restmapper.NewDiscoveryRESTMapper(apiGroupRes)
	gvr, err := rm.ResourceFor(gr)
	if err != nil {
		return nil, err
	}

	return &gvr, nil
}

// InitializeAPIGroupRes initializes and populates the discovery client.
func InitializeAPIGroupRes(discovery discovery.DiscoveryInterface) error {
	var err error
	apiGroupRes, err = restmapper.GetAPIGroupResources(discovery)
	if err != nil {
		return err
	}
	return nil
}

func containsUsername(approvers []v1alpha1.ApproverDetails, user *cli.Options) bool {
	for _, approver := range approvers {
		if approver.Name == user.Username {
			return true
		}
	}

	for _, approval := range approvers {
		switch approval.Type {
		case "User":
			if approval.Name == user.Username {
				return true
			}
		case "Group":
			for _, groupName := range user.Groups {
				if approval.Name == groupName {
					return true
				}
			}
		}
	}
	return false
}
