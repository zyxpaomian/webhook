package controller

import (
	"encoding/json"
	admissionv1 "k8s.io/api/admission/v1"
	//corev1 "k8s.io/api/core/v1"
	"net/http"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func Validate(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	req := ar.Request
	var (
		allowed = true
		code    = http.StatusOK
		message = ""
	)

	//klog.Infof("[webhook] AdmissionReview for Kind=%s, Namespace=%s Name=%s UID=%s", req.Kind.Kind, req.Namespace, req.Name, req.UID)

	var dep appsv1.Deployment
	if err := json.Unmarshal(req.Object.Raw, &dep); err != nil {
		klog.Errorf("[webhook] 无法解析AdmissionReview object raw: %v", err)
		allowed = false
		code = http.StatusBadRequest
		return &admissionv1.AdmissionResponse{
			Allowed: allowed,
			Result: &metav1.Status{
				Code:    int32(code),
				Message: err.Error(),
			},
		}
	}

	// 处理真正的业务逻辑
	replicas := *dep.Spec.Replicas
	if replicas < 3 {
		klog.Infof("[webhook] deployment不满足最低副本数量，无法创建")
		allowed = false
		code = http.StatusForbidden
		message = "need 3 replicas at least, create deployment failed"
	}

	// 返回具体的admissionresponse
	return &admissionv1.AdmissionResponse{
		Allowed: allowed,
		Result: &metav1.Status{
			Code:    int32(code),
			Message: message,
		},
	}
}
