
// GENERATED FILE -- DO NOT EDIT

package annotation

type ResourceTypes int

const (
	Unknown ResourceTypes = iota
    Any
    Ingress
    Pod
    Service
    ServiceEntry
)

func (r ResourceTypes) String() string {
	switch r {
	case 1:
		return "Any"
	case 2:
		return "Ingress"
	case 3:
		return "Pod"
	case 4:
		return "Service"
	case 5:
		return "ServiceEntry"
	}
	return "Unknown"
}

// Instance describes a single resource annotation
type Instance struct {
	// The name of the annotation.
	Name string

	// Description of the annotation.
	Description string

	// Hide the existence of this annotation when outputting usage information.
	Hidden bool

	// Mark this annotation as deprecated when generating usage information.
	Deprecated bool

	// The types of resources this annotation applies to.
	Resources []ResourceTypes
}

var (

		AlphaCanonicalServiceAccounts = Instance {
          Name: "alpha.polarismesh.cn/canonical-serviceaccounts",
          Description: "Specifies the non-Kubernetes service accounts that are "+
                        "allowed to run this service. NOTE This API is Alpha and "+
                        "has no stability guarantees.",
          Hidden: true,
          Deprecated: false,
		  Resources: []ResourceTypes{ Service, },
        }

		AlphaIdentity = Instance {
          Name: "alpha.polarismesh.cn/identity",
          Description: "Identity for the workload. NOTE This API is Alpha and has "+
                        "no stability guarantees.",
          Hidden: true,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		AlphaKubernetesServiceAccounts = Instance {
          Name: "alpha.polarismesh.cn/kubernetes-serviceaccounts",
          Description: "Specifies the Kubernetes service accounts that are "+
                        "allowed to run this service on the VMs. NOTE This API is "+
                        "Alpha and has no stability guarantees.",
          Hidden: true,
          Deprecated: false,
		  Resources: []ResourceTypes{ Service, },
        }

		GalleyAnalyzeSuppress = Instance {
          Name: "galley.polarismesh.cn/analyze-suppress",
          Description: "A comma separated list of configuration analysis message "+
                        "codes to suppress when Istio analyzers are run. For "+
                        "example, to suppress reporting of IST0103 "+
                        "(PodMissingProxy) and IST0108 (UnknownAnnotation) on a "+
                        "resource, apply the annotation "+
                        "'galley.polarismesh.cn/analyze-suppress=IST0108,IST0103'. If "+
                        "the value is '*', then all configuration analysis "+
                        "messages are suppressed.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Any, },
        }

		OperatorInstallChartOwner = Instance {
          Name: "install.operator.polarismesh.cn/chart-owner",
          Description: "Represents the name of the chart used to create this "+
                        "resource.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Any, },
        }

		OperatorInstallOwnerGeneration = Instance {
          Name: "install.operator.polarismesh.cn/owner-generation",
          Description: "Represents the generation to which the resource was last "+
                        "reconciled.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Any, },
        }

		OperatorInstallVersion = Instance {
          Name: "install.operator.polarismesh.cn/version",
          Description: "Represents the Istio version associated with the resource",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Any, },
        }

		IoKubernetesIngressClass = Instance {
          Name: "kubernetes.io/ingress.class",
          Description: "Annotation on an Ingress resources denoting the class of "+
                        "controllers responsible for it.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Ingress, },
        }

		AlphaNetworkingEndpointsVersion = Instance {
          Name: "networking.alpha.po la ri s me sh/endpointsVersion",
          Description: "Added to synthetic ServiceEntry resources to provide the "+
                        "raw resource version from the most recent Kubernetes "+
                        "endpoints update (if available). NOTE This API is Alpha "+
                        "and has no stability guarantees.",
          Hidden: true,
          Deprecated: false,
		  Resources: []ResourceTypes{ ServiceEntry, },
        }

		AlphaNetworkingNotReadyEndpoints = Instance {
          Name: "networking.alpha.polarismesh.cn/notReadyEndpoints",
          Description: "Added to synthetic ServiceEntry resources to provide the "+
                        "'NotReadyAddresses' from the Kubernetes Endpoints "+
                        "resource. The value is a comma-separated list of IP:port. "+
                        "NOTE This API is Alpha and has no stability guarantees.",
          Hidden: true,
          Deprecated: false,
		  Resources: []ResourceTypes{ ServiceEntry, },
        }

		AlphaNetworkingServiceVersion = Instance {
          Name: "networking.alpha.polarismesh.cn/serviceVersion",
          Description: "Added to synthetic ServiceEntry resources to provide the "+
                        "raw resource version from the most recent Kubernetes "+
                        "service update. This will always be available for "+
                        "synthetic service entries. NOTE This API is Alpha and has "+
                        "no stability guarantees.",
          Hidden: true,
          Deprecated: false,
		  Resources: []ResourceTypes{ ServiceEntry, },
        }

		NetworkingExportTo = Instance {
          Name: "networking.polarismesh.cn/exportTo",
          Description: "Specifies the namespaces to which this service should be "+
                        "exported to. A value of '*' indicates it is reachable "+
                        "within the mesh '.' indicates it is reachable within its "+
                        "namespace.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Service, },
        }

		PolicyCheck = Instance {
          Name: "policy.polarismesh.cn/check",
          Description: "Determines the policy for behavior when unable to connect "+
                        "to Mixer. If not set, FAIL_CLOSE is set, rejecting "+
                        "requests.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		PolicyCheckBaseRetryWaitTime = Instance {
          Name: "policy.polarismesh.cn/checkBaseRetryWaitTime",
          Description: "Base time to wait between retries, will be adjusted by "+
                        "backoff and jitter. In duration format. If not set, this "+
                        "will be 80ms.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		PolicyCheckMaxRetryWaitTime = Instance {
          Name: "policy.polarismesh.cn/checkMaxRetryWaitTime",
          Description: "Maximum time to wait between retries to Mixer. In "+
                        "duration format. If not set, this will be 1000ms.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		PolicyCheckRetries = Instance {
          Name: "policy.polarismesh.cn/checkRetries",
          Description: "The maximum number of retries on transport errors to "+
                        "Mixer. If not set, this will be 0, indicating no retries.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		PolicyLang = Instance {
          Name: "policy.polarismesh.cn/lang",
          Description: "Selects the attribute expression language runtime for "+
                        "Mixer.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarStatusReadinessApplicationPorts = Instance {
          Name: "readiness.status.sidecar.polarismesh.cn/applicationPorts",
          Description: "Specifies the list of ports exposed by the application "+
                        "container. Used by the Envoy sidecar readiness probe to "+
                        "determine that Envoy is configured and ready to receive "+
                        "traffic.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarStatusReadinessFailureThreshold = Instance {
          Name: "readiness.status.sidecar.polarismesh.cn/failureThreshold",
          Description: "Specifies the failure threshold for the Envoy sidecar "+
                        "readiness probe.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarStatusReadinessInitialDelaySeconds = Instance {
          Name: "readiness.status.sidecar.polarismesh.cn/initialDelaySeconds",
          Description: "Specifies the initial delay (in seconds) for the Envoy "+
                        "sidecar readiness probe.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarStatusReadinessPeriodSeconds = Instance {
          Name: "readiness.status.sidecar.polarismesh.cn/periodSeconds",
          Description: "Specifies the period (in seconds) for the Envoy sidecar "+
                        "readiness probe.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SecurityAutoMTLS = Instance {
          Name: "security.polarismesh.cn/autoMTLS",
          Description: "Determines whether the client proxy uses auto mTLS. This "+
                        "overrides the mesh default specified in "+
                        "MeshConfig.enable_auto_mtls.",
          Hidden: true,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarBootstrapOverride = Instance {
          Name: "sidecar.polarismesh.cn/bootstrapOverride",
          Description: "Specifies an alternative Envoy bootstrap configuration "+
                        "file.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarComponentLogLevel = Instance {
          Name: "sidecar.polarismesh.cn/componentLogLevel",
          Description: "Specifies the component log level for Envoy.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarControlPlaneAuthPolicy = Instance {
          Name: "sidecar.polarismesh.cn/controlPlaneAuthPolicy",
          Description: "Specifies the auth policy used by the Istio control "+
                        "plane. If NONE, traffic will not be encrypted. If "+
                        "MUTUAL_TLS, traffic between Envoy sidecar will be wrapped "+
                        "into mutual TLS connections.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarDiscoveryAddress = Instance {
          Name: "sidecar.polarismesh.cn/discoveryAddress",
          Description: "Specifies the XDS discovery address to be used by the "+
                        "Envoy sidecar.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarEnableCoreDump = Instance {
          Name: "sidecar.polarismesh.cn/enableCoreDump",
          Description: "Specifies whether or not an Envoy sidecar should enable "+
                        "core dump.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarInject = Instance {
          Name: "sidecar.polarismesh.cn/inject",
          Description: "Specifies whether or not an Polaris sidecar should be "+
                        "automatically injected into the workload.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarInterceptionMode = Instance {
          Name: "sidecar.polarismesh.cn/interceptionMode",
          Description: "Specifies the mode used to redirect inbound connections "+
                        "to Envoy (REDIRECT or TPROXY).",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarLogLevel = Instance {
          Name: "sidecar.polarismesh.cn/logLevel",
          Description: "Specifies the log level for Envoy.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarProxyCPU = Instance {
          Name: "sidecar.polarismesh.cn/proxyCPU",
          Description: "Specifies the requested CPU setting for the Envoy "+
                        "sidecar.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarProxyImage = Instance {
          Name: "sidecar.polarismesh.cn/proxyImage",
          Description: "Specifies the Docker image to be used by the Envoy "+
                        "sidecar.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarProxyMemory = Instance {
          Name: "sidecar.polarismesh.cn/proxyMemory",
          Description: "Specifies the requested memory setting for the Envoy "+
                        "sidecar.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarRewriteAppHTTPProbers = Instance {
          Name: "sidecar.polarismesh.cn/rewriteAppHTTPProbers",
          Description: "Rewrite HTTP readiness and liveness probes to be "+
                        "redirected to the Envoy sidecar.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarStatsInclusionPrefixes = Instance {
          Name: "sidecar.polarismesh.cn/statsInclusionPrefixes",
          Description: "Specifies the comma separated list of prefixes of the "+
                        "stats to be emitted by Envoy.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarStatsInclusionRegexps = Instance {
          Name: "sidecar.polarismesh.cn/statsInclusionRegexps",
          Description: "Specifies the comma separated list of regexes the stats "+
                        "should match to be emitted by Envoy.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarStatsInclusionSuffixes = Instance {
          Name: "sidecar.polarismesh.cn/statsInclusionSuffixes",
          Description: "Specifies the comma separated list of suffixes of the "+
                        "stats to be emitted by Envoy.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarStatus = Instance {
          Name: "sidecar.polarismesh.cn/status",
          Description: "Generated by Envoy sidecar injection that indicates the "+
                        "status of the operation. Includes a version hash of the "+
                        "executed template, as well as names of injected "+
                        "resources.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarUserVolume = Instance {
          Name: "sidecar.polarismesh.cn/userVolume",
          Description: "Specifies one or more user volumes (as a JSON array) to "+
                        "be added to the Envoy sidecar.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarUserVolumeMount = Instance {
          Name: "sidecar.polarismesh.cn/userVolumeMount",
          Description: "Specifies one or more user volume mounts (as a JSON "+
                        "array) to be added to the Envoy sidecar.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarStatusPort = Instance {
          Name: "status.sidecar.polarismesh.cn/port",
          Description: "Specifies the HTTP status Port for the Envoy sidecar. If "+
                        "zero, the sidecar will not provide status.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarTrafficExcludeInboundPorts = Instance {
          Name: "traffic.sidecar.polarismesh.cn/excludeInboundPorts",
          Description: "A comma separated list of inbound ports to be excluded "+
                        "from redirection to Envoy. Only applies when all inbound "+
                        "traffic (i.e. '*') is being redirected.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarTrafficExcludeOutboundIPRanges = Instance {
          Name: "traffic.sidecar.polarismesh.cn/excludeOutboundIPRanges",
          Description: "A comma separated list of IP ranges in CIDR form to be "+
                        "excluded from redirection. Only applies when all outbound "+
                        "traffic (i.e. '*') is being redirected.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarTrafficExcludeOutboundPorts = Instance {
          Name: "traffic.sidecar.polarismesh.cn/excludeOutboundPorts",
          Description: "A comma separated list of outbound ports to be excluded "+
                        "from redirection to Envoy.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarTrafficIncludeInboundPorts = Instance {
          Name: "traffic.sidecar.polarismesh.cn/includeInboundPorts",
          Description: "A comma separated list of inbound ports for which traffic "+
                        "is to be redirected to Envoy. The wildcard character '*' "+
                        "can be used to configure redirection for all ports. An "+
                        "empty list will disable all inbound redirection.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarTrafficIncludeOutboundIPRanges = Instance {
          Name: "traffic.sidecar.polarismesh.cn/includeOutboundIPRanges",
          Description: "A comma separated list of IP ranges in CIDR form to "+
                        "redirect to Envoy (optional). The wildcard character '*' "+
                        "can be used to redirect all outbound traffic. An empty "+
                        "list will disable all outbound redirection.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

		SidecarTrafficKubevirtInterfaces = Instance {
          Name: "traffic.sidecar.polarismesh.cn/kubevirtInterfaces",
          Description: "A comma separated list of virtual interfaces whose "+
                        "inbound traffic (from VM) will be treated as outbound.",
          Hidden: false,
          Deprecated: false,
		  Resources: []ResourceTypes{ Pod, },
        }

)
