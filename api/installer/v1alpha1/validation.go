package v1alpha1

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrInvalidInstallation = errors.New("invalid platform installation")
	ErrInvalidRelease      = errors.New("invalid platform release")
)

type FieldError struct {
	Field   string `json:"field" yaml:"field"`
	Message string `json:"message" yaml:"message"`
}

func (e FieldError) Error() string {
	return e.Field + ": " + e.Message
}

type FieldErrors []FieldError

func (e FieldErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	parts := make([]string, 0, len(e))
	for _, item := range e {
		parts = append(parts, item.Error())
	}

	return strings.Join(parts, "; ")
}

func (e FieldErrors) HasErrors() bool {
	return len(e) > 0
}

func (p Profile) IsValid() bool {
	switch p {
	case ProfileDemo, ProfileDevelopment, ProfileStaging, ProfileProduction, ProfileAirGapped:
		return true
	default:
		return false
	}
}

func (m ComponentMode) IsValid(allowed ...ComponentMode) bool {
	if m == "" {
		return true
	}
	for _, item := range allowed {
		if m == item {
			return true
		}
	}
	return false
}

func (in PlatformInstallation) Defaulted() PlatformInstallation {
	out := in
	out.Spec = out.Spec.Defaulted()
	if out.APIVersion == "" {
		out.APIVersion = GroupName + "/" + Version
	}
	if out.Kind == "" {
		out.Kind = "PlatformInstallation"
	}
	return out
}

