package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	GroupName = "installer.m8.io"
	Version   = "v1alpha1"
)

type Profile string

const (
	ProfileDemo        Profile = "demo"
	ProfileDevelopment Profile = "development"
	ProfileStaging     Profile = "staging"
	ProfileProduction  Profile = "production"
	ProfileAirGapped   Profile = "air-gapped"
)

type Phase string

const (
	PhasePending       Phase = "Pending"
	PhaseValidating    Phase = "Validating"
	PhasePlanning      Phase = "Planning"
	PhaseBootstrapping Phase = "Bootstrapping"
	PhaseInstalling    Phase = "Installing"
	PhaseMigrating     Phase = "Migrating"
	PhaseVerifying     Phase = "Verifying"
	PhaseReady         Phase = "Ready"
	PhaseDegraded      Phase = "Degraded"
	PhaseFailed        Phase = "Failed"
	PhaseUpgrading     Phase = "Upgrading"
	PhaseRollingBack   Phase = "RollingBack"
	PhaseBackingUp     Phase = "BackingUp"
	PhaseRestoring     Phase = "Restoring"
	PhaseUninstalling  Phase = "Uninstalling"
)

type ComponentMode string

const (
	ModeExternal      ComponentMode = "external"
	ModeOperator      ComponentMode = "operator"
	ModeDisabled      ComponentMode = "disabled"
	ModeCloudNativePG ComponentMode = "cloudnative-pg"
	ModeStrimzi       ComponentMode = "strimzi"
	ModeSentinel      ComponentMode = "sentinel"
	ModeCluster       ComponentMode = "cluster"
	ModeSingle        ComponentMode = "single"
)

type ManagementPolicy string

const (
	ManagementPolicyCreate  ManagementPolicy = "create"
	ManagementPolicyAdopt   ManagementPolicy = "adopt"
	ManagementPolicyObserve ManagementPolicy = "observe"
)

type PlatformInstallation struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   PlatformInstallationSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status PlatformInstallationStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type PlatformInstallationSpec struct {
	PlatformVersion string            `json:"platformVersion" yaml:"platformVersion"`
	Profile         Profile           `json:"profile,omitempty" yaml:"profile,omitempty"`
	Cluster         ClusterSpec       `json:"cluster,omitempty" yaml:"cluster,omitempty"`
	Network         NetworkSpec       `json:"network,omitempty" yaml:"network,omitempty"`
	Gateway         GatewaySpec       `json:"gateway,omitempty" yaml:"gateway,omitempty"`
	Certificates    CertificatesSpec  `json:"certificates,omitempty" yaml:"certificates,omitempty"`
	Trust           TrustSpec         `json:"trust,omitempty" yaml:"trust,omitempty"`
	GitOps          GitOpsSpec        `json:"gitOps,omitempty" yaml:"gitOps,omitempty"`
	Secrets         SecretsSpec       `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	Databases       DatabasesSpec     `json:"databases,omitempty" yaml:"databases,omitempty"`
	Messaging       MessagingSpec     `json:"messaging,omitempty" yaml:"messaging,omitempty"`
	Workflows       WorkflowsSpec     `json:"workflows,omitempty" yaml:"workflows,omitempty"`
	Identity        IdentitySpec      `json:"identity,omitempty" yaml:"identity,omitempty"`
	Authorization   AuthorizationSpec `json:"authorization,omitempty" yaml:"authorization,omitempty"`
	Observability   ObservabilitySpec `json:"observability,omitempty" yaml:"observability,omitempty"`
	Security        SecuritySpec      `json:"security,omitempty" yaml:"security,omitempty"`
	Backup          BackupSpec        `json:"backup,omitempty" yaml:"backup,omitempty"`
	Modules         ModulesSpec       `json:"modules,omitempty" yaml:"modules,omitempty"`
	Upgrade         UpgradeSpec       `json:"upgrade,omitempty" yaml:"upgrade,omitempty"`
	AirGap          AirGapSpec        `json:"airGap,omitempty" yaml:"airGap,omitempty"`
}

type ClusterSpec struct {
	Name                  string             `json:"name,omitempty" yaml:"name,omitempty"`
	MinimumNodes          int32              `json:"minimumNodes,omitempty" yaml:"minimumNodes,omitempty"`
	RequiredArchitectures []string           `json:"requiredArchitectures,omitempty" yaml:"requiredArchitectures,omitempty"`
	RequireTopologyLabels bool               `json:"requireTopologyLabels,omitempty" yaml:"requireTopologyLabels,omitempty"`
	StorageClasses        []StorageClassSpec `json:"storageClasses,omitempty" yaml:"storageClasses,omitempty"`
	Clusters              []MemberCluster    `json:"clusters,omitempty" yaml:"clusters,omitempty"`
}

type StorageClassSpec struct {
	Name         string `json:"name" yaml:"name"`
	Required     bool   `json:"required,omitempty" yaml:"required,omitempty"`
	DefaultClass bool   `json:"defaultClass,omitempty" yaml:"defaultClass,omitempty"`
}

type ClusterRole string

const (
	ClusterRoleManagement       ClusterRole = "management"
	ClusterRoleWorkload         ClusterRole = "workload"
	ClusterRoleData             ClusterRole = "data"
	ClusterRoleEdge             ClusterRole = "edge"
	ClusterRoleObservability    ClusterRole = "observability"
	ClusterRoleDisasterRecovery ClusterRole = "disaster-recovery"
)

type MemberCluster struct {
	Name    string        `json:"name" yaml:"name"`
	Role    ClusterRole   `json:"role" yaml:"role"`
	Region  string        `json:"region,omitempty" yaml:"region,omitempty"`
	Labels  []NameValue   `json:"labels,omitempty" yaml:"labels,omitempty"`
	GitOps  GitOpsBinding `json:"gitOps,omitempty" yaml:"gitOps,omitempty"`
	Enabled bool          `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}

