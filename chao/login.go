package chao

import (
	"encoding/json"
	"fmt"
	"net/url"
)

const (
	LoginUrl = "https://passport2-api.chaoxing.com/v11/loginregister"
)

type LoginResponse struct {
	Message string `json:"mes"`
	Type    int    `json:"type"`
	Url     string `json:"url"`
	Status  bool   `json:"status"`
}

func (c *ChaoxingClient) GenLoginData() url.Values {
	return url.Values{
		"uname": {c.UserID},
		"code":  {c.Password},
	}
}

func (c *ChaoxingClient) Login() error {
	// 发送登录请求，获取cookie存储在jar中
	resp, err := c.Client.PostForm(LoginUrl, c.GenLoginData())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 读取响应体LoginResponse
	loginResponse := &LoginResponse{}
	if err := json.NewDecoder(resp.Body).Decode(loginResponse); err != nil {
		return err
	}
	// 判断登录是否成功
	if !loginResponse.Status {
		return fmt.Errorf("登录失败，原因：%s", loginResponse.Message)
	}
	return nil
}
