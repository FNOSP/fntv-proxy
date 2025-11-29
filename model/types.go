package model

type ProxyInfo struct {
	URL    string `json:"url"`
	Cookie string `json:"cookie"`
}

type CommonResponse struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ProxyInfoRequest struct {
	URL    string `json:"url"`
	Cookie string `json:"cookie"`
}