type NameValue struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type NetworkSpec struct {
	Cilium               CiliumSpec       `json:"cilium,omitempty" yaml:"cilium,omitempty"`
	ExistingCNI          CNISpec          `json:"existingCNI,omitempty" yaml:"existingCNI,omitempty"`
	NetworkPolicies      bool             `json:"networkPolicies,omitempty" yaml:"networkPolicies,omitempty"`
	WireGuardEncryption  bool             `json:"wireGuardEncryption,omitempty" yaml:"wireGuardEncryption,omitempty"`
	KubeProxyReplacement string           `json:"kubeProxyReplacement,omitempty" yaml:"kubeProxyReplacement,omitempty"`
	DNSPolicies          bool             `json:"dnsPolicies,omitempty" yaml:"dnsPolicies,omitempty"`
	EastWestServiceMesh  bool             `json:"eastWestServiceMesh,omitempty" yaml:"eastWestServiceMesh,omitempty"`
	AllowCNIReplacement  bool             `json:"allowCNIReplacement,omitempty" yaml:"allowCNIReplacement,omitempty"`
	ManagementPolicy     ManagementPolicy `json:"managementPolicy,omitempty" yaml:"managementPolicy,omitempty"`
}

type CiliumSpec struct {
	Enabled       bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Version       string `json:"version,omitempty" yaml:"version,omitempty"`
	HubbleRelay   bool   `json:"hubbleRelay,omitempty" yaml:"hubbleRelay,omitempty"`
	HubbleUI      bool   `json:"hubbleUI,omitempty" yaml:"hubbleUI,omitempty"`
	RackAwareness bool   `json:"rackAwareness,omitempty" yaml:"rackAwareness,omitempty"`
}

type CNISpec struct {
	Name       string `json:"name,omitempty" yaml:"name,omitempty"`
	Version    string `json:"version,omitempty" yaml:"version,omitempty"`
	ManagedBy  string `json:"managedBy,omitempty" yaml:"managedBy,omitempty"`
	Compatible bool   `json:"compatible,omitempty" yaml:"compatible,omitempty"`
}

type GatewaySpec struct {
	EnvoyGateway     EnvoyGatewaySpec `json:"envoyGateway,omitempty" yaml:"envoyGateway,omitempty"`
	GatewayClassName string           `json:"gatewayClassName,omitempty" yaml:"gatewayClassName,omitempty"`
	ControllerName   string           `json:"controllerName,omitempty" yaml:"controllerName,omitempty"`
	ExternalDNS      bool             `json:"externalDNS,omitempty" yaml:"externalDNS,omitempty"`
	TLSTermination   bool             `json:"tlsTermination,omitempty" yaml:"tlsTermination,omitempty"`
	AllowedHostnames []string         `json:"allowedHostnames,omitempty" yaml:"allowedHostnames,omitempty"`
	ManagementPolicy ManagementPolicy `json:"managementPolicy,omitempty" yaml:"managementPolicy,omitempty"`
}

type EnvoyGatewaySpec struct {
	Enabled  bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Version  string `json:"version,omitempty" yaml:"version,omitempty"`
	Replicas int32  `json:"replicas,omitempty" yaml:"replicas,omitempty"`
}

type CertificatesSpec struct {
	CertManager CertManagerSpec `json:"certManager,omitempty" yaml:"certManager,omitempty"`
	Issuers     []IssuerSpec    `json:"issuers,omitempty" yaml:"issuers,omitempty"`
}

type CertManagerSpec struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

type IssuerType string

const (
	IssuerTypeInternalCA  IssuerType = "internal-ca"
	IssuerTypeACME        IssuerType = "acme"
	IssuerTypeCorporateCA IssuerType = "corporate-ca"
	IssuerTypeVaultPKI    IssuerType = "vault-pki"
)

type IssuerSpec struct {
	Name      string     `json:"name" yaml:"name"`
	Type      IssuerType `json:"type" yaml:"type"`
	Default   bool       `json:"default,omitempty" yaml:"default,omitempty"`
	SecretRef string     `json:"secretRef,omitempty" yaml:"secretRef,omitempty"`
	Server    string     `json:"server,omitempty" yaml:"server,omitempty"`
	Email     string     `json:"email,omitempty" yaml:"email,omitempty"`
	Rotation  Duration   `json:"rotation,omitempty" yaml:"rotation,omitempty"`
}

type TrustSpec struct {
	TrustManager       TrustManagerSpec  `json:"trustManager,omitempty" yaml:"trustManager,omitempty"`
	SPIRE              SPIRESpec         `json:"spire,omitempty" yaml:"spire,omitempty"`
	TrustDomain        string            `json:"trustDomain,omitempty" yaml:"trustDomain,omitempty"`
	Federation         []SPIREFederation `json:"federation,omitempty" yaml:"federation,omitempty"`
	BundleConfigMapRef string            `json:"bundleConfigMapRef,omitempty" yaml:"bundleConfigMapRef,omitempty"`
}

