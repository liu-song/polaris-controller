// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package inject

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/hashicorp/go-multierror"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/batch/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yamlDecoder "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/polarismesh/polaris-controller/common/log"
	"github.com/polarismesh/polaris-controller/pkg/inject/api/annotation"
	meshconfig "github.com/polarismesh/polaris-controller/pkg/inject/api/mesh/v1alpha1"
	"github.com/polarismesh/polaris-controller/pkg/util"
	utils "github.com/polarismesh/polaris-controller/pkg/util"
)

type annotationValidationFunc func(value string) error

// per-sidecar policy and status
var (
	alwaysValidFunc = func(value string) error {
		return nil
	}

	annotationRegistry = map[string]annotationValidationFunc{
		annotation.SidecarInject.Name:                             alwaysValidFunc,
		annotation.SidecarStatus.Name:                             alwaysValidFunc,
		annotation.SidecarRewriteAppHTTPProbers.Name:              alwaysValidFunc,
		annotation.SidecarControlPlaneAuthPolicy.Name:             alwaysValidFunc,
		annotation.SidecarDiscoveryAddress.Name:                   alwaysValidFunc,
		annotation.SidecarProxyImage.Name:                         alwaysValidFunc,
		annotation.SidecarProxyCPU.Name:                           alwaysValidFunc,
		annotation.SidecarProxyMemory.Name:                        alwaysValidFunc,
		annotation.SidecarInterceptionMode.Name:                   alwaysValidFunc,
		annotation.SidecarBootstrapOverride.Name:                  alwaysValidFunc,
		annotation.SidecarStatsInclusionPrefixes.Name:             alwaysValidFunc,
		annotation.SidecarStatsInclusionSuffixes.Name:             alwaysValidFunc,
		annotation.SidecarStatsInclusionRegexps.Name:              alwaysValidFunc,
		annotation.SidecarUserVolume.Name:                         alwaysValidFunc,
		annotation.SidecarUserVolumeMount.Name:                    alwaysValidFunc,
		annotation.SidecarEnableCoreDump.Name:                     validateBool,
		annotation.SidecarStatusPort.Name:                         validateStatusPort,
		annotation.SidecarStatusReadinessInitialDelaySeconds.Name: validateUInt32,
		annotation.SidecarStatusReadinessPeriodSeconds.Name:       validateUInt32,
		annotation.SidecarStatusReadinessFailureThreshold.Name:    validateUInt32,
		annotation.SidecarTrafficIncludeOutboundIPRanges.Name:     ValidateIncludeIPRanges,
		annotation.SidecarTrafficExcludeOutboundIPRanges.Name:     ValidateExcludeIPRanges,
		annotation.SidecarTrafficIncludeInboundPorts.Name:         ValidateIncludeInboundPorts,
		annotation.SidecarTrafficExcludeInboundPorts.Name:         ValidateExcludeInboundPorts,
		annotation.SidecarTrafficExcludeOutboundPorts.Name:        ValidateExcludeOutboundPorts,
		annotation.SidecarTrafficKubevirtInterfaces.Name:          alwaysValidFunc,
	}
)

func validateAnnotations(annotations map[string]string) (err error) {
	for name, value := range annotations {
		if v, ok := annotationRegistry[name]; ok {
			if e := v(value); e != nil {
				err = multierror.Append(err, fmt.Errorf("invalid value '%s' for annotation '%s': %v", value, name, e))
			}
		} else if strings.Contains(name, "istio") {
			log.Warnf("Potentially misspelled annotation '%s' with value '%s' encountered", name, value)
		}
	}
	return
}

// InjectionPolicy determines the policy for injecting the
// sidecar proxy into the watched namespace(s).
type InjectionPolicy string

const (
	// InjectionPolicyDisabled specifies that the sidecar injector
	// will not inject the sidecar into resources by default for the
	// namespace(s) being watched. Resources can enable injection
	// using the "sidecar.polarismesh.cn/inject" annotation with value of
	// true.
	InjectionPolicyDisabled InjectionPolicy = "disabled"

	// InjectionPolicyEnabled specifies that the sidecar injector will
	// inject the sidecar into resources by default for the
	// namespace(s) being watched. Resources can disable injection
	// using the "sidecar.polarismesh.cn/inject" annotation with value of
	// false.
	InjectionPolicyEnabled InjectionPolicy = "enabled"
)

