package chao

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"signfunc/libs"
)

const (
	ChaoxingUrl = "https://chaoxing.com"
)

type ChaoxingClient struct {
	Client    *http.Client
	UserID    string
	Password  string
	Latitude  string
	Longitude string
	Name      string
	Uid       int
	Fid       int
}

func NewChaoxingClient(userInfo *libs.User) *ChaoxingClient {
	// 检查信息是否为空
	if userInfo.UserID == "" || userInfo.Password == "" {
		return nil
	}
	jar, _ := cookiejar.New(nil)
	return &ChaoxingClient{
		Client: &http.Client{
			Jar: jar,
		},
		UserID:    userInfo.UserID,
		Password:  userInfo.Password,
		Latitude:  userInfo.Latitude,
		Longitude: userInfo.Longitude,
	}
}

func (c *ChaoxingClient) ReportError(msg string, err error) {
	fmt.Println(c.UserID, msg, err)
}