type TrustManagerSpec struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

type SPIRESpec struct {
	Enabled               bool     `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Version               string   `json:"version,omitempty" yaml:"version,omitempty"`
	ServerReplicas        int32    `json:"serverReplicas,omitempty" yaml:"serverReplicas,omitempty"`
	OIDCDiscoveryProvider bool     `json:"oidcDiscoveryProvider,omitempty" yaml:"oidcDiscoveryProvider,omitempty"`
	CSIDriver             bool     `json:"csiDriver,omitempty" yaml:"csiDriver,omitempty"`
	NodeAttestation       string   `json:"nodeAttestation,omitempty" yaml:"nodeAttestation,omitempty"`
	WorkloadAttestation   string   `json:"workloadAttestation,omitempty" yaml:"workloadAttestation,omitempty"`
	SVIDRotation          Duration `json:"svidRotation,omitempty" yaml:"svidRotation,omitempty"`
}

type SPIREFederation struct {
	TrustDomain string `json:"trustDomain" yaml:"trustDomain"`
	BundleURL   string `json:"bundleURL" yaml:"bundleURL"`
}

type GitOpsSpec struct {
	ArgoCD            ArgoCDSpec    `json:"argoCD,omitempty" yaml:"argoCD,omitempty"`
	Repository        RepositoryRef `json:"repository,omitempty" yaml:"repository,omitempty"`
	PrivateRepository bool          `json:"privateRepository,omitempty" yaml:"privateRepository,omitempty"`
	AutomatedSync     bool          `json:"automatedSync,omitempty" yaml:"automatedSync,omitempty"`
	SelfHeal          bool          `json:"selfHeal,omitempty" yaml:"selfHeal,omitempty"`
	Prune             bool          `json:"prune,omitempty" yaml:"prune,omitempty"`
}

type GitOpsBinding struct {
	Project     string `json:"project,omitempty" yaml:"project,omitempty"`
	Destination string `json:"destination,omitempty" yaml:"destination,omitempty"`
}

type ArgoCDSpec struct {
	Enabled         bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Version         string `json:"version,omitempty" yaml:"version,omitempty"`
	Namespace       string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	SSOWithKeycloak bool   `json:"ssoWithKeycloak,omitempty" yaml:"ssoWithKeycloak,omitempty"`
}

type RepositoryRef struct {
	URL             string `json:"url,omitempty" yaml:"url,omitempty"`
	Revision        string `json:"revision,omitempty" yaml:"revision,omitempty"`
	Path            string `json:"path,omitempty" yaml:"path,omitempty"`
	CredentialsRef  string `json:"credentialsRef,omitempty" yaml:"credentialsRef,omitempty"`
	SignaturePolicy string `json:"signaturePolicy,omitempty" yaml:"signaturePolicy,omitempty"`
}

type SecretsProvider string

const (
	SecretsProviderExternalSecrets SecretsProvider = "external-secrets"
	SecretsProviderVault           SecretsProvider = "vault"
	SecretsProviderAWS             SecretsProvider = "aws-secrets-manager"
	SecretsProviderAzure           SecretsProvider = "azure-key-vault"
	SecretsProviderGoogle          SecretsProvider = "google-secret-manager"
	SecretsProviderYandex          SecretsProvider = "yandex-lockbox"
	SecretsProviderSOPS            SecretsProvider = "sops-age"
	SecretsProviderKubernetes      SecretsProvider = "kubernetes"
)

type SecretsSpec struct {
	Provider                SecretsProvider `json:"provider,omitempty" yaml:"provider,omitempty"`
	ExternalSecretsOperator OperatorSpec    `json:"externalSecretsOperator,omitempty" yaml:"externalSecretsOperator,omitempty"`
	StoreRef                string          `json:"storeRef,omitempty" yaml:"storeRef,omitempty"`
	AllowKubernetesSecrets  bool            `json:"allowKubernetesSecrets,omitempty" yaml:"allowKubernetesSecrets,omitempty"`
}

type DatabasesSpec struct {
	PostgreSQL PostgreSQLSpec `json:"postgresql,omitempty" yaml:"postgresql,omitempty"`
	YDB        YDBSpec        `json:"ydb,omitempty" yaml:"ydb,omitempty"`
	Redis      RedisSpec      `json:"redis,omitempty" yaml:"redis,omitempty"`
}

type PostgreSQLSpec struct {
	Mode              ComponentMode        `json:"mode,omitempty" yaml:"mode,omitempty"`
	External          ExternalEndpointSpec `json:"external,omitempty" yaml:"external,omitempty"`
	HA                bool                 `json:"ha,omitempty" yaml:"ha,omitempty"`
	Instances         int32                `json:"instances,omitempty" yaml:"instances,omitempty"`
	WALArchiving      bool                 `json:"walArchiving,omitempty" yaml:"walArchiving,omitempty"`
	PITR              bool                 `json:"pitr,omitempty" yaml:"pitr,omitempty"`
	ConnectionPooling bool                 `json:"connectionPooling,omitempty" yaml:"connectionPooling,omitempty"`
	TLS               bool                 `json:"tls,omitempty" yaml:"tls,omitempty"`
	Monitoring        bool                 `json:"monitoring,omitempty" yaml:"monitoring,omitempty"`
	Backup            ComponentBackupSpec  `json:"backup,omitempty" yaml:"backup,omitempty"`
}

type YDBSpec struct {
	Mode                ComponentMode        `json:"mode,omitempty" yaml:"mode,omitempty"`
	External            ExternalEndpointSpec `json:"external,omitempty" yaml:"external,omitempty"`
	StorageTopology     string               `json:"storageTopology,omitempty" yaml:"storageTopology,omitempty"`
	ReplicationFactor   int32                `json:"replicationFactor,omitempty" yaml:"replicationFactor,omitempty"`
	FailureDomains      []string             `json:"failureDomains,omitempty" yaml:"failureDomains,omitempty"`
	PodDisruptionBudget bool                 `json:"podDisruptionBudget,omitempty" yaml:"podDisruptionBudget,omitempty"`
	TLS                 bool                 `json:"tls,omitempty" yaml:"tls,omitempty"`
	Backup              ComponentBackupSpec  `json:"backup,omitempty" yaml:"backup,omitempty"`
}

type RedisSpec struct {
	Mode     ComponentMode        `json:"mode,omitempty" yaml:"mode,omitempty"`
	External ExternalEndpointSpec `json:"external,omitempty" yaml:"external,omitempty"`
	TLS      bool                 `json:"tls,omitempty" yaml:"tls,omitempty"`
}

type MessagingSpec struct {
	Kafka KafkaSpec `json:"kafka,omitempty" yaml:"kafka,omitempty"`
}

type KafkaSpec struct {
	Mode          ComponentMode        `json:"mode,omitempty" yaml:"mode,omitempty"`
	External      ExternalEndpointSpec `json:"external,omitempty" yaml:"external,omitempty"`
	KRaft         bool                 `json:"kraft,omitempty" yaml:"kraft,omitempty"`
	NodePools     bool                 `json:"nodePools,omitempty" yaml:"nodePools,omitempty"`
	TLS           bool                 `json:"tls,omitempty" yaml:"tls,omitempty"`
	SASL          bool                 `json:"sasl,omitempty" yaml:"sasl,omitempty"`
	Quotas        bool                 `json:"quotas,omitempty" yaml:"quotas,omitempty"`
	RackAwareness bool                 `json:"rackAwareness,omitempty" yaml:"rackAwareness,omitempty"`
}

type WorkflowsSpec struct {
	Temporal TemporalSpec `json:"temporal,omitempty" yaml:"temporal,omitempty"`
}

type TemporalSpec struct {
	Mode             ComponentMode        `json:"mode,omitempty" yaml:"mode,omitempty"`
	External         ExternalEndpointSpec `json:"external,omitempty" yaml:"external,omitempty"`
	Namespaces       []string             `json:"namespaces,omitempty" yaml:"namespaces,omitempty"`
	Retention        Duration             `json:"retention,omitempty" yaml:"retention,omitempty"`
	SearchAttributes []string             `json:"searchAttributes,omitempty" yaml:"searchAttributes,omitempty"`
	Archival         bool                 `json:"archival,omitempty" yaml:"archival,omitempty"`
	MTLS             bool                 `json:"mTLS,omitempty" yaml:"mTLS,omitempty"`
	VisibilityStore  string               `json:"visibilityStore,omitempty" yaml:"visibilityStore,omitempty"`
	SmokeWorkflow    bool                 `json:"smokeWorkflow,omitempty" yaml:"smokeWorkflow,omitempty"`
}

type IdentitySpec struct {
	Keycloak KeycloakSpec `json:"keycloak,omitempty" yaml:"keycloak,omitempty"`
}

type KeycloakSpec struct {
	Mode              ComponentMode `json:"mode,omitempty" yaml:"mode,omitempty"`
	Realm             string        `json:"realm,omitempty" yaml:"realm,omitempty"`
	Operator          OperatorSpec  `json:"operator,omitempty" yaml:"operator,omitempty"`
	AdminSecretRef    string        `json:"adminSecretRef,omitempty" yaml:"adminSecretRef,omitempty"`
	WebAuthn          bool          `json:"webAuthn,omitempty" yaml:"webAuthn,omitempty"`
	OTP               bool          `json:"otp,omitempty" yaml:"otp,omitempty"`
	CIBA              bool          `json:"ciba,omitempty" yaml:"ciba,omitempty"`
	StepUp            bool          `json:"stepUp,omitempty" yaml:"stepUp,omitempty"`
	IdentityProviders []string      `json:"identityProviders,omitempty" yaml:"identityProviders,omitempty"`
}

type AuthorizationSpec struct {
	SpiceDB SpiceDBSpec `json:"spiceDB,omitempty" yaml:"spiceDB,omitempty"`
}

type SpiceDBSpec struct {
	Mode            ComponentMode `json:"mode,omitempty" yaml:"mode,omitempty"`
	Operator        OperatorSpec  `json:"operator,omitempty" yaml:"operator,omitempty"`
	Datastore       string        `json:"datastore,omitempty" yaml:"datastore,omitempty"`
	SchemaRef       string        `json:"schemaRef,omitempty" yaml:"schemaRef,omitempty"`
	Replicas        int32         `json:"replicas,omitempty" yaml:"replicas,omitempty"`
	TLS             bool          `json:"tls,omitempty" yaml:"tls,omitempty"`
	SPIFFEIdentity  bool          `json:"spiffeIdentity,omitempty" yaml:"spiffeIdentity,omitempty"`
	Consistency     string        `json:"consistency,omitempty" yaml:"consistency,omitempty"`
	Metrics         bool          `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	BackupDatastore bool          `json:"backupDatastore,omitempty" yaml:"backupDatastore,omitempty"`
}

