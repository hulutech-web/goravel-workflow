package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/goravel/framework/facades"
	"io/ioutil"
	"net/http"
)

// 微信服务
type WechatService struct {
}

type PhoneData struct {
	Errcode   int    `json:"errcode"`
	Errmsg    string `json:"errmsg"`
	PhoneInfo struct {
		PhoneNumber     string `json:"phoneNumber"`
		PurePhoneNumber string `json:"purePhoneNumber"`
		CountryCode     int    `json:"countryCode"`
		Watermark       struct {
			Timestamp int    `json:"timestamp"`
			Appid     string `json:"appid"`
		} `json:"watermark"`
	} `json:"phone_info"`
}

type AccessData struct {
	Accesstoken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// GetOpenidByCode 获取openid
func (t *WechatService) GetOpenidByCode(code string) (string, string, error) {
	appid := facades.Config().GetString("mini.app_id")
	secret := facades.Config().GetString("mini.app_secret")
	url := "https://api.weixin.qq.com/sns/jscode2session?appid=" + appid + "&secret=" + secret + "&js_code=" + code + "&grant_type=authorization_code"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	response, _ := client.Do(req)
	defer response.Body.Close()
	res, _ := ioutil.ReadAll(response.Body)

	type ResData struct {
		SessionKey string `json:"session_key"`
		OpenId     string `json:"openid"`
		Errcode    int    `json:"errcode"`
		Errmsg     string `json:"errmsg"`
		Unionid    string `json:"unionid"`
	}

	var result ResData
	if err1 := json.Unmarshal(res, &result); err1 != nil {
		return "", "", err1
	}

	if result.Errcode != 0 {
		return "", "", errors.New(result.Errmsg)
	}
	return result.OpenId, result.Unionid, nil
}

// GetPhoneNumberByCode 获取手机号
func (t *WechatService) GetPhoneNumberByCode(code string) (string, error) {
	url := "https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=" + t.CacheAccess()
	client := &http.Client{}
	var data = map[string]string{}
	data["code"] = code
	dataStr, _ := json.Marshal(data)
	dataBuf := bytes.NewBuffer(dataStr)
	req, _ := http.NewRequest("POST", url, dataBuf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	response, err := client.Do(req)
	defer response.Body.Close()

	var phoneData PhoneData
	body, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(body, &phoneData)
	if phoneData.Errcode != 0 {
		return "", err
	}
	return phoneData.PhoneInfo.PhoneNumber, nil
}

// CacheAccess 缓存access_token
func (t *WechatService) CacheAccess() string {
	access_token := facades.Cache().Get("access_token", "")
	if access_token != "" {
		return access_token.(string)
	}
	appid := facades.Config().GetString("mini.app_id")
	secret := facades.Config().GetString("mini.app_secret")
	url := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + appid + "&secret=" + secret
	response, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	var accessData AccessData
	json.Unmarshal(body, &accessData)
	facades.Cache().Put("access_token", accessData.Accesstoken, 7200)
	return accessData.Accesstoken
}