const (
	// ProxyContainerName is used by e2e integration tests for fetching logs
	ProxyContainerName = "polaris-proxy"
)

// SidecarInjectionSpec collects all container types and volumes for
// sidecar mesh injection
type SidecarInjectionSpec struct {
	// RewriteHTTPProbe indicates whether Kubernetes HTTP prober in the PodSpec
	// will be rewritten to be redirected by pilot agent.
	PodRedirectAnnot    map[string]string             `yaml:"podRedirectAnnot"`
	RewriteAppHTTPProbe bool                          `yaml:"rewriteAppHTTPProbe"`
	InitContainers      []corev1.Container            `yaml:"initContainers"`
	Containers          []corev1.Container            `yaml:"containers"`
	Volumes             []corev1.Volume               `yaml:"volumes"`
	DNSConfig           *corev1.PodDNSConfig          `yaml:"dnsConfig"`
	ImagePullSecrets    []corev1.LocalObjectReference `yaml:"imagePullSecrets"`
}

// SidecarTemplateData is the data object to which the templated
// version of `SidecarInjectionSpec` is applied.
type SidecarTemplateData struct {
	TypeMeta       *metav1.TypeMeta
	DeploymentMeta *metav1.ObjectMeta
	ObjectMeta     *metav1.ObjectMeta
	Spec           *corev1.PodSpec
	ProxyConfig    *meshconfig.ProxyConfig
	MeshConfig     *meshconfig.MeshConfig
	Values         map[string]interface{}
}

// Config specifies the sidecar injection configuration This includes
// the sidecar template and cluster-side injection policy. It is used
// by kube-inject, sidecar injector, and http endpoint.
type Config struct {
	Policy InjectionPolicy `json:"policy"`

	// Template is the templated version of `SidecarInjectionSpec` prior to
	// expansion over the `SidecarTemplateData`.
	Template string `json:"template"`

	// NeverInjectSelector: Refuses the injection on pods whose labels match this selector.
	// It's an array of label selectors, that will be OR'ed, meaning we will iterate
	// over it and stop at the first match
	// Takes precedence over AlwaysInjectSelector.
	NeverInjectSelector []metav1.LabelSelector `json:"neverInjectSelector"`

	// AlwaysInjectSelector: Forces the injection on pods whose labels match this selector.
	// It's an array of label selectors, that will be OR'ed, meaning we will iterate
	// over it and stop at the first match
	AlwaysInjectSelector []metav1.LabelSelector `json:"alwaysInjectSelector"`

	// InjectedAnnotations are additional annotations that will be added to the pod spec after injection
	// This is primarily to support PSP annotations.
	InjectedAnnotations map[string]string `json:"injectedAnnotations"`
}

func validateCIDRList(cidrs string) error {
	if len(cidrs) > 0 {
		for _, cidr := range strings.Split(cidrs, ",") {
			if _, _, err := net.ParseCIDR(cidr); err != nil {
				return fmt.Errorf("failed parsing cidr '%s': %v", cidr, err)
			}
		}
	}
	return nil
}

func splitPorts(portsString string) []string {
	return strings.Split(portsString, ",")
}

func parsePort(portStr string) (int, error) {
	port, err := strconv.ParseUint(strings.TrimSpace(portStr), 10, 16)
	if err != nil {
		return 0, fmt.Errorf("failed parsing port '%d': %v", port, err)
	}
	return int(port), nil
}

func parsePorts(portsString string) ([]int, error) {
	portsString = strings.TrimSpace(portsString)
	ports := make([]int, 0)
	if len(portsString) > 0 {
		for _, portStr := range splitPorts(portsString) {
			port, err := parsePort(portStr)
			if err != nil {
				return nil, fmt.Errorf("failed parsing port '%d': %v", port, err)
			}
			ports = append(ports, port)
		}
	}
	return ports, nil
}

