package kubernetes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	installerv1alpha1 "github.com/m8platform/platform/api/installer/v1alpha1"
	"github.com/m8platform/platform/internal/installer/preflight"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset kubernetes.Interface
	dynamic   dynamic.Interface
	mapper    *restmapper.DeferredDiscoveryRESTMapper
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
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(clientset.Discovery()))
	return &Client{clientset: clientset, dynamic: dynamicClient, mapper: mapper}, nil
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

var platformReleaseGVR = schema.GroupVersionResource{
	Group:    installerv1alpha1.GroupName,
	Version:  installerv1alpha1.Version,
	Resource: "platformreleases",
}

var installationOperationGVR = schema.GroupVersionResource{
	Group:    installerv1alpha1.GroupName,
	Version:  installerv1alpha1.Version,
	Resource: "installationoperations",
}

var customResourceDefinitionGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1",
	Resource: "customresourcedefinitions",
}

var namespaceGVR = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "namespaces",
}

func (c *Client) ApplyInstallerCRDs(ctx context.Context) error {
	for _, crd := range installerCRDs() {
		if err := c.applyUnstructured(ctx, customResourceDefinitionGVR, "", crd); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) WaitForInstallerAPI(ctx context.Context, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	required := []struct {
		groupVersion string
		kind         string
	}{
		{installerv1alpha1.GroupName + "/" + installerv1alpha1.Version, "PlatformInstallation"},
		{installerv1alpha1.GroupName + "/" + installerv1alpha1.Version, "PlatformRelease"},
		{installerv1alpha1.GroupName + "/" + installerv1alpha1.Version, "InstallationOperation"},
	}

	for {
		allFound := true
		for _, item := range required {
			found, err := c.HasAPIResource(ctx, item.groupVersion, item.kind)
			if err != nil {
				allFound = false
				break
			}
			if !found {
				allFound = false
				break
			}
		}
		if allFound {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for installer API discovery: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

func (c *Client) WaitForAPIResource(ctx context.Context, groupVersion string, kind string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		found, err := c.HasAPIResource(ctx, groupVersion, kind)
		if err == nil && found {
			if c.mapper != nil {
				c.mapper.Reset()
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for API resource %s %s: %w", groupVersion, kind, ctx.Err())
		case <-ticker.C:
		}
	}
}

func (c *Client) ApplyYAMLDocuments(ctx context.Context, data []byte, defaultNamespace string) ([]string, error) {
	decoder := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	var applied []string

	for {
		var raw map[string]any
		err := decoder.Decode(&raw)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return applied, fmt.Errorf("decode Kubernetes manifest: %w", err)
		}
		if len(raw) == 0 {
			continue
		}

		object := &unstructured.Unstructured{Object: raw}
		if object.GetKind() == "" || object.GetAPIVersion() == "" {
			continue
		}
		if object.GetName() == "" {
			return applied, fmt.Errorf("manifest object %s is missing metadata.name", object.GroupVersionKind())
		}

		mapping, err := c.restMapping(object.GroupVersionKind())
		if err != nil {
			return applied, err
		}

		namespace := ""
		if mapping.Scope.Name() != meta.RESTScopeNameRoot {
			if object.GetNamespace() == "" {
				object.SetNamespace(defaultNamespace)
			}
			namespace = object.GetNamespace()
			if namespace == "" {
				return applied, fmt.Errorf("manifest object %s/%s requires a namespace", object.GetKind(), object.GetName())
			}
		}

		if err := c.applyUnstructured(ctx, mapping.Resource, namespace, object); err != nil {
			return applied, err
		}
		applied = append(applied, objectReference(object, namespace))
	}

	return applied, nil
}

func (c *Client) restMapping(gvk schema.GroupVersionKind) (*meta.RESTMapping, error) {
	if c.mapper == nil {
		return nil, fmt.Errorf("REST mapper is not configured")
	}
	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err == nil {
		return mapping, nil
	}
	if meta.IsNoMatchError(err) {
		c.mapper.Reset()
		mapping, err = c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	}
	if err != nil {
		return nil, fmt.Errorf("resolve REST mapping for %s: %w", gvk.String(), err)
	}
	return mapping, nil
}

func (c *Client) ApplyNamespace(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("namespace name is required")
	}
	namespace := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "v1",
		"kind":       "Namespace",
		"metadata": map[string]any{
			"name": name,
			"labels": map[string]any{
				"app.kubernetes.io/managed-by":       "m8ctl",
				"pod-security.kubernetes.io/enforce": "restricted",
				"pod-security.kubernetes.io/audit":   "restricted",
				"pod-security.kubernetes.io/warn":    "restricted",
			},
		},
	}}
	return c.applyUnstructured(ctx, namespaceGVR, "", namespace)
}