type ObservabilitySpec struct {
	Mode                  ComponentMode `json:"mode,omitempty" yaml:"mode,omitempty"`
	OpenTelemetryOperator OperatorSpec  `json:"openTelemetryOperator,omitempty" yaml:"openTelemetryOperator,omitempty"`
	CollectorAgent        bool          `json:"collectorAgent,omitempty" yaml:"collectorAgent,omitempty"`
	CollectorGateway      bool          `json:"collectorGateway,omitempty" yaml:"collectorGateway,omitempty"`
	PrometheusOperator    OperatorSpec  `json:"prometheusOperator,omitempty" yaml:"prometheusOperator,omitempty"`
	Prometheus            bool          `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
	Alertmanager          bool          `json:"alertmanager,omitempty" yaml:"alertmanager,omitempty"`
	Grafana               bool          `json:"grafana,omitempty" yaml:"grafana,omitempty"`
	Loki                  bool          `json:"loki,omitempty" yaml:"loki,omitempty"`
	Tempo                 bool          `json:"tempo,omitempty" yaml:"tempo,omitempty"`
	KubeStateMetrics      bool          `json:"kubeStateMetrics,omitempty" yaml:"kubeStateMetrics,omitempty"`
	NodeExporter          bool          `json:"nodeExporter,omitempty" yaml:"nodeExporter,omitempty"`
	BlackboxExporter      bool          `json:"blackboxExporter,omitempty" yaml:"blackboxExporter,omitempty"`
	ExternalOTLPEndpoint  string        `json:"externalOTLPEndpoint,omitempty" yaml:"externalOTLPEndpoint,omitempty"`
	ExternalBackend       string        `json:"externalBackend,omitempty" yaml:"externalBackend,omitempty"`
}

type SecuritySpec struct {
	Kyverno               OperatorSpec          `json:"kyverno,omitempty" yaml:"kyverno,omitempty"`
	PodSecurityStandards  bool                  `json:"podSecurityStandards,omitempty" yaml:"podSecurityStandards,omitempty"`
	RunAsNonRootRequired  bool                  `json:"runAsNonRootRequired,omitempty" yaml:"runAsNonRootRequired,omitempty"`
	ReadOnlyRootFS        bool                  `json:"readOnlyRootFilesystem,omitempty" yaml:"readOnlyRootFilesystem,omitempty"`
	RequireRequestsLimits bool                  `json:"requireRequestsLimits,omitempty" yaml:"requireRequestsLimits,omitempty"`
	ForbidPrivileged      bool                  `json:"forbidPrivileged,omitempty" yaml:"forbidPrivileged,omitempty"`
	ForbidHostPath        bool                  `json:"forbidHostPath,omitempty" yaml:"forbidHostPath,omitempty"`
	ForbidHostNetwork     bool                  `json:"forbidHostNetwork,omitempty" yaml:"forbidHostNetwork,omitempty"`
	TrustedRegistries     []string              `json:"trustedRegistries,omitempty" yaml:"trustedRegistries,omitempty"`
	ForbidLatestTag       bool                  `json:"forbidLatestTag,omitempty" yaml:"forbidLatestTag,omitempty"`
	Cosign                CosignSpec            `json:"cosign,omitempty" yaml:"cosign,omitempty"`
	SBOMRequired          bool                  `json:"sbomRequired,omitempty" yaml:"sbomRequired,omitempty"`
	TrivyOperator         OperatorSpec          `json:"trivyOperator,omitempty" yaml:"trivyOperator,omitempty"`
	AuditPolicies         bool                  `json:"auditPolicies,omitempty" yaml:"auditPolicies,omitempty"`
	Exceptions            []PolicyExceptionSpec `json:"exceptions,omitempty" yaml:"exceptions,omitempty"`
}

type CosignSpec struct {
	Enforce      bool   `json:"enforce,omitempty" yaml:"enforce,omitempty"`
	PublicKeyRef string `json:"publicKeyRef,omitempty" yaml:"publicKeyRef,omitempty"`
}

type PolicyExceptionSpec struct {
	Name      string      `json:"name" yaml:"name"`
	Reason    string      `json:"reason" yaml:"reason"`
	ExpiresAt metav1.Time `json:"expiresAt" yaml:"expiresAt"`
}

type BackupSpec struct {
	Enabled        bool               `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Velero         OperatorSpec       `json:"velero,omitempty" yaml:"velero,omitempty"`
	ObjectStorage  ObjectStorageSpec  `json:"objectStorage,omitempty" yaml:"objectStorage,omitempty"`
	Schedule       string             `json:"schedule,omitempty" yaml:"schedule,omitempty"`
	Retention      Duration           `json:"retention,omitempty" yaml:"retention,omitempty"`
	NativeBackups  []NativeBackupSpec `json:"nativeBackups,omitempty" yaml:"nativeBackups,omitempty"`
	VerifyOnCreate bool               `json:"verifyOnCreate,omitempty" yaml:"verifyOnCreate,omitempty"`
}

