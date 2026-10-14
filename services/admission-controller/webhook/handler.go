package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/snisid/platform/internal/platform/logger"
	"go.uber.org/zap"
)

type AdmissionReview struct {
	Request  *AdmissionRequest  `json:"request,omitempty"`
	Response *AdmissionResponse `json:"response,omitempty"`
}

type AdmissionRequest struct {
	UID       string               `json:"uid"`
	Kind      GroupVersionKind     `json:"kind"`
	Resource  GroupVersionResource `json:"resource"`
	Name      string               `json:"name,omitempty"`
	Namespace string               `json:"namespace,omitempty"`
	Operation string               `json:"operation"`
	Object    json.RawMessage      `json:"object,omitempty"`
	OldObject json.RawMessage      `json:"oldObject,omitempty"`
	UserInfo  UserInfo             `json:"userInfo"`
}

type GroupVersionKind struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

type GroupVersionResource struct {
	Group    string `json:"group"`
	Version  string `json:"version"`
	Resource string `json:"resource"`
}

type UserInfo struct {
	Username string              `json:"username"`
	UID      string              `json:"uid"`
	Groups   []string            `json:"groups"`
	Extra    map[string][]string `json:"extra"`
}

type AdmissionResponse struct {
	UID      string   `json:"uid"`
	Allowed  bool     `json:"allowed"`
	Result   *Result  `json:"status,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type Result struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
	Reason  string `json:"reason,omitempty"`
}

type PodSpec struct {
	Containers       []Container         `json:"containers"`
	InitContainers   []Container         `json:"initContainers,omitempty"`
	ServiceAccount   string              `json:"serviceAccountName,omitempty"`
	ImagePullSecrets []PullSecret        `json:"imagePullSecrets,omitempty"`
	NodeSelector     map[string]string   `json:"nodeSelector,omitempty"`
	SecurityContext  *PodSecurityContext `json:"securityContext,omitempty"`
}

type Container struct {
	Name            string                    `json:"name"`
	Image           string                    `json:"image"`
	Resources       ResourceRequirements      `json:"resources,omitempty"`
	SecurityContext *ContainerSecurityContext `json:"securityContext,omitempty"`
	Env             []EnvVar                  `json:"env,omitempty"`
	VolumeMounts    []VolumeMount             `json:"volumeMounts,omitempty"`
}

type PodSecurityContext struct {
	RunAsNonRoot   *bool           `json:"runAsNonRoot,omitempty"`
	SeccompProfile *SeccompProfile `json:"seccompProfile,omitempty"`
}

type SeccompProfile struct {
	Type string `json:"type"`
}

type ContainerSecurityContext struct {
	Privileged               *bool         `json:"privileged,omitempty"`
	AllowPrivilegeEscalation *bool         `json:"allowPrivilegeEscalation,omitempty"`
	ReadOnlyRootFilesystem   *bool         `json:"readOnlyRootFilesystem,omitempty"`
	RunAsNonRoot             *bool         `json:"runAsNonRoot,omitempty"`
	Capabilities             *Capabilities `json:"capabilities,omitempty"`
}

type Capabilities struct {
	Drop []string `json:"drop,omitempty"`
}

type ResourceRequirements struct {
	Limits   map[string]string `json:"limits,omitempty"`
	Requests map[string]string `json:"requests,omitempty"`
}

type PullSecret struct {
	Name string `json:"name"`
}

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
}

type ObjectMeta struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type Pod struct {
	APIVersion string     `json:"apiVersion"`
	Kind       string     `json:"kind"`
	Metadata   ObjectMeta `json:"metadata"`
	Spec       PodSpec    `json:"spec"`
}

type PolicyViolation struct {
	Field    string `json:"field"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // ERROR, WARNING
}

var requiredLabels = []string{
	"snisid.gouv.ht/component",
	"snisid.gouv.ht/environment",
}

var allowedRegistries = []string{
	"registry.snisid.gouv.ht",
	"docker.snisid.gouv.ht",
	"ghcr.io/snisid",
}

var forbiddenPaths = []string{
	"/host",
	"/var/run/docker",
	"/proc",
}

func HandleAdmission(w http.ResponseWriter, r *http.Request) {
	var review AdmissionReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, `{"response":{"allowed":false,"status":{"code":400,"message":"invalid request body"}}}`, http.StatusBadRequest)
		return
	}

	if review.Request == nil {
		http.Error(w, `{"response":{"allowed":false,"status":{"code":400,"message":"empty request"}}}`, http.StatusBadRequest)
		return
	}

	response := &AdmissionResponse{
		UID:     review.Request.UID,
		Allowed: true,
	}

	violations := validatePod(review.Request)
	if len(violations) > 0 {
		for _, v := range violations {
			if v.Severity == "ERROR" {
				response.Allowed = false
				response.Result = &Result{
					Code:    403,
					Message: buildDenyMessage(violations),
					Reason:  "PolicyViolation",
				}
				break
			}
		}
		for _, v := range violations {
			if v.Severity == "WARNING" {
				response.Warnings = append(response.Warnings, fmt.Sprintf("%s: %s", v.Field, v.Message))
			}
		}
	}

	if response.Allowed {
		logger.Info(context.Background(), "ADMISSION: pod allowed", zap.String("name", review.Request.Name), zap.String("namespace", review.Request.Namespace))
	} else {
		logger.Warn(context.Background(), "ADMISSION: pod rejected", zap.String("name", review.Request.Name), zap.String("namespace", review.Request.Namespace), zap.String("reason", response.Result.Message))
	}

	review.Response = response
	json.NewEncoder(w).Encode(review)
}

