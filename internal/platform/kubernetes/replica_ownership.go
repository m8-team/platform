package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type metadata struct {
	UID, Name, Namespace string
	DeletionTimestamp    *string
	OwnerReferences      []struct {
		UID, Kind, APIVersion string
		Controller            bool
	}
}

// ObserveReplica reads current, named resources using authenticated Kubernetes
// GETs. A missing or replaced controller retires its incarnation, while a
// missing Pod leaves its slot available to a future Pod of the same controller.
func ObserveReplica(ctx context.Context, client *http.Client, server, namespace, setName, setUID, podName string) (string, bool, error) {
	base, err := url.Parse(server)
	if err != nil || base.Scheme != "https" || base.Hostname() == "" || base.User != nil ||
		base.Path != "" && base.Path != "/" || base.RawQuery != "" || base.Fragment != "" ||
		client == nil || !validUID(setUID) {
		return "", false, fmt.Errorf("invalid replica observation configuration")
	}
	controller, present, err := getMetadata(ctx, client,
		base.JoinPath("apis", "apps", "v1", "namespaces", namespace, "statefulsets", setName))
	if err != nil {
		return "", false, err
	}
	if !present {
		return "", true, nil
	}
	if controller.Name != setName || controller.Namespace != namespace || !validUID(controller.UID) {
		return "", false, fmt.Errorf("invalid StatefulSet metadata")
	}
	if controller.UID != setUID || controller.DeletionTimestamp != nil {
		return "", true, nil
	}
	pod, present, err := getMetadata(ctx, client,
		base.JoinPath("api", "v1", "namespaces", namespace, "pods", podName))
	if err != nil || !present {
		return "", false, err
	}
	if pod.Name != podName || pod.Namespace != namespace || !validUID(pod.UID) {
		return "", false, fmt.Errorf("invalid Pod metadata")
	}
	if pod.DeletionTimestamp != nil {
		return "", false, nil
	}
	for _, ref := range pod.OwnerReferences {
		if ref.Controller && ref.APIVersion == "apps/v1" && ref.Kind == "StatefulSet" && ref.UID == setUID {
			return pod.UID, false, nil
		}
	}
	return "", false, nil
}

func validUID(value string) bool {
	id, err := uuid.Parse(value)
	return err == nil && id != uuid.Nil && id.String() == value
}

func getMetadata(ctx context.Context, client *http.Client, endpoint *url.URL) (metadata, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return metadata{}, false, err
	}
	kubeClient := *client
	kubeClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := kubeClient.Do(req)
	if err != nil {
		return metadata{}, false, fmt.Errorf("observe Kubernetes metadata: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return metadata{}, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return metadata{}, false, fmt.Errorf("observe Kubernetes metadata: HTTP %d", resp.StatusCode)
	}
	var object struct{ Metadata metadata }
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&object); err != nil {
		return metadata{}, false, fmt.Errorf("decode Kubernetes metadata: %w", err)
	}
	return object.Metadata, true, nil
}
