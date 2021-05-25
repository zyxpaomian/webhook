package handle

import (
	"webhook/http"
)

func InitHandle(r *http.WWWMux) {
	// api相关的接口
	initAPIMapping(r)
}

func initAPIMapping(r *http.WWWMux) {
	// 用户认证
	r.RegistURLMapping("/v1/api/validate", "POST", valiDateAdmission)	
}