func validatePod(req *AdmissionRequest) []PolicyViolation {
	if req.Kind.Kind != "Pod" || req.Operation == "DELETE" {
		return nil
	}

	var pod Pod
	if err := json.Unmarshal(req.Object, &pod); err != nil {
		return []PolicyViolation{{Field: "object", Message: "cannot parse pod spec", Severity: "ERROR"}}
	}

	violations := []PolicyViolation{}

	violations = append(violations, validateLabels(pod.Metadata.Labels)...)
	violations = append(violations, validateAnnotations(pod.Metadata.Annotations, req.Namespace)...)
	violations = append(violations, validateContainers(pod.Spec.Containers)...)
	violations = append(violations, validateContainers(pod.Spec.InitContainers)...)
	violations = append(violations, validateImagePullSecrets(pod.Spec.ImagePullSecrets)...)
	violations = append(violations, validatePodSecurityContext(pod.Spec.SecurityContext)...)
	violations = append(violations, validateUserInfo(req.UserInfo)...)

	return violations
}

func validateLabels(labels map[string]string) []PolicyViolation {
	var v []PolicyViolation
	for _, required := range requiredLabels {
		if _, ok := labels[required]; !ok {
			v = append(v, PolicyViolation{
				Field:    fmt.Sprintf("metadata.labels['%s']", required),
				Message:  fmt.Sprintf("required label %s is missing", required),
				Severity: "ERROR",
			})
		}
	}
	return v
}

func validateAnnotations(annotations map[string]string, namespace string) []PolicyViolation {
	var v []PolicyViolation
	if annotations == nil {
		return v
	}
	if ns, ok := annotations["snisid.gouv.ht/namespace"]; ok && ns != namespace {
		v = append(v, PolicyViolation{
			Field:    "metadata.annotations['snisid.gouv.ht/namespace']",
			Message:  "annotation namespace does not match request namespace",
			Severity: "ERROR",
		})
	}
	if val, ok := annotations["snisid.gouv.ht/security-tier"]; ok {
		allowed := map[string]bool{"gold": true, "silver": true, "bronze": true}
		if !allowed[val] {
			v = append(v, PolicyViolation{
				Field:    "metadata.annotations['snisid.gouv.ht/security-tier']",
				Message:  "invalid security tier, must be gold/silver/bronze",
				Severity: "ERROR",
			})
		}
	}
	return v
}

func validateContainers(containers []Container) []PolicyViolation {
	var v []PolicyViolation
	for _, c := range containers {
		v = append(v, validateContainerImage(c)...)
		v = append(v, validateContainerResources(c)...)
		v = append(v, validateContainerSecurity(c)...)
		v = append(v, validateContainerVolumeMounts(c)...)
	}
	return v
}

func validateContainerImage(c Container) []PolicyViolation {
	var v []PolicyViolation
	if c.Image == "" {
		v = append(v, PolicyViolation{Field: fmt.Sprintf("spec.containers[%s].image", c.Name), Message: "container image must not be empty", Severity: "ERROR"})
		return v
	}
	if strings.HasSuffix(c.Image, ":latest") {
		v = append(v, PolicyViolation{Field: fmt.Sprintf("spec.containers[%s].image", c.Name), Message: "using :latest tag is forbidden", Severity: "ERROR"})
	}
	registry := extractRegistry(c.Image)
	if registry != "" {
		allowed := false
		for _, a := range allowedRegistries {
			if registry == a || strings.HasPrefix(c.Image, a+"/") {
				allowed = true
				break
			}
		}
		if !allowed {
			v = append(v, PolicyViolation{Field: fmt.Sprintf("spec.containers[%s].image", c.Name), Message: fmt.Sprintf("image registry %s is not allowed", registry), Severity: "WARNING"})
		}
	}
	return v
}

