package cli

import (
	"context"
	"fmt"

	"github.com/openshift-pipelines/manual-approval-gate/pkg/client/clientset/versioned"
	userv1typedclient "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	k8s "k8s.io/client-go/kubernetes"
	authenticationv1client "k8s.io/client-go/kubernetes/typed/authentication/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Clients struct {
	Config       *rest.Config
	Kube         k8s.Interface
	Dynamic      dynamic.Interface
	ApprovalTask versioned.Interface
}

type ApprovalTaskParams struct {
	clients        *Clients
	kubeConfigPath string
	kubeContext    string
	namespace      string
}

type Options struct {
	Namespace     string
	Name          string
	Input         string
	Username      string
	Message       string
	AllNamespaces bool
	Groups        []string
}

type Params interface {
	// SetKubeConfigPath uses the kubeconfig path to instantiate tekton
	// returned by Clientset function
	SetKubeConfigPath(string)
	// SetKubeContext extends the specificity of the above SetKubeConfigPath
	// by using a context other than the default context in the given kubeconfig
	SetKubeContext(string)
	SetNamespace(string)
	KubeClient() (k8s.Interface, error)
	Clients(...*rest.Config) (*Clients, error)
	Namespace() string
	GetUserInfo() (string, []string, error)
}

// ensure that TektonParams complies with cli.Params interface
var _ Params = (*ApprovalTaskParams)(nil)

func (p *ApprovalTaskParams) SetKubeConfigPath(path string) {
	p.kubeConfigPath = path
}

func (p *ApprovalTaskParams) SetKubeContext(context string) {
	p.kubeContext = context
}

func (p *ApprovalTaskParams) Namespace() string {
	return p.namespace
}

func (p *ApprovalTaskParams) GetUserInfo() (string, []string, error) {
	authV1Client, err := authenticationv1client.NewForConfig(p.clients.Config)
	if err != nil {
		return "", []string{}, err
	}

	userInterface, err := userv1typedclient.NewForConfig(p.clients.Config)
	if err != nil {
		return "", []string{}, err
	}

	// Get username
	username, groups, err := getUserInfo(authV1Client, userInterface)
	if err != nil {
		return "", []string{}, err
	}
	return username, groups, err
}

func (p *ApprovalTaskParams) SetNamespace(ns string) {
	p.namespace = ns
}

func (p *ApprovalTaskParams) kubeClient(config *rest.Config) (k8s.Interface, error) {
	k8scs, err := k8s.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create k8s client from config")
	}

	return k8scs, nil
}

func (p *ApprovalTaskParams) dynamicClient(config *rest.Config) (dynamic.Interface, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create dynamic client from config")

	}
	return dynamicClient, err
}

func (p *ApprovalTaskParams) approvalTaskClient(config *rest.Config) (versioned.Interface, error) {
	approvalClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create dynamic client from config")

	}
	return approvalClient, err
}

func (p *ApprovalTaskParams) config() (*rest.Config, error) {

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if p.kubeConfigPath != "" {
		loadingRules.ExplicitPath = p.kubeConfigPath
	}
	configOverrides := &clientcmd.ConfigOverrides{}
	if p.kubeContext != "" {
		configOverrides.CurrentContext = p.kubeContext
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	if p.namespace == "" {
		namespace, _, err := kubeConfig.Namespace()
		if err != nil {
			return nil, errors.Wrap(err, "Couldn't get kubeConfiguration namespace")
		}
		p.namespace = namespace
	}
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Parsing kubeconfig failed")
	}

	// set values as done in kubectl
	config.QPS = 50.0
	config.Burst = 300

	return config, nil
}

// Only returns kube client, not tekton client
func (p *ApprovalTaskParams) KubeClient() (k8s.Interface, error) {
	config, err := p.config()
	if err != nil {
		return nil, err
	}

	kube, err := p.kubeClient(config)
	if err != nil {
		return nil, err
	}

	return kube, nil
}

func (p *ApprovalTaskParams) Clients(cfg ...*rest.Config) (*Clients, error) {
	var config *rest.Config

	if len(cfg) != 0 && cfg[0] != nil {
		config = cfg[0]
	} else {
		defaultConfig, err := p.config()
		if err != nil {
			return nil, err
		}
		config = defaultConfig
	}

	kube, err := p.kubeClient(config)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := p.dynamicClient(config)
	if err != nil {
		return nil, err
	}

	approvalClient, err := p.approvalTaskClient(config)
	if err != nil {
		return nil, err
	}

	p.clients = &Clients{
		Config:       config,
		Kube:         kube,
		ApprovalTask: approvalClient,
		Dynamic:      dynamicClient,
	}

	return p.clients, nil
}

func getUserInfo(authV1Client *authenticationv1client.AuthenticationV1Client, userInterface userv1typedclient.UserV1Interface) (string, []string, error) {
	var username string
	res, err := authV1Client.SelfSubjectReviews().Create(context.TODO(), &v1.SelfSubjectReview{}, metav1.CreateOptions{})
	if err == nil {
		username = res.Status.UserInfo.Username
		return username, res.Status.UserInfo.Groups, nil
	} else {
		fmt.Errorf("selfsubjectreview request error %v, falling back to user object", err)
	}

	user, err := userInterface.Users().Get(context.TODO(), "~", metav1.GetOptions{})
	if err != nil {
		return "", []string{}, nil
	}
	username = user.Name

	return username, user.Groups, nil
}