func (s PlatformInstallationSpec) Defaulted() PlatformInstallationSpec {
	out := s
	if out.Profile == "" {
		out.Profile = ProfileDevelopment
	}
	if out.Cluster.MinimumNodes == 0 {
		switch out.Profile {
		case ProfileProduction, ProfileAirGapped:
			out.Cluster.MinimumNodes = 3
		default:
			out.Cluster.MinimumNodes = 1
		}
	}
	if len(out.Cluster.RequiredArchitectures) == 0 {
		out.Cluster.RequiredArchitectures = []string{"amd64", "arm64"}
	}
	if out.Network.ManagementPolicy == "" {
		out.Network.ManagementPolicy = ManagementPolicyCreate
	}
	if out.Gateway.ManagementPolicy == "" {
		out.Gateway.ManagementPolicy = ManagementPolicyCreate
	}
	if out.Gateway.GatewayClassName == "" {
		out.Gateway.GatewayClassName = "m8-public"
	}
	if out.Gateway.ControllerName == "" {
		out.Gateway.ControllerName = "gateway.envoyproxy.io/gatewayclass-controller"
	}
	if out.GitOps.ArgoCD.Namespace == "" {
		out.GitOps.ArgoCD.Namespace = "argocd"
	}
	if out.Identity.Keycloak.Realm == "" {
		out.Identity.Keycloak.Realm = "m8"
	}
	if out.Authorization.SpiceDB.Datastore == "" {
		out.Authorization.SpiceDB.Datastore = "postgresql"
	}
	if len(out.Workflows.Temporal.Namespaces) == 0 {
		out.Workflows.Temporal.Namespaces = []string{
			"m8-system",
			"m8-operations",
			"m8-provisioning",
			"m8-authentication",
			"m8-analytics",
		}
	}
	modulesWereEmpty := !out.Modules.anyEnabled()
	if modulesWereEmpty {
		out.Modules = defaultModules()
	}
	if out.Network.ExistingCNI.Name == "" {
		out.Network.Cilium.Enabled = true
	}
	if out.Modules.Gateway {
		out.Gateway.EnvoyGateway.Enabled = true
	}
	out.Certificates.CertManager.Enabled = true
	out.Trust.TrustManager.Enabled = true
	out.GitOps.ArgoCD.Enabled = true
	if out.Observability.Mode == "" {
		out.Observability.Mode = ModeOperator
	}
	out.Observability.OpenTelemetryOperator.Enabled = true
	out.Observability.PrometheusOperator.Enabled = true

	switch out.Profile {
	case ProfileDemo:
		out.Secrets.Provider = defaultSecretsProvider(out.Secrets.Provider, SecretsProviderKubernetes)
		out.Databases.PostgreSQL.Mode = defaultMode(out.Databases.PostgreSQL.Mode, ModeSingle)
		out.Databases.YDB.Mode = defaultMode(out.Databases.YDB.Mode, ModeOperator)
		out.Databases.Redis.Mode = defaultMode(out.Databases.Redis.Mode, ModeSingle)
		out.Messaging.Kafka.Mode = defaultMode(out.Messaging.Kafka.Mode, ModeStrimzi)
		out.Workflows.Temporal.Mode = defaultMode(out.Workflows.Temporal.Mode, ModeOperator)
		out.Identity.Keycloak.Mode = defaultMode(out.Identity.Keycloak.Mode, ModeOperator)
		out.Authorization.SpiceDB.Mode = defaultMode(out.Authorization.SpiceDB.Mode, ModeOperator)
	case ProfileDevelopment:
		out.Secrets.Provider = defaultSecretsProvider(out.Secrets.Provider, SecretsProviderKubernetes)
		out.Databases.PostgreSQL.Mode = defaultMode(out.Databases.PostgreSQL.Mode, ModeCloudNativePG)
		out.Databases.YDB.Mode = defaultMode(out.Databases.YDB.Mode, ModeOperator)
		out.Databases.Redis.Mode = defaultMode(out.Databases.Redis.Mode, ModeSingle)
		out.Messaging.Kafka.Mode = defaultMode(out.Messaging.Kafka.Mode, ModeStrimzi)
		out.Workflows.Temporal.Mode = defaultMode(out.Workflows.Temporal.Mode, ModeOperator)
		out.Identity.Keycloak.Mode = defaultMode(out.Identity.Keycloak.Mode, ModeOperator)
		out.Authorization.SpiceDB.Mode = defaultMode(out.Authorization.SpiceDB.Mode, ModeOperator)
	case ProfileStaging, ProfileProduction, ProfileAirGapped:
		out.Secrets.Provider = defaultSecretsProvider(out.Secrets.Provider, SecretsProviderExternalSecrets)
		out.Databases.PostgreSQL.Mode = defaultMode(out.Databases.PostgreSQL.Mode, ModeCloudNativePG)
		out.Databases.YDB.Mode = defaultMode(out.Databases.YDB.Mode, ModeOperator)
		out.Databases.Redis.Mode = defaultMode(out.Databases.Redis.Mode, ModeSentinel)
		out.Messaging.Kafka.Mode = defaultMode(out.Messaging.Kafka.Mode, ModeStrimzi)
		out.Workflows.Temporal.Mode = defaultMode(out.Workflows.Temporal.Mode, ModeOperator)
		out.Identity.Keycloak.Mode = defaultMode(out.Identity.Keycloak.Mode, ModeOperator)
		out.Authorization.SpiceDB.Mode = defaultMode(out.Authorization.SpiceDB.Mode, ModeOperator)
		out.Backup.Enabled = true
		out.Upgrade.RequireBackup = true
		out.Network.NetworkPolicies = true
		out.Security.PodSecurityStandards = true
		out.Security.RunAsNonRootRequired = true
		out.Security.ReadOnlyRootFS = true
		out.Security.RequireRequestsLimits = true
		out.Security.ForbidPrivileged = true
		out.Security.ForbidHostPath = true
		out.Security.ForbidLatestTag = true
		out.Security.SBOMRequired = true
		out.Security.Cosign.Enforce = true
		out.Trust.SPIRE.Enabled = true
		out.Databases.PostgreSQL.HA = true
		out.Databases.PostgreSQL.Instances = 3
		out.Gateway.EnvoyGateway.Replicas = 2
	}

	if out.Profile == ProfileAirGapped {
		out.AirGap.Enabled = true
	}
	if out.Secrets.Provider != SecretsProviderKubernetes {
		out.Secrets.ExternalSecretsOperator.Enabled = true
	}

	return out
}

func defaultMode(value ComponentMode, fallback ComponentMode) ComponentMode {
	if value == "" {
		return fallback
	}
	return value
}

func defaultSecretsProvider(value SecretsProvider, fallback SecretsProvider) SecretsProvider {
	if value == "" {
		return fallback
	}
	return value
}

func defaultModules() ModulesSpec {
	return ModulesSpec{
		Console:         true,
		ResourceManager: true,
		Identity:        true,
		Authentication:  true,
		Access:          true,
		Gateway:         true,
		Audit:           true,
		Operations:      true,
		Idempotency:     true,
		Health:          true,
		Alerting:        true,
		Provisioning:    true,
	}
}