type ComponentBackupSpec struct {
	Enabled   bool     `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Schedule  string   `json:"schedule,omitempty" yaml:"schedule,omitempty"`
	Retention Duration `json:"retention,omitempty" yaml:"retention,omitempty"`
}

type NativeBackupSpec struct {
	Component string `json:"component" yaml:"component"`
	Enabled   bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}

type ObjectStorageSpec struct {
	Endpoint       string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Bucket         string `json:"bucket,omitempty" yaml:"bucket,omitempty"`
	Region         string `json:"region,omitempty" yaml:"region,omitempty"`
	CredentialsRef string `json:"credentialsRef,omitempty" yaml:"credentialsRef,omitempty"`
}

type ModulesSpec struct {
	Console         bool `json:"console,omitempty" yaml:"console,omitempty"`
	ResourceManager bool `json:"resourceManager,omitempty" yaml:"resourceManager,omitempty"`
	Identity        bool `json:"identity,omitempty" yaml:"identity,omitempty"`
	Authentication  bool `json:"authentication,omitempty" yaml:"authentication,omitempty"`
	Access          bool `json:"access,omitempty" yaml:"access,omitempty"`
	Gateway         bool `json:"gateway,omitempty" yaml:"gateway,omitempty"`
	PolicyStudio    bool `json:"policyStudio,omitempty" yaml:"policyStudio,omitempty"`
	Audit           bool `json:"audit,omitempty" yaml:"audit,omitempty"`
	Operations      bool `json:"operations,omitempty" yaml:"operations,omitempty"`
	Idempotency     bool `json:"idempotency,omitempty" yaml:"idempotency,omitempty"`
	Health          bool `json:"health,omitempty" yaml:"health,omitempty"`
	Alerting        bool `json:"alerting,omitempty" yaml:"alerting,omitempty"`
	Notifications   bool `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	Provisioning    bool `json:"provisioning,omitempty" yaml:"provisioning,omitempty"`
}

