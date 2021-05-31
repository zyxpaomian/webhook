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

// valiDateAdmission类型的拦截器，只做拦截，不做修改
func valiDateAdmission(res http.ResponseWriter, req *http.Request) {
	// 解析收到的报文
	reqContent, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		klog.Errorf("[webhook] 请求报文解析失败")
		common.ReqBodyInvalid(res)
		return
	}

	// 定义resp的报文
	var admissionResponse *admissionv1.AdmissionResponse
	requestedAdmissionReview := admissionv1.AdmissionReview{}
	// 对api-server传过来的报文做解析
	_, _, err = deserializer.Decode(reqContent, nil, &requestedAdmissionReview)
	if err != nil {
		klog.Errorf("[webhook] 请求报文解析失败: %v", err)
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

	klog.Infof("[webhook] 回调api-server, 发送报文: %v",responseAdmissionReview.Response)
	// send response
	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		klog.Errorf("[webhook] 无法解析回调报文: %v", err)
		http.Error(res, fmt.Sprintf("Can't write reponse: %v", err), http.StatusBadRequest)
		return
	}
	klog.Info("[webhook] 准备发送回包")

	if _, err := res.Write(respBytes); err != nil {
		klog.Errorf("[webhook] 无法发送回调报文: %v", err)
		http.Error(res, fmt.Sprintf("Can't write reponse: %v", err), http.StatusBadRequest)
	}
}


// mutatingadmission类型的拦截器，不仅做拦截，还需要修改
func mutateAdmission(res http.ResponseWriter, req *http.Request) {
	// 解析收到的报文
	reqContent, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		klog.Errorf("[webhook] 请求报文解析失败")
		common.ReqBodyInvalid(res)
		return
	}

	// 定义resp的报文
	var admissionResponse *admissionv1.AdmissionResponse
	requestedAdmissionReview := admissionv1.AdmissionReview{}
	// 对api-server传过来的报文做解析
	_, _, err = deserializer.Decode(reqContent, nil, &requestedAdmissionReview)
	if err != nil {
		klog.Errorf("[webhook] 请求报文解析失败: %v", err)
		admissionResponse = &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			},
		}
	// 解析成功，进行具体的拦截
	} else {
		admissionResponse = controller.Mutate(&requestedAdmissionReview)
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

	klog.Infof("[webhook] 回调api-server, 发送报文: %v",responseAdmissionReview.Response)
	// send response
	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		klog.Errorf("[webhook] 无法解析回调报文: %v", err)
		http.Error(res, fmt.Sprintf("Can't write reponse: %v", err), http.StatusBadRequest)
		return
	}
	klog.Info("[webhook] 准备发送回包")

	if _, err := res.Write(respBytes); err != nil {
		klog.Errorf("[webhook] 无法发送回调报文: %v", err)
		http.Error(res, fmt.Sprintf("Can't write reponse: %v", err), http.StatusBadRequest)
	}
}