func (m ModulesSpec) anyEnabled() bool {
	return m.Console ||
		m.ResourceManager ||
		m.Identity ||
		m.Authentication ||
		m.Access ||
		m.Gateway ||
		m.PolicyStudio ||
		m.Audit ||
		m.Operations ||
		m.Idempotency ||
		m.Health ||
		m.Alerting ||
		m.Notifications ||
		m.Provisioning
}

func (in PlatformInstallation) Validate() error {
	installation := in.Defaulted()
	var errs FieldErrors
	spec := installation.Spec

	if installation.Name == "" {
		errs = append(errs, FieldError{Field: "metadata.name", Message: "name is required"})
	}
	if spec.PlatformVersion == "" {
		errs = append(errs, FieldError{Field: "spec.platformVersion", Message: "platform version is required"})
	}
	if !spec.Profile.IsValid() {
		errs = append(errs, FieldError{Field: "spec.profile", Message: "unsupported profile"})
	}
	if spec.Cluster.MinimumNodes < 1 {
		errs = append(errs, FieldError{Field: "spec.cluster.minimumNodes", Message: "must be at least 1"})
	}
	if spec.Profile == ProfileProduction && spec.Cluster.MinimumNodes < 3 {
		errs = append(errs, FieldError{Field: "spec.cluster.minimumNodes", Message: "production requires at least 3 worker nodes"})
	}
	if spec.Profile == ProfileProduction && spec.Secrets.Provider == SecretsProviderKubernetes {
		errs = append(errs, FieldError{Field: "spec.secrets.provider", Message: "production must use an external secrets provider"})
	}
	if spec.Profile == ProfileProduction && !spec.Backup.Enabled {
		errs = append(errs, FieldError{Field: "spec.backup.enabled", Message: "production requires backup"})
	}
	if spec.Profile == ProfileProduction && !spec.Security.Cosign.Enforce {
		errs = append(errs, FieldError{Field: "spec.security.cosign.enforce", Message: "production requires signature enforcement"})
	}
	if spec.Profile == ProfileAirGapped && spec.AirGap.PrivateRegistry == "" {
		errs = append(errs, FieldError{Field: "spec.airGap.privateRegistry", Message: "air-gapped profile requires a private registry"})
	}
	if spec.Network.Cilium.Enabled && spec.Network.ExistingCNI.Name != "" && !spec.Network.AllowCNIReplacement {
		errs = append(errs, FieldError{Field: "spec.network.allowCNIReplacement", Message: "CNI replacement requires explicit allowCNIReplacement=true"})
	}
	if spec.Gateway.EnvoyGateway.Enabled && spec.Gateway.GatewayClassName == "" {
		errs = append(errs, FieldError{Field: "spec.gateway.gatewayClassName", Message: "gateway class name is required when Envoy Gateway is enabled"})
	}

	validateModes(&errs, spec)
	validateModuleDependencies(&errs, spec)

	if errs.HasErrors() {
		return fmt.Errorf("%w: %s", ErrInvalidInstallation, errs.Error())
	}

	return nil
}

func validateModes(errs *FieldErrors, spec PlatformInstallationSpec) {
	if !spec.Databases.PostgreSQL.Mode.IsValid(ModeExternal, ModeCloudNativePG, ModeSingle, ModeDisabled) {
		*errs = append(*errs, FieldError{Field: "spec.databases.postgresql.mode", Message: "must be external, cloudnative-pg, single, or disabled"})
	}
	if !spec.Databases.YDB.Mode.IsValid(ModeExternal, ModeOperator, ModeDisabled) {
		*errs = append(*errs, FieldError{Field: "spec.databases.ydb.mode", Message: "must be external, operator, or disabled"})
	}
	if !spec.Messaging.Kafka.Mode.IsValid(ModeExternal, ModeStrimzi, ModeDisabled) {
		*errs = append(*errs, FieldError{Field: "spec.messaging.kafka.mode", Message: "must be external, strimzi, or disabled"})
	}
	if !spec.Databases.Redis.Mode.IsValid(ModeExternal, ModeSentinel, ModeCluster, ModeSingle, ModeDisabled) {
		*errs = append(*errs, FieldError{Field: "spec.databases.redis.mode", Message: "must be external, sentinel, cluster, single, or disabled"})
	}
	if !spec.Workflows.Temporal.Mode.IsValid(ModeExternal, ModeOperator, ModeDisabled) {
		*errs = append(*errs, FieldError{Field: "spec.workflows.temporal.mode", Message: "must be external, operator, or disabled"})
	}
	if !spec.Identity.Keycloak.Mode.IsValid(ModeExternal, ModeOperator, ModeDisabled) {
		*errs = append(*errs, FieldError{Field: "spec.identity.keycloak.mode", Message: "must be external, operator, or disabled"})
	}
	if !spec.Authorization.SpiceDB.Mode.IsValid(ModeExternal, ModeOperator, ModeDisabled) {
		*errs = append(*errs, FieldError{Field: "spec.authorization.spiceDB.mode", Message: "must be external, operator, or disabled"})
	}
}

