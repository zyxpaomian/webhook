package handle

import (
	"encoding/json"
	"io/ioutil"
	"webhook/common"
	"webhook/controller"
	"k8s.io/klog"
	"net/http"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"fmt"
)


var (
	runtimeScheme = runtime.NewScheme()
	codeFactory   = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codeFactory.UniversalDeserializer()
)

// 只做拦截，不做修改
func valiDateAdmission(res http.ResponseWriter, req *http.Request) {
	reqContent, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		klog.Errorf("[http] 请求报文解析失败")
		common.ReqBodyInvalid(res)
		return
	}

	var admissionResponse *admissionv1.AdmissionResponse
	requestedAdmissionReview := admissionv1.AdmissionReview{}
	// 对api-server传过来的报文做解析
	_, _, err = deserializer.Decode(reqContent, nil, &requestedAdmissionReview)
	if err != nil {
		klog.Errorf("Can't decode body: %v", err)
		admissionResponse = &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			},
		}
	// 解析成功，进行具体的拦截
	} else {
		admissionResponse = controller.Validate(&requestedAdmissionReview)
	}

	// 构造返回的 AdmissionReview 这个结构体
	responseAdmissionReview := admissionv1.AdmissionReview{}
	// admission/v1
	responseAdmissionReview.APIVersion = requestedAdmissionReview.APIVersion
	responseAdmissionReview.Kind = requestedAdmissionReview.Kind
	if admissionResponse != nil {
		responseAdmissionReview.Response = admissionResponse
		if requestedAdmissionReview.Request != nil { // 返回相同的 UID
			responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		}
	}

	klog.Info(fmt.Sprintf("sending response: %v", responseAdmissionReview.Response))
	// send response
	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		klog.Errorf("Can't encode response: %v", err)
		common.ResMsg(res, 500, err.Error())
		return
	}
	klog.Info("Ready to write response...")

	if _, err := res.Write(respBytes); err != nil {
		klog.Errorf("Can't write response: %v", err)
		http.Error(res, fmt.Sprintf("Can't write reponse: %v", err), http.StatusBadRequest)
		// return
		// http.Error(res, fmt.Sprintf("Can't write reponse: %v", err), http.StatusBadRequest)
	}
}