type UpgradeSpec struct {
	Channel               string   `json:"channel,omitempty" yaml:"channel,omitempty"`
	AllowedWindows        []string `json:"allowedWindows,omitempty" yaml:"allowedWindows,omitempty"`
	RequireBackup         bool     `json:"requireBackup,omitempty" yaml:"requireBackup,omitempty"`
	AutoApplyPatchUpdates bool     `json:"autoApplyPatchUpdates,omitempty" yaml:"autoApplyPatchUpdates,omitempty"`
}

type AirGapSpec struct {
	Enabled              bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	BundleRef            string `json:"bundleRef,omitempty" yaml:"bundleRef,omitempty"`
	PrivateRegistry      string `json:"privateRegistry,omitempty" yaml:"privateRegistry,omitempty"`
	PrivateGitRepository string `json:"privateGitRepository,omitempty" yaml:"privateGitRepository,omitempty"`
	PrivateOCIRegistry   string `json:"privateOCIRegistry,omitempty" yaml:"privateOCIRegistry,omitempty"`
	InternalCARef        string `json:"internalCARef,omitempty" yaml:"internalCARef,omitempty"`
}

type OperatorSpec struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

type ExternalEndpointSpec struct {
	Endpoint       string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	CredentialsRef string `json:"credentialsRef,omitempty" yaml:"credentialsRef,omitempty"`
	TLS            bool   `json:"tls,omitempty" yaml:"tls,omitempty"`
}

type PlatformInstallationStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty" yaml:"observedGeneration,omitempty"`
	Phase              Phase              `json:"phase,omitempty" yaml:"phase,omitempty"`
	PlatformVersion    string             `json:"platformVersion,omitempty" yaml:"platformVersion,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Components         []ComponentStatus  `json:"components,omitempty" yaml:"components,omitempty"`
	Endpoints          []EndpointStatus   `json:"endpoints,omitempty" yaml:"endpoints,omitempty"`
	LastOperation      *OperationRef      `json:"lastOperation,omitempty" yaml:"lastOperation,omitempty"`
}

type ComponentStatus struct {
	Name       string             `json:"name" yaml:"name"`
	Namespace  string             `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Version    string             `json:"version,omitempty" yaml:"version,omitempty"`
	Phase      Phase              `json:"phase,omitempty" yaml:"phase,omitempty"`
	Ready      bool               `json:"ready,omitempty" yaml:"ready,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	LastError  string             `json:"lastError,omitempty" yaml:"lastError,omitempty"`
}

type EndpointStatus struct {
	Name     string `json:"name" yaml:"name"`
	Type     string `json:"type" yaml:"type"`
	URL      string `json:"url,omitempty" yaml:"url,omitempty"`
	Internal bool   `json:"internal,omitempty" yaml:"internal,omitempty"`
	Ready    bool   `json:"ready,omitempty" yaml:"ready,omitempty"`
}

type OperationRef struct {
	Name      string      `json:"name" yaml:"name"`
	Type      string      `json:"type" yaml:"type"`
	Phase     Phase       `json:"phase" yaml:"phase"`
	StartedAt metav1.Time `json:"startedAt" yaml:"startedAt"`
}

type PlatformRelease struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   PlatformReleaseSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status PlatformReleaseStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type PlatformReleaseSpec struct {
	Kubernetes    VersionRange                `json:"kubernetes" yaml:"kubernetes"`
	Components    map[string]ComponentRelease `json:"components" yaml:"components"`
	Security      ReleaseSecurityMetadata     `json:"security,omitempty" yaml:"security,omitempty"`
	Compatibility CompatibilityMetadata       `json:"compatibility,omitempty" yaml:"compatibility,omitempty"`
	ReleaseNotes  string                      `json:"releaseNotes,omitempty" yaml:"releaseNotes,omitempty"`
	Signature     SignatureRef                `json:"signature,omitempty" yaml:"signature,omitempty"`
}

type VersionRange struct {
	MinVersion string `json:"minVersion" yaml:"minVersion"`
	MaxVersion string `json:"maxVersion,omitempty" yaml:"maxVersion,omitempty"`
}

type ComponentRelease struct {
	Version                 string                    `json:"version" yaml:"version"`
	Chart                   ArtifactRef               `json:"chart,omitempty" yaml:"chart,omitempty"`
	Images                  []ArtifactRef             `json:"images,omitempty" yaml:"images,omitempty"`
	CRDVersions             []string                  `json:"crdVersions,omitempty" yaml:"crdVersions,omitempty"`
	Dependencies            []string                  `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	SupportedUpgradePaths   []UpgradePath             `json:"supportedUpgradePaths,omitempty" yaml:"supportedUpgradePaths,omitempty"`
	RollbackSupport         RollbackSupport           `json:"rollbackSupport,omitempty" yaml:"rollbackSupport,omitempty"`
	MigrationRequirements   []MigrationRequirement    `json:"migrationRequirements,omitempty" yaml:"migrationRequirements,omitempty"`
	KubernetesCompatibility VersionRange              `json:"kubernetesCompatibility,omitempty" yaml:"kubernetesCompatibility,omitempty"`
	Resources               ResourceEstimate          `json:"resources,omitempty" yaml:"resources,omitempty"`
	Security                ComponentSecurityMetadata `json:"security,omitempty" yaml:"security,omitempty"`
	ReleaseNotes            string                    `json:"releaseNotes,omitempty" yaml:"releaseNotes,omitempty"`
}