func validateContainerResources(c Container) []PolicyViolation {
	var v []PolicyViolation
	if c.Resources.Limits == nil {
		v = append(v, PolicyViolation{Field: fmt.Sprintf("spec.containers[%s].resources.limits", c.Name), Message: "resource limits must be specified", Severity: "WARNING"})
		return v
	}
	if _, ok := c.Resources.Limits["memory"]; !ok {
		v = append(v, PolicyViolation{Field: fmt.Sprintf("spec.containers[%s].resources.limits.memory", c.Name), Message: "memory limit must be specified", Severity: "WARNING"})
	}
	if _, ok := c.Resources.Limits["cpu"]; !ok {
		v = append(v, PolicyViolation{Field: fmt.Sprintf("spec.containers[%s].resources.limits.cpu", c.Name), Message: "cpu limit must be specified", Severity: "WARNING"})
	}
	return v
}

func validateContainerSecurity(c Container) []PolicyViolation {
	var v []PolicyViolation
	sc := c.SecurityContext
	if sc == nil {
		v = append(v, PolicyViolation{Field: fmt.Sprintf("spec.containers[%s].securityContext", c.Name), Message: "security context must be set", Severity: "WARNING"})
		return v
	}
	if sc.Privileged != nil && *sc.Privileged {
		v = append(v, PolicyViolation{Field: fmt.Sprintf("spec.containers[%s].securityContext.privileged", c.Name), Message: "privileged container is forbidden", Severity: "ERROR"})
	}
	if sc.AllowPrivilegeEscalation != nil && *sc.AllowPrivilegeEscalation {
		v = append(v, PolicyViolation{Field: fmt.Sprintf("spec.containers[%s].securityContext.allowPrivilegeEscalation", c.Name), Message: "privilege escalation is forbidden", Severity: "ERROR"})
	}
	if sc.ReadOnlyRootFilesystem != nil && !*sc.ReadOnlyRootFilesystem {
		v = append(v, PolicyViolation{Field: fmt.Sprintf("spec.containers[%s].securityContext.readOnlyRootFilesystem", c.Name), Message: "read-only root fs is recommended", Severity: "WARNING"})
	}
	if sc.Capabilities != nil {
		hasDropAll := false
		for _, d := range sc.Capabilities.Drop {
			if d == "ALL" {
				hasDropAll = true
				break
			}
		}
		if !hasDropAll {
			v = append(v, PolicyViolation{Field: fmt.Sprintf("spec.containers[%s].securityContext.capabilities", c.Name), Message: "containers should drop all capabilities", Severity: "WARNING"})
		}
	}
	return v
}

func validateContainerVolumeMounts(c Container) []PolicyViolation {
	var v []PolicyViolation
	for _, vm := range c.VolumeMounts {
		for _, fp := range forbiddenPaths {
			if strings.HasPrefix(vm.MountPath, fp) {
				v = append(v, PolicyViolation{Field: fmt.Sprintf("spec.containers[%s].volumeMounts[%s]", c.Name, vm.Name), Message: fmt.Sprintf("mount path %s is forbidden", vm.MountPath), Severity: "ERROR"})
			}
		}
	}
	return v
}

func validateImagePullSecrets(secrets []PullSecret) []PolicyViolation {
	var v []PolicyViolation
	if len(secrets) == 0 {
		v = append(v, PolicyViolation{Field: "spec.imagePullSecrets", Message: "imagePullSecrets should be configured for private registry", Severity: "WARNING"})
	}
	return v
}

func validatePodSecurityContext(sc *PodSecurityContext) []PolicyViolation {
	var v []PolicyViolation
	if sc == nil {
		v = append(v, PolicyViolation{Field: "spec.securityContext", Message: "pod security context must be set", Severity: "WARNING"})
		return v
	}
	if sc.RunAsNonRoot != nil && !*sc.RunAsNonRoot {
		v = append(v, PolicyViolation{Field: "spec.securityContext.runAsNonRoot", Message: "runAsNonRoot should be true", Severity: "WARNING"})
	}
	return v
}

func validateUserInfo(user UserInfo) []PolicyViolation {
	var v []PolicyViolation
	if user.Username == "" {
		v = append(v, PolicyViolation{Field: "userInfo.username", Message: "authenticated user must be present", Severity: "ERROR"})
	}
	hasSNISIDGroup := false
	for _, g := range user.Groups {
		if strings.HasPrefix(g, "snisid:") {
			hasSNISIDGroup = true
			break
		}
	}
	if !hasSNISIDGroup {
		v = append(v, PolicyViolation{Field: "userInfo.groups", Message: "user must belong to a snisid group", Severity: "WARNING"})
	}
	return v
}

func extractRegistry(image string) string {
	parts := strings.SplitN(image, "/", 2)
	if len(parts) < 2 || !strings.Contains(parts[0], ".") {
		return ""
	}
	return parts[0]
}

func buildDenyMessage(violations []PolicyViolation) string {
	var msgs []string
	for _, v := range violations {
		if v.Severity == "ERROR" {
			msgs = append(msgs, fmt.Sprintf("[%s] %s", v.Field, v.Message))
		}
	}
	return "ADMISSION BLOCKED: " + strings.Join(msgs, "; ")
}
