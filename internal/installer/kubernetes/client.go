package kubernetes

import (
	"context"
	"fmt"
	"strings"

	installerv1alpha1 "github.com/m8platform/platform/api/installer/v1alpha1"
	"github.com/m8platform/platform/internal/installer/preflight"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset kubernetes.Interface
	dynamic   dynamic.Interface
}

type ClientOptions struct {
	Kubeconfig string
	Context    string
}

func NewClient(options ClientOptions) (*Client, error) {
	config, err := RESTConfig(options)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create Kubernetes client: %w", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create Kubernetes dynamic client: %w", err)
	}
	return &Client{clientset: clientset, dynamic: dynamicClient}, nil
}

func RESTConfig(options ClientOptions) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if options.Kubeconfig != "" {
		loadingRules.ExplicitPath = options.Kubeconfig
	}
	overrides := &clientcmd.ConfigOverrides{}
	if options.Context != "" {
		overrides.CurrentContext = options.Context
	}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}
	return config, nil
}

func (c *Client) ServerVersion(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	version, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return "", fmt.Errorf("get Kubernetes server version: %w", err)
	}
	return version.GitVersion, nil
}

func (c *Client) NodeSummary(ctx context.Context) (preflight.NodeSummary, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return preflight.NodeSummary{}, fmt.Errorf("list nodes: %w", err)
	}
	summary := preflight.NodeSummary{
		Total:         len(nodes.Items),
		Architectures: map[string]int{},
		Zones:         map[string]int{},
	}
	for _, node := range nodes.Items {
		if isNodeReady(node) {
			summary.Ready++
		}
		if arch := node.Labels[corev1.LabelArchStable]; arch != "" {
			summary.Architectures[arch]++
		}
		if zone := node.Labels[corev1.LabelTopologyZone]; zone != "" {
			summary.Zones[zone]++
		}
	}
	return summary, nil
}

func (c *Client) HasAPIResource(ctx context.Context, groupVersion string, kind string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	resources, err := c.clientset.Discovery().ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("discover API resources for %s: %w", groupVersion, err)
	}
	for _, resource := range resources.APIResources {
		if resource.Kind == kind {
			return true, nil
		}
	}
	return false, nil
}

func (c *Client) StorageClasses(ctx context.Context) ([]string, error) {
	classes, err := c.clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list storage classes: %w", err)
	}
	names := make([]string, 0, len(classes.Items))
	for _, class := range classes.Items {
		names = append(names, class.Name)
	}
	return names, nil
}

var platformInstallationGVR = schema.GroupVersionResource{
	Group:    installerv1alpha1.GroupName,
	Version:  installerv1alpha1.Version,
	Resource: "platforminstallations",
}

func (c *Client) GetPlatformInstallation(ctx context.Context, namespace string, name string) (installerv1alpha1.PlatformInstallation, error) {
	if name == "" {
		return installerv1alpha1.PlatformInstallation{}, fmt.Errorf("platform installation name is required")
	}
	if namespace == "" {
		namespace = "m8-system"
	}
	item, err := c.dynamic.Resource(platformInstallationGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return installerv1alpha1.PlatformInstallation{}, fmt.Errorf("platform installation %s/%s not found", namespace, name)
		}
		return installerv1alpha1.PlatformInstallation{}, fmt.Errorf("get platform installation %s/%s: %w", namespace, name, err)
	}

	var installation installerv1alpha1.PlatformInstallation
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &installation); err != nil {
		return installerv1alpha1.PlatformInstallation{}, fmt.Errorf("decode platform installation %s/%s: %w", namespace, name, err)
	}
	return installation, nil
}

func (c *Client) ListPlatformInstallations(ctx context.Context, namespace string) ([]installerv1alpha1.PlatformInstallation, error) {
	list, err := c.dynamic.Resource(platformInstallationGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		scope := "all namespaces"
		if namespace != "" {
			scope = "namespace " + namespace
		}
		return nil, fmt.Errorf("list platform installations in %s: %w", scope, err)
	}

	installations := make([]installerv1alpha1.PlatformInstallation, 0, len(list.Items))
	for _, item := range list.Items {
		var installation installerv1alpha1.PlatformInstallation
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &installation); err != nil {
			return nil, fmt.Errorf("decode platform installation %s/%s: %w", item.GetNamespace(), item.GetName(), err)
		}
		installations = append(installations, installation)
	}
	return installations, nil
}

func isNodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