func validatePortList(parameterName, ports string) error {
	if _, err := parsePorts(ports); err != nil {
		return fmt.Errorf("%s invalid: %v", parameterName, err)
	}
	return nil
}

// ValidateIncludeIPRanges validates the includeIPRanges parameter
func ValidateIncludeIPRanges(ipRanges string) error {
	if ipRanges != "*" {
		if e := validateCIDRList(ipRanges); e != nil {
			return fmt.Errorf("includeIPRanges invalid: %v", e)
		}
	}
	return nil
}

// ValidateExcludeIPRanges validates the excludeIPRanges parameter
func ValidateExcludeIPRanges(ipRanges string) error {
	if e := validateCIDRList(ipRanges); e != nil {
		return fmt.Errorf("excludeIPRanges invalid: %v", e)
	}
	return nil
}

// ValidateIncludeInboundPorts validates the includeInboundPorts parameter
func ValidateIncludeInboundPorts(ports string) error {
	if ports != "*" {
		return validatePortList("includeInboundPorts", ports)
	}
	return nil
}

// ValidateExcludeInboundPorts validates the excludeInboundPorts parameter
func ValidateExcludeInboundPorts(ports string) error {
	return validatePortList("excludeInboundPorts", ports)
}

// ValidateExcludeOutboundPorts validates the excludeOutboundPorts parameter
func ValidateExcludeOutboundPorts(ports string) error {
	return validatePortList("excludeOutboundPorts", ports)
}

// validateStatusPort validates the statusPort parameter
func validateStatusPort(port string) error {
	if _, e := parsePort(port); e != nil {
		return fmt.Errorf("excludeInboundPorts invalid: %v", e)
	}
	return nil
}

// validateUInt32 validates that the given annotation value is a positive integer.
func validateUInt32(value string) error {
	_, err := strconv.ParseUint(value, 10, 32)
	return err
}

// validateBool validates that the given annotation value is a boolean.
func validateBool(value string) error {
	_, err := strconv.ParseBool(value)
	return err
}

func (wh *Webhook) getSidecarMode(namespace string, pod *corev1.Pod) utils.SidecarMode {
	// 这里主要是处理北极星 sidecar, 优先级: pod.annotations > namespace.labels > configmap
	sidecarMode := ""
	// 1. pod.annotations
	if val, ok := pod.Annotations[utils.PolarisSidecarMode]; ok {
		sidecarMode = val
		return utils.ParseSidecarMode(sidecarMode)
	}

	// 2. namespace.labels
	ns, err := wh.k8sClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
	if err != nil {
		// 如果出现异常，则就采用默认dns的注入方式
		log.Errorf("get pod namespace %q failed: %v", namespace, err)
	} else {
		if val, ok := ns.Labels[utils.PolarisSidecarModeLabel]; ok {
			sidecarMode = val
		}
		return utils.ParseSidecarMode(sidecarMode)
	}

	return wh.defaultSidecarMode

}