func (c *Client) ApplyPlatformRelease(ctx context.Context, release installerv1alpha1.PlatformRelease) error {
	object, err := toUnstructured(release)
	if err != nil {
		return err
	}
	return c.applyUnstructured(ctx, platformReleaseGVR, "", object)
}

func (c *Client) ApplyPlatformInstallation(ctx context.Context, installation installerv1alpha1.PlatformInstallation) error {
	if installation.Namespace == "" {
		installation.Namespace = "m8-system"
	}
	object, err := toUnstructured(installation)
	if err != nil {
		return err
	}
	return c.applyUnstructured(ctx, platformInstallationGVR, installation.Namespace, object)
}

func (c *Client) ApplyInstallationOperation(ctx context.Context, operation installerv1alpha1.InstallationOperation) error {
	if operation.Namespace == "" {
		operation.Namespace = "m8-system"
	}
	operation.Status = installerv1alpha1.InstallationOperationStatus{}
	object, err := toUnstructured(operation)
	if err != nil {
		return err
	}
	return c.applyUnstructured(ctx, installationOperationGVR, operation.Namespace, object)
}

func (c *Client) applyUnstructured(ctx context.Context, gvr schema.GroupVersionResource, namespace string, object *unstructured.Unstructured) error {
	options := metav1.ApplyOptions{
		FieldManager: "m8ctl",
		Force:        true,
	}
	resource := c.dynamic.Resource(gvr)
	if namespace != "" {
		if _, err := resource.Namespace(namespace).Apply(ctx, object.GetName(), object, options); err != nil {
			return fmt.Errorf("server-side apply %s/%s %s/%s: %w", gvr.Group, gvr.Resource, namespace, object.GetName(), err)
		}
		return nil
	}
	if _, err := resource.Apply(ctx, object.GetName(), object, options); err != nil {
		return fmt.Errorf("server-side apply %s/%s %s/%s: %w", gvr.Group, gvr.Resource, namespace, object.GetName(), err)
	}
	return nil
}

func toUnstructured(value any) (*unstructured.Unstructured, error) {
	if value == nil {
		return nil, fmt.Errorf("encode Kubernetes object: value is nil")
	}

	input := value
	reflected := reflect.ValueOf(value)
	if reflected.Kind() != reflect.Pointer {
		pointer := reflect.New(reflected.Type())
		pointer.Elem().Set(reflected)
		input = pointer.Interface()
	}

	object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(input)
	if err != nil {
		return nil, fmt.Errorf("encode Kubernetes object: %w", err)
	}
	return &unstructured.Unstructured{Object: object}, nil
}

func objectReference(object *unstructured.Unstructured, namespace string) string {
	gvk := object.GroupVersionKind()
	if namespace == "" {
		return strings.ToLower(gvk.Kind) + "/" + object.GetName()
	}
	return strings.ToLower(gvk.Kind) + "/" + namespace + "/" + object.GetName()
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

func installerCRDs() []*unstructured.Unstructured {
	return []*unstructured.Unstructured{
		installerCRD("platforminstallations", "platforminstallation", "PlatformInstallation", "m8pi", "Namespaced"),
		installerCRD("platformreleases", "platformrelease", "PlatformRelease", "m8rel", "Cluster"),
		installerCRD("installationoperations", "installationoperation", "InstallationOperation", "m8op", "Namespaced"),
		installerCRD("backups", "backup", "Backup", "m8backup", "Namespaced"),
		installerCRD("restoreplans", "restoreplan", "RestorePlan", "m8restore", "Namespaced"),
	}
}

func installerCRD(plural string, singular string, kind string, shortName string, scope string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "CustomResourceDefinition",
		"metadata": map[string]any{
			"name": plural + "." + installerv1alpha1.GroupName,
			"labels": map[string]any{
				"app.kubernetes.io/managed-by": "m8ctl",
				"app.kubernetes.io/part-of":    "m8-installer",
			},
		},
		"spec": map[string]any{
			"group": installerv1alpha1.GroupName,
			"scope": scope,
			"names": map[string]any{
				"plural":     plural,
				"singular":   singular,
				"kind":       kind,
				"shortNames": []any{shortName},
			},
			"versions": []any{map[string]any{
				"name":    installerv1alpha1.Version,
				"served":  true,
				"storage": true,
				"subresources": map[string]any{
					"status": map[string]any{},
				},
				"schema": map[string]any{
					"openAPIV3Schema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"apiVersion": map[string]any{"type": "string"},
							"kind":       map[string]any{"type": "string"},
							"metadata":   map[string]any{"type": "object"},
							"spec": map[string]any{
								"type":                                 "object",
								"x-kubernetes-preserve-unknown-fields": true,
							},
							"status": map[string]any{
								"type":                                 "object",
								"x-kubernetes-preserve-unknown-fields": true,
							},
						},
					},
				},
			}},
		},
	}}
}

func isNodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
