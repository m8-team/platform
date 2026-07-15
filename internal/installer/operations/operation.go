package operations

import (
	"time"

	"github.com/google/uuid"
	installerv1alpha1 "github.com/m8platform/platform/api/installer/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewInstallationOperation(
	installationName string,
	operationType installerv1alpha1.OperationType,
	requestedVersion string,
	planDigest string,
	requestedBy string,
	now time.Time,
) installerv1alpha1.InstallationOperation {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	startedAt := metav1.NewTime(now)
	return installerv1alpha1.InstallationOperation{
		TypeMeta: metav1.TypeMeta{
			APIVersion: installerv1alpha1.GroupName + "/" + installerv1alpha1.Version,
			Kind:       "InstallationOperation",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: installationName + "-" + string(operationType) + "-" + now.Format("20060102150405"),
		},
		Spec: installerv1alpha1.InstallationOperationSpec{
			OperationID:      uuid.NewString(),
			Type:             operationType,
			InstallationRef:  installationName,
			RequestedVersion: requestedVersion,
			PlanDigest:       planDigest,
			RequestedBy:      requestedBy,
		},
		Status: installerv1alpha1.InstallationOperationStatus{
			Phase:     installerv1alpha1.PhasePending,
			StartedAt: &startedAt,
		},
	}
}