func (wh *Webhook) injectRequired(ignored []string, config *Config, podSpec *corev1.PodSpec, metadata *metav1.ObjectMeta) bool { // nolint: lll
	// Skip injection when host networking is enabled. The problem is
	// that the iptable changes are assumed to be within the pod when,
	// in fact, they are changing the routing at the host level. This
	// often results in routing failures within a node which can
	// affect the network provider within the cluster causing
	// additional pod failures.
	if podSpec.HostNetwork {
		return false
	}

	// skip special kubernetes system namespaces
	for _, namespace := range ignored {
		if metadata.Namespace == namespace {
			return false
		}
	}

	annos := metadata.GetAnnotations()
	if annos == nil {
		annos = map[string]string{}
	}

	var useDefault bool
	var inject bool
	switch strings.ToLower(annos[annotation.SidecarInject.Name]) {
	// http://yaml.org/type/bool.html
	case "y", "yes", "true", "on":
		inject = true
	case "n", "no", "false", "off":
		inject = false
	case "":
		useDefault = true
	}

	// If an annotation is not explicitly given, check the LabelSelectors, starting with NeverInject
	if useDefault {
		for _, neverSelector := range config.NeverInjectSelector {
			selector, err := metav1.LabelSelectorAsSelector(&neverSelector)
			if err != nil {
				log.Warnf("Invalid selector for NeverInjectSelector: %v (%v)", neverSelector, err)
			} else if !selector.Empty() && selector.Matches(labels.Set(metadata.Labels)) {
				log.Infof("Explicitly disabling injection for pod %s/%s due to pod labels matching NeverInjectSelector config map entry.",
					metadata.Namespace, potentialPodName(metadata))
				inject = false
				useDefault = false
				break
			}
		}
	}

	// If there's no annotation nor a NeverInjectSelector, check the AlwaysInject one
	if useDefault {
		for _, alwaysSelector := range config.AlwaysInjectSelector {
			selector, err := metav1.LabelSelectorAsSelector(&alwaysSelector)
			if err != nil {
				log.Warnf("Invalid selector for AlwaysInjectSelector: %v (%v)", alwaysSelector, err)
			} else if !selector.Empty() && selector.Matches(labels.Set(metadata.Labels)) {
				log.Infof("Explicitly enabling injection for pod %s/%s due to pod labels matching AlwaysInjectSelector config map entry.",
					metadata.Namespace, potentialPodName(metadata))
				inject = true
				useDefault = false
				break
			}
		}
	}

	var required bool
	switch config.Policy {
	default: // InjectionPolicyOff
		log.Errorf("Illegal value for autoInject:%s, must be one of [%s,%s]. Auto injection disabled!",
			config.Policy, InjectionPolicyDisabled, InjectionPolicyEnabled)
		required = false
	case InjectionPolicyDisabled:
		if useDefault {
			required = false
		} else {
			required = inject
		}
	case InjectionPolicyEnabled:
		if useDefault {
			required = true
		} else {
			required = inject
		}
	}

	// Build a log message for the annotations.
	annotationStr := ""
	for name := range annotationRegistry {
		value, ok := annos[name]
		if !ok {
			value = "(unset)"
		}
		annotationStr += fmt.Sprintf("%s:%s ", name, value)
	}

	log.Infof("Sidecar injection policy for %v/%v: namespacePolicy:%v useDefault:%v inject:%v required:%v %s",
		metadata.Namespace,
		potentialPodName(metadata),
		config.Policy,
		useDefault,
		inject,
		required,
		annotationStr)

	return required
}

func formatDuration(in *types.Duration) string {
	dur, err := types.DurationFromProto(in)
	if err != nil {
		return "1s"
	}
	return dur.String()
}

func isset(m map[string]string, key string) bool {
	_, ok := m[key]
	return ok
}

func openTlsMode(m map[string]string, key string) bool {
	if len(m) == 0 {
		return false
	}

	v, ok := m[key]
	if !ok {
		return false
	}

	return v != util.MTLSModeNone
}

func directory(filepath string) string {
	dir, _ := path.Split(filepath)
	return dir
}

func flippedContains(needle, haystack string) bool {
	return strings.Contains(haystack, needle)
}

