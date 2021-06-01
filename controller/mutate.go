package controller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ghodss/yaml"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// 定义patch对象
type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// 需要插入的sidecar 容器
type sideConfig struct {
	Containers []corev1.Container `yaml:"containers"`
}

func loadSideCarConfig() (*sideConfig, error) {
	data, err := ioutil.ReadFile("/config/sidecar.yaml")
	if err != nil {
		return nil, err
	}

	var cfg sideConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
func deploymentPatch(deployment *appsv1.Deployment, sidecarConfig *sideConfig) ([]byte, error) {
	var patch []patchOperation
	patch = append(patch, addContainer(deployment.Spec.Template.Spec.Containers, sidecarConfig.Containers, "/spec/template/spec/containers")...)
	return json.Marshal(patch)
}

func addContainer(target, added []corev1.Container, basePath string) (patch []patchOperation) {
	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		path := basePath
		if first {
			first = false
			value = []corev1.Container{add}
		} else {
			path = path + "/-"
		}
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}
	return patch
}

func Mutate(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	req := ar.Request
	var (
		allowed = true
	)

	//klog.Infof("[webhook] AdmissionReview for Kind=%s, Namespace=%s Name=%s UID=%s", req.Kind.Kind, req.Namespace, req.Name, req.UID)

	var dep appsv1.Deployment
	if err := json.Unmarshal(req.Object.Raw, &dep); err != nil {
		klog.Errorf("[webhook] 无法解析AdmissionReview object raw: %v", err)
		allowed = false
		return &admissionv1.AdmissionResponse{
			Allowed: allowed,
			Result: &metav1.Status{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			},
		}
	}

	// 处理真正的业务逻辑
	siderConfig, err := loadSideCarConfig()
	if err != nil {
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			},
		}
	}
	patchBytes, err := deploymentPatch(&dep, siderConfig)
	if err != nil {
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			},
		}
	}

	klog.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))

	// 返回具体的admissionresponse
	return &admissionv1.AdmissionResponse{
		Allowed: allowed,
		Patch:   patchBytes,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}