type ArtifactRef struct {
	Name       string `json:"name,omitempty" yaml:"name,omitempty"`
	Repository string `json:"repository" yaml:"repository"`
	Version    string `json:"version,omitempty" yaml:"version,omitempty"`
	Digest     string `json:"digest" yaml:"digest"`
	Signature  string `json:"signature,omitempty" yaml:"signature,omitempty"`
	SBOM       string `json:"sbom,omitempty" yaml:"sbom,omitempty"`
}

type UpgradePath struct {
	From           string `json:"from" yaml:"from"`
	To             string `json:"to" yaml:"to"`
	Classification string `json:"classification" yaml:"classification"`
}

type RollbackSupport struct {
	Supported bool   `json:"supported" yaml:"supported"`
	Boundary  string `json:"boundary,omitempty" yaml:"boundary,omitempty"`
}

type MigrationRequirement struct {
	Name         string `json:"name" yaml:"name"`
	Component    string `json:"component,omitempty" yaml:"component,omitempty"`
	Irreversible bool   `json:"irreversible,omitempty" yaml:"irreversible,omitempty"`
}

type ResourceEstimate struct {
	CPU     string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory  string `json:"memory,omitempty" yaml:"memory,omitempty"`
	Storage string `json:"storage,omitempty" yaml:"storage,omitempty"`
}

type ReleaseSecurityMetadata struct {
	SBOMLocation     string `json:"sbomLocation,omitempty" yaml:"sbomLocation,omitempty"`
	CosignSignature  string `json:"cosignSignature,omitempty" yaml:"cosignSignature,omitempty"`
	VulnerabilityRef string `json:"vulnerabilityRef,omitempty" yaml:"vulnerabilityRef,omitempty"`
}

type ComponentSecurityMetadata struct {
	SBOMLocation     string `json:"sbomLocation,omitempty" yaml:"sbomLocation,omitempty"`
	CosignSignature  string `json:"cosignSignature,omitempty" yaml:"cosignSignature,omitempty"`
	VulnerabilityRef string `json:"vulnerabilityRef,omitempty" yaml:"vulnerabilityRef,omitempty"`
}

type CompatibilityMetadata struct {
	SupportedProfiles []Profile `json:"supportedProfiles,omitempty" yaml:"supportedProfiles,omitempty"`
}

type SignatureRef struct {
	Algorithm string `json:"algorithm,omitempty" yaml:"algorithm,omitempty"`
	Value     string `json:"value,omitempty" yaml:"value,omitempty"`
	KeyRef    string `json:"keyRef,omitempty" yaml:"keyRef,omitempty"`
}

