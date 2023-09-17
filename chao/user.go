package chao

import (
	"encoding/json"
	"fmt"
)

const (
	UserInfoUrl = "https://sso.chaoxing.com/apis/login/userLogin4Uname.do"
)

type UserInfoMsg struct {
	Fid  int    `json:"fid"`
	Uid  int    `json:"uid"`
	Name string `json:"name"`
}

type UserInfoResponse struct {
	Msg      UserInfoMsg `json:"msg"`
	Result   int         `json:"result"`
	ErrorMsg string      `json:"errorMsg,omitempty"`
}

func (c *ChaoxingClient) GetUserInfo() error {
	resp, err := c.Client.Get(UserInfoUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取响应体UserInfoResponse
	userInfoResponse := &UserInfoResponse{}
	if err := json.NewDecoder(resp.Body).Decode(userInfoResponse); err != nil {
		return err
	}
	if userInfoResponse.Result != 1 {
		return fmt.Errorf("获取用户信息失败，原因：%s", userInfoResponse.ErrorMsg)
	}
	// 读取参数保存到Client里
	c.Name = userInfoResponse.Msg.Name
	c.Fid = userInfoResponse.Msg.Fid
	c.Uid = userInfoResponse.Msg.Uid
	return nil
}