func validateModuleDependencies(errs *FieldErrors, spec PlatformInstallationSpec) {
	if spec.Modules.Authentication {
		if !spec.Modules.Identity {
			*errs = append(*errs, FieldError{Field: "spec.modules.authentication", Message: "authentication requires identity module"})
		}
		if spec.Identity.Keycloak.Mode == ModeDisabled {
			*errs = append(*errs, FieldError{Field: "spec.identity.keycloak.mode", Message: "authentication requires Keycloak"})
		}
	}
	if spec.Modules.Access && spec.Authorization.SpiceDB.Mode == ModeDisabled {
		*errs = append(*errs, FieldError{Field: "spec.authorization.spiceDB.mode", Message: "access requires SpiceDB"})
	}
	if spec.Modules.Provisioning && spec.Workflows.Temporal.Mode == ModeDisabled {
		*errs = append(*errs, FieldError{Field: "spec.workflows.temporal.mode", Message: "provisioning requires Temporal"})
	}
	if spec.Modules.Gateway && !spec.Gateway.EnvoyGateway.Enabled {
		*errs = append(*errs, FieldError{Field: "spec.gateway.envoyGateway.enabled", Message: "gateway module requires Envoy Gateway"})
	}
	if spec.Modules.Audit && spec.Databases.YDB.Mode == ModeDisabled {
		*errs = append(*errs, FieldError{Field: "spec.databases.ydb.mode", Message: "audit requires YDB"})
	}
	if spec.Modules.Operations && spec.Workflows.Temporal.Mode == ModeDisabled {
		*errs = append(*errs, FieldError{Field: "spec.workflows.temporal.mode", Message: "operations requires Temporal"})
	}
}

func (in PlatformRelease) Validate() error {
	var errs FieldErrors
	if in.Name == "" {
		errs = append(errs, FieldError{Field: "metadata.name", Message: "name is required"})
	}
	if in.Spec.Kubernetes.MinVersion == "" {
		errs = append(errs, FieldError{Field: "spec.kubernetes.minVersion", Message: "minimum Kubernetes version is required"})
	}
	if len(in.Spec.Components) == 0 {
		errs = append(errs, FieldError{Field: "spec.components", Message: "at least one component is required"})
	}
	for name, component := range in.Spec.Components {
		prefix := "spec.components." + name
		if component.Version == "" {
			errs = append(errs, FieldError{Field: prefix + ".version", Message: "version is required"})
		}
		if strings.EqualFold(component.Version, "latest") || component.Version == "main" || component.Version == "master" {
			errs = append(errs, FieldError{Field: prefix + ".version", Message: "floating versions are not allowed"})
		}
		validateArtifact(&errs, prefix+".chart", component.Chart, component.Chart.Repository != "")
		for i, image := range component.Images {
			validateArtifact(&errs, fmt.Sprintf("%s.images[%d]", prefix, i), image, true)
		}
	}
	if errs.HasErrors() {
		return fmt.Errorf("%w: %s", ErrInvalidRelease, errs.Error())
	}
	return nil
}

var digestPattern = regexp.MustCompile(`^sha256:[a-fA-F0-9]{64}$`)

func validateArtifact(errs *FieldErrors, field string, artifact ArtifactRef, required bool) {
	if !required {
		return
	}
	if artifact.Repository == "" {
		*errs = append(*errs, FieldError{Field: field + ".repository", Message: "repository is required"})
	}
	if artifact.Digest == "" {
		*errs = append(*errs, FieldError{Field: field + ".digest", Message: "digest is required"})
	} else if !digestPattern.MatchString(artifact.Digest) {
		*errs = append(*errs, FieldError{Field: field + ".digest", Message: "digest must be a sha256 digest"})
	}
	if strings.EqualFold(artifact.Version, "latest") || artifact.Version == "main" || artifact.Version == "master" {
		*errs = append(*errs, FieldError{Field: field + ".version", Message: "floating versions are not allowed"})
	}
}