type PlatformReleaseStatus struct {
	Verified   bool               `json:"verified,omitempty" yaml:"verified,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}

type InstallationOperation struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   InstallationOperationSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status InstallationOperationStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type OperationType string

const (
	OperationTypeBootstrap OperationType = "bootstrap"
	OperationTypeInstall   OperationType = "install"
	OperationTypeUpgrade   OperationType = "upgrade"
	OperationTypeRollback  OperationType = "rollback"
	OperationTypeBackup    OperationType = "backup"
	OperationTypeRestore   OperationType = "restore"
	OperationTypeUninstall OperationType = "uninstall"
)

type InstallationOperationSpec struct {
	OperationID      string        `json:"operationID" yaml:"operationID"`
	Type             OperationType `json:"type" yaml:"type"`
	InstallationRef  string        `json:"installationRef" yaml:"installationRef"`
	RequestedVersion string        `json:"requestedVersion,omitempty" yaml:"requestedVersion,omitempty"`
	PlanDigest       string        `json:"planDigest,omitempty" yaml:"planDigest,omitempty"`
	RequestedBy      string        `json:"requestedBy,omitempty" yaml:"requestedBy,omitempty"`
}

type InstallationOperationStatus struct {
	Phase          Phase                 `json:"phase,omitempty" yaml:"phase,omitempty"`
	StartedAt      *metav1.Time          `json:"startedAt,omitempty" yaml:"startedAt,omitempty"`
	CompletedAt    *metav1.Time          `json:"completedAt,omitempty" yaml:"completedAt,omitempty"`
	CurrentStep    string                `json:"currentStep,omitempty" yaml:"currentStep,omitempty"`
	CompletedSteps []string              `json:"completedSteps,omitempty" yaml:"completedSteps,omitempty"`
	FailedStep     string                `json:"failedStep,omitempty" yaml:"failedStep,omitempty"`
	RetryCount     int32                 `json:"retryCount,omitempty" yaml:"retryCount,omitempty"`
	Error          *OperationErrorStatus `json:"error,omitempty" yaml:"error,omitempty"`
	Checkpoints    []OperationCheckpoint `json:"checkpoints,omitempty" yaml:"checkpoints,omitempty"`
	LogsRef        string                `json:"logsRef,omitempty" yaml:"logsRef,omitempty"`
}

type OperationErrorStatus struct {
	Code        string `json:"code" yaml:"code"`
	Message     string `json:"message" yaml:"message"`
	Transient   bool   `json:"transient,omitempty" yaml:"transient,omitempty"`
	Remediation string `json:"remediation,omitempty" yaml:"remediation,omitempty"`
}

type OperationCheckpoint struct {
	Name      string      `json:"name" yaml:"name"`
	StepID    string      `json:"stepID" yaml:"stepID"`
	Digest    string      `json:"digest,omitempty" yaml:"digest,omitempty"`
	CreatedAt metav1.Time `json:"createdAt" yaml:"createdAt"`
}

type Backup struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   BackupCRSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status BackupCRStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type BackupCRSpec struct {
	InstallationRef string   `json:"installationRef" yaml:"installationRef"`
	Components      []string `json:"components,omitempty" yaml:"components,omitempty"`
	StorageRef      string   `json:"storageRef,omitempty" yaml:"storageRef,omitempty"`
	Verify          bool     `json:"verify,omitempty" yaml:"verify,omitempty"`
	RequestedBy     string   `json:"requestedBy,omitempty" yaml:"requestedBy,omitempty"`
}

type BackupCRStatus struct {
	Phase              Phase                     `json:"phase,omitempty" yaml:"phase,omitempty"`
	StartedAt          *metav1.Time              `json:"startedAt,omitempty" yaml:"startedAt,omitempty"`
	CompletedAt        *metav1.Time              `json:"completedAt,omitempty" yaml:"completedAt,omitempty"`
	PlatformVersion    string                    `json:"platformVersion,omitempty" yaml:"platformVersion,omitempty"`
	ArtifactDigest     string                    `json:"artifactDigest,omitempty" yaml:"artifactDigest,omitempty"`
	ComponentArtifacts []ComponentBackupArtifact `json:"componentArtifacts,omitempty" yaml:"componentArtifacts,omitempty"`
	Verified           bool                      `json:"verified,omitempty" yaml:"verified,omitempty"`
	Error              *OperationErrorStatus     `json:"error,omitempty" yaml:"error,omitempty"`
}

type ComponentBackupArtifact struct {
	Component string `json:"component" yaml:"component"`
	Kind      string `json:"kind" yaml:"kind"`
	URI       string `json:"uri" yaml:"uri"`
	Digest    string `json:"digest,omitempty" yaml:"digest,omitempty"`
}

type RestorePlan struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec   RestorePlanSpec   `json:"spec,omitempty" yaml:"spec,omitempty"`
	Status RestorePlanStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

type RestorePlanSpec struct {
	InstallationRef       string   `json:"installationRef" yaml:"installationRef"`
	BackupRef             string   `json:"backupRef" yaml:"backupRef"`
	SourcePlatformVersion string   `json:"sourcePlatformVersion,omitempty" yaml:"sourcePlatformVersion,omitempty"`
	TargetPlatformVersion string   `json:"targetPlatformVersion,omitempty" yaml:"targetPlatformVersion,omitempty"`
	MaintenanceActions    []string `json:"maintenanceActions,omitempty" yaml:"maintenanceActions,omitempty"`
	ReplaceResources      []string `json:"replaceResources,omitempty" yaml:"replaceResources,omitempty"`
	ExpectedRPO           Duration `json:"expectedRPO,omitempty" yaml:"expectedRPO,omitempty"`
	ExpectedRTO           Duration `json:"expectedRTO,omitempty" yaml:"expectedRTO,omitempty"`
	Risks                 []string `json:"risks,omitempty" yaml:"risks,omitempty"`
}

type RestorePlanStatus struct {
	Phase      Phase                 `json:"phase,omitempty" yaml:"phase,omitempty"`
	Approved   bool                  `json:"approved,omitempty" yaml:"approved,omitempty"`
	Verified   bool                  `json:"verified,omitempty" yaml:"verified,omitempty"`
	Conditions []metav1.Condition    `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Error      *OperationErrorStatus `json:"error,omitempty" yaml:"error,omitempty"`
}

type Duration = metav1.Duration