// InjectionData renders sidecarTemplate with valuesConfig.
func InjectionData(sidecarTemplate, valuesConfig, version string, typeMetadata *metav1.TypeMeta, deploymentMetadata *metav1.ObjectMeta, spec *corev1.PodSpec,
	metadata *metav1.ObjectMeta, proxyConfig *meshconfig.ProxyConfig, meshConfig *meshconfig.MeshConfig) (
	*SidecarInjectionSpec, string, error,
) {
	// If DNSPolicy is not ClusterFirst, the Envoy sidecar may not able to connect to polaris.
	if spec.DNSPolicy != "" && spec.DNSPolicy != corev1.DNSClusterFirst {
		podName := potentialPodName(metadata)
		log.Warnf("[Webhook] %q's DNSPolicy is not %q. The Envoy sidecar may not able to connect to Istio Pilot",
			metadata.Namespace+"/"+podName, corev1.DNSClusterFirst)
	}

	if err := validateAnnotations(metadata.GetAnnotations()); err != nil {
		log.Errorf("[Webhook] injection failed due to invalid annotations: %v", err)
		return nil, "", err
	}

	values := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(valuesConfig), &values); err != nil {
		log.Infof("[Webhook] failed to parse values config: %v [%v]\n", err, valuesConfig)
		return nil, "", multierror.Prefix(err, "could not parse configuration values:")
	}

	tlsMode := util.MTLSModeNone
	if mode, ok := metadata.Annotations[util.PolarisTLSMode]; ok {
		mode = strings.ToLower(mode)
		if mode == util.MTLSModeStrict || mode == util.MTLSModePermissive {
			tlsMode = mode
		}
	}

	if len(metadata.Annotations) != 0 {
		metadata.Annotations[util.PolarisTLSMode] = tlsMode
	}

	data := SidecarTemplateData{
		TypeMeta:       typeMetadata,
		DeploymentMeta: deploymentMetadata,
		ObjectMeta:     metadata,
		Spec:           spec,
		ProxyConfig:    proxyConfig,
		MeshConfig:     meshConfig,
		Values:         values,
	}

	funcMap := template.FuncMap{
		"formatDuration":      formatDuration,
		"isset":               isset,
		"excludeInboundPort":  excludeInboundPort,
		"includeInboundPorts": includeInboundPorts,
		"kubevirtInterfaces":  kubevirtInterfaces,
		"applicationPorts":    applicationPorts,
		"annotation":          getAnnotation,
		"valueOrDefault":      valueOrDefault,
		"toJSON":              toJSON,
		"toJson":              toJSON, // Used by, e.g. Istio 1.0.5 template sidecar-injector-configmap.yaml
		"fromJSON":            fromJSON,
		"structToJSON":        structToJSON,
		"protoToJSON":         protoToJSON,
		"toYaml":              toYaml,
		"indent":              indent,
		"directory":           directory,
		"contains":            flippedContains,
		"toLower":             strings.ToLower,
		"openTlsMode":         openTlsMode,
	}

	// Allows the template to use env variables from istiod.
	// Istiod will use a custom template, without 'values.yaml', and the pod will have
	// an optional 'vendor' configmap where additional settings can be defined.
	funcMap["env"] = func(key string, def string) string {
		val := os.Getenv(key)
		if val == "" {
			return def
		}
		return val
	}

	// Need to use FuncMap and SidecarTemplateData context
	funcMap["render"] = func(template string) string {
		bbuf, err := parseTemplate(template, funcMap, data)
		if err != nil {
			return ""
		}

		return bbuf.String()
	}

	bbuf, err := parseTemplate(sidecarTemplate, funcMap, data)
	if err != nil {
		return nil, "", err
	}

	var sic SidecarInjectionSpec
	if err := yaml.Unmarshal(bbuf.Bytes(), &sic); err != nil {
		// This usually means an invalid injector template; we can't check
		// the template itself because it is merely a string.
		log.Warnf("Failed to unmarshal template %v %s", err, bbuf.String())
		return nil, "", multierror.Prefix(err, "failed parsing generated injected YAML (check sidecar injector configuration):")
	}

	// set sidecar --concurrency
	applyConcurrency(sic.Containers)

	status := &SidecarInjectionStatus{Version: version}
	for _, c := range sic.InitContainers {
		status.InitContainers = append(status.InitContainers, c.Name)
	}
	for _, c := range sic.Containers {
		status.Containers = append(status.Containers, c.Name)
	}
	for _, c := range sic.Volumes {
		status.Volumes = append(status.Volumes, c.Name)
	}
	for _, c := range sic.ImagePullSecrets {
		status.ImagePullSecrets = append(status.ImagePullSecrets, c.Name)
	}
	statusAnnotationValue, err := json.Marshal(status)
	if err != nil {
		return nil, "", fmt.Errorf("error encoded injection status: %v", err)
	}
	return &sic, string(statusAnnotationValue), nil
}

func parseTemplate(tmplStr string, funcMap map[string]interface{}, data SidecarTemplateData) (bytes.Buffer, error) {
	var tmpl bytes.Buffer
	temp := template.New("inject")
	t, err := temp.Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		log.Infof("Failed to parse template: %v %v\n", err, tmplStr)
		return bytes.Buffer{}, err
	}
	if err := t.Execute(&tmpl, &data); err != nil {
		log.Infof("Invalid template: %v %v\n", err, tmplStr)
		return bytes.Buffer{}, err
	}

	return tmpl, nil
}

