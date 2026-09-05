package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/tigerowo/infinite-canvas/model"
	"github.com/tigerowo/infinite-canvas/service"
)

const modelListRequestBodyLimit = 64 << 10

func AIModels(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, modelListRequestBodyLimit)
	var request struct {
		Channel model.ModelChannel `json:"channel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		Fail(w, "渠道参数格式错误")
		return
	}

	channel := request.Channel
	channel.BaseURL = strings.TrimSpace(channel.BaseURL)
	channel.APIKey = strings.TrimSpace(channel.APIKey)
	if channel.BaseURL == "" || channel.APIKey == "" {
		Fail(w, "请填写渠道的 Base URL 和 API Key")
		return
	}
	parsed, err := url.Parse(channel.BaseURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		Fail(w, "渠道地址格式错误")
		return
	}

	models, err := service.FetchModelChannelModels(channel)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, models)
}