// IntoResourceFile injects the istio proxy into the specified
// kubernetes YAML file.
func IntoResourceFile(sidecarTemplate string, valuesConfig string, meshconfig *meshconfig.MeshConfig, in io.Reader, out io.Writer) error {
	reader := yamlDecoder.NewYAMLReader(bufio.NewReaderSize(in, 4096))
	for {
		raw, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		obj, err := FromRawToObject(raw)
		if err != nil && !runtime.IsNotRegisteredError(err) {
			return err
		}

		var updated []byte
		if err == nil {
			outObject, err := IntoObject(sidecarTemplate, valuesConfig, meshconfig, obj) // nolint: vetshadow
			if err != nil {
				return err
			}
			if updated, err = yaml.Marshal(outObject); err != nil {
				return err
			}
		} else {
			updated = raw // unchanged
		}

		if _, err = out.Write(updated); err != nil {
			return err
		}
		if _, err = fmt.Fprint(out, "---\n"); err != nil {
			return err
		}
	}
	return nil
}

// FromRawToObject is used to convert from raw to the runtime object
func FromRawToObject(raw []byte) (runtime.Object, error) {
	var typeMeta metav1.TypeMeta
	if err := yaml.Unmarshal(raw, &typeMeta); err != nil {
		return nil, err
	}

	gvk := schema.FromAPIVersionAndKind(typeMeta.APIVersion, typeMeta.Kind)
	obj, err := injectScheme.New(gvk)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(raw, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

// IntoObject convert the incoming resources into Injected resources
func IntoObject(sidecarTemplate string, valuesConfig string, meshconfig *meshconfig.MeshConfig, in runtime.Object) (interface{}, error) {
	out := in.DeepCopyObject()

	var deploymentMetadata *metav1.ObjectMeta
	var metadata *metav1.ObjectMeta
	var podSpec *corev1.PodSpec
	var typeMeta *metav1.TypeMeta

	// Handle Lists
	if list, ok := out.(*corev1.List); ok {
		result := list

		for i, item := range list.Items {
			obj, err := FromRawToObject(item.Raw)
			if runtime.IsNotRegisteredError(err) {
				continue
			}
			if err != nil {
				return nil, err
			}

			r, err := IntoObject(sidecarTemplate, valuesConfig, meshconfig, obj) // nolint: vetshadow
			if err != nil {
				return nil, err
			}

			re := runtime.RawExtension{}
			re.Object = r.(runtime.Object)
			result.Items[i] = re
		}
		return result, nil
	}

	// CronJobs have JobTemplates in them, instead of Templates, so we
	// special case them.
	switch v := out.(type) {
	case *v2alpha1.CronJob:
		job := v
		typeMeta = &job.TypeMeta
		metadata = &job.Spec.JobTemplate.ObjectMeta
		deploymentMetadata = &job.ObjectMeta
		podSpec = &job.Spec.JobTemplate.Spec.Template.Spec
	case *corev1.Pod:
		pod := v
		typeMeta = &pod.TypeMeta
		metadata = &pod.ObjectMeta
		deploymentMetadata = &pod.ObjectMeta
		podSpec = &pod.Spec
	case *appsv1.Deployment: // Added to be explicit about the most expected case
		deploy := v
		typeMeta = &deploy.TypeMeta
		deploymentMetadata = &deploy.ObjectMeta
		metadata = &deploy.Spec.Template.ObjectMeta
		podSpec = &deploy.Spec.Template.Spec
	default:
		// `in` is a pointer to an Object. Dereference it.
		outValue := reflect.ValueOf(out).Elem()

		typeMeta = outValue.FieldByName("TypeMeta").Addr().Interface().(*metav1.TypeMeta)

		deploymentMetadata = outValue.FieldByName("ObjectMeta").Addr().Interface().(*metav1.ObjectMeta)

		templateValue := outValue.FieldByName("Spec").FieldByName("Template")
		// `Template` is defined as a pointer in some older API
		// definitions, e.g. ReplicationController
		if templateValue.Kind() == reflect.Ptr {
			if templateValue.IsNil() {
				return out, fmt.Errorf("spec.template is required value")
			}
			templateValue = templateValue.Elem()
		}
		metadata = templateValue.FieldByName("ObjectMeta").Addr().Interface().(*metav1.ObjectMeta)
		podSpec = templateValue.FieldByName("Spec").Addr().Interface().(*corev1.PodSpec)
	}

	name := metadata.Name
	if name == "" {
		name = deploymentMetadata.Name
	}
	// Skip injection when host networking is enabled. The problem is
	// that the iptable changes are assumed to be within the pod when,
	// in fact, they are changing the routing at the host level. This
	// often results in routing failures within a node which can
	// affect the network provider within the cluster causing
	// additional pod failures.
	if podSpec.HostNetwork {
		_, _ = fmt.Fprintf(os.Stderr, "Skipping injection because %q has host networking enabled\n",
			name)
		return out, nil
	}

	// skip injection for injected pods
	if len(podSpec.Containers) > 1 {
		for _, c := range podSpec.Containers {
			if c.Name == ProxyContainerName {
				_, _ = fmt.Fprintf(os.Stderr, "Skipping injection because %q has injected %q sidecar already\n",
					name, ProxyContainerName)
				return out, nil
			}
		}
	}

	spec, status, err := InjectionData(
		sidecarTemplate,
		valuesConfig,
		sidecarTemplateVersionHash(sidecarTemplate),
		typeMeta,
		deploymentMetadata,
		podSpec,
		metadata,
		meshconfig.DefaultConfig,
		meshconfig)
	if err != nil {
		return nil, err
	}

	podSpec.InitContainers = append(podSpec.InitContainers, spec.InitContainers...)

	podSpec.Containers = append(podSpec.Containers, spec.Containers...)
	podSpec.Volumes = append(podSpec.Volumes, spec.Volumes...)

	podSpec.DNSConfig = spec.DNSConfig

	// due to bug https://github.com/kubernetes/kubernetes/issues/57923,
	// k8s sa jwt token volume mount file is only accessible to root user, not istio-proxy(the user that istio proxy runs as).
	// workaround by https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-pod
	grp := int64(1337)
	if podSpec.SecurityContext == nil {
		podSpec.SecurityContext = &corev1.PodSecurityContext{
			FSGroup: &grp,
		}
	} else {
		podSpec.SecurityContext.FSGroup = &grp
	}

	if metadata.Annotations == nil {
		metadata.Annotations = make(map[string]string)
	}

	if len(spec.PodRedirectAnnot) != 0 {
		rewriteCniPodSpec(metadata.Annotations, spec)
	}

	if len(status) != 0 {
		metadata.Annotations[annotation.SidecarStatus.Name] = status
	}
	return out, nil
}

func getPortsForContainer(container corev1.Container) []string {
	parts := make([]string, 0)
	for _, p := range container.Ports {
		if p.Protocol == corev1.ProtocolUDP || p.Protocol == corev1.ProtocolSCTP {
			continue
		}
		parts = append(parts, strconv.Itoa(int(p.ContainerPort)))
	}
	return parts
}

func getContainerPorts(containers []corev1.Container, shouldIncludePorts func(corev1.Container) bool) string {
	parts := make([]string, 0)
	for _, c := range containers {
		if shouldIncludePorts(c) {
			parts = append(parts, getPortsForContainer(c)...)
		}
	}

	return strings.Join(parts, ",")
}

// this function is no longer used by the template but kept around for backwards compatibility
func applicationPorts(containers []corev1.Container) string {
	return getContainerPorts(containers, func(c corev1.Container) bool {
		return c.Name != ProxyContainerName
	})
}

func includeInboundPorts(containers []corev1.Container) string {
	// Include the ports from all containers in the deployment.
	return getContainerPorts(containers, func(corev1.Container) bool { return true })
}

func kubevirtInterfaces(s string) string {
	return s
}

func structToJSON(v interface{}) string {
	if v == nil {
		return "{}"
	}

	ba, err := json.Marshal(v)
	if err != nil {
		log.Warnf("Unable to marshal %v", v)
		return "{}"
	}

	return string(ba)
}

func protoToJSON(v proto.Message) string {
	if v == nil {
		return "{}"
	}

	m := jsonpb.Marshaler{}
	ba, err := m.MarshalToString(v)
	if err != nil {
		log.Warnf("Unable to marshal %v: %v", v, err)
		return "{}"
	}

	return ba
}

func toJSON(m map[string]string) string {
	if m == nil {
		return "{}"
	}

	ba, err := json.Marshal(m)
	if err != nil {
		log.Warnf("Unable to marshal %v", m)
		return "{}"
	}

	return string(ba)
}

func fromJSON(j string) interface{} {
	var m interface{}
	err := json.Unmarshal([]byte(j), &m)
	if err != nil {
		log.Warnf("Unable to unmarshal %s", j)
		return "{}"
	}

	log.Warnf("%v", m)
	return m
}

func indent(spaces int, source string) string {
	res := strings.Split(source, "\n")
	for i, line := range res {
		if i > 0 {
			res[i] = fmt.Sprintf(fmt.Sprintf("%% %ds%%s", spaces), "", line)
		}
	}
	return strings.Join(res, "\n")
}

func toYaml(value interface{}) string {
	y, err := yaml.Marshal(value)
	if err != nil {
		log.Warnf("Unable to marshal %v", value)
		return ""
	}

	return string(y)
}

func getAnnotation(meta metav1.ObjectMeta, name string, defaultValue interface{}) string {
	value, ok := meta.Annotations[name]
	if !ok {
		value = fmt.Sprint(defaultValue)
	}
	return value
}

func excludeInboundPort(port interface{}, excludedInboundPorts string) string {
	portStr := strings.TrimSpace(fmt.Sprint(port))
	if len(portStr) == 0 || portStr == "0" {
		// Nothing to do.
		return excludedInboundPorts
	}

	// Exclude the readiness port if not already excluded.
	ports := splitPorts(excludedInboundPorts)
	outPorts := make([]string, 0, len(ports))
	for _, port := range ports {
		if port == portStr {
			// The port is already excluded.
			return excludedInboundPorts
		}
		port = strings.TrimSpace(port)
		if len(port) > 0 {
			outPorts = append(outPorts, port)
		}
	}

	// The port was not already excluded - exclude it now.
	outPorts = append(outPorts, portStr)
	return strings.Join(outPorts, ",")
}

func valueOrDefault(value interface{}, defaultValue interface{}) interface{} {
	if value == "" || value == nil {
		return defaultValue
	}
	return value
}

// SidecarInjectionStatus contains basic information about the
// injected sidecar. This includes the names of added containers and
// volumes.
type SidecarInjectionStatus struct {
	Version          string   `json:"version"`
	InitContainers   []string `json:"initContainers"`
	Containers       []string `json:"containers"`
	Volumes          []string `json:"volumes"`
	ImagePullSecrets []string `json:"imagePullSecrets"`
}

// helper function to generate a template version identifier from a
// hash of the un-executed template contents.
func sidecarTemplateVersionHash(in string) string {
	hash := sha256.Sum256([]byte(in))
	return hex.EncodeToString(hash[:])
}

func potentialPodName(metadata *metav1.ObjectMeta) string {
	if metadata.Name != "" {
		return metadata.Name
	}
	if metadata.GenerateName != "" {
		return metadata.GenerateName + "***** (actual name not yet known)"
	}
	return ""
}

// rewriteCniPodSpec will check if values from the sidecar injector Helm
// values need to be inserted as Pod annotations so the CNI will apply
// the proper redirection rules.
func rewriteCniPodSpec(annotations map[string]string, spec *SidecarInjectionSpec) {
	if spec == nil {
		return
	}
	if len(spec.PodRedirectAnnot) == 0 {
		return
	}
	for k := range annotationRegistry {
		if spec.PodRedirectAnnot[k] != "" {
			if annotations[k] == spec.PodRedirectAnnot[k] {
				continue
			}
			annotations[k] = spec.PodRedirectAnnot[k]
		}
	}
}
