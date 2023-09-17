package chao

import (
	"fmt"
	"io"
	"net/url"
)

const (
	SIGN_SUCCESS = "success"
	SIGN_DONE    = "您已签到过了"

	PreSignUrl        = "https://mobilelearn.chaoxing.com/newsign/preSign?general=1&sys=1&ls=1&appType=15&tid=&ut=s"
	NormalSignUrl     = "https://mobilelearn.chaoxing.com/pptSign/stuSignajax"
	CodeAndGestureUrl = "https://mobilelearn.chaoxing.com/widget/sign/pcStuSignController/signIn"
	QRCodeSignUrl     = "https://mobilelearn.chaoxing.com/pptSign/stuSignajax"
)

// 关于传参的说明
// SignRequest 有两个来源
// 1.用户请求 -> SignDispatcher -> SignPool -> SignRequest
// 2.用户请求 -> SignDispatcher -> SignPool -> TrySign -> SignRequest

func (c *ChaoxingClient) GenPreSignData(req *SignRequest) url.Values {
	return url.Values{
		"courseId":        {fmt.Sprintf("%d", req.TargetCourse.CourseId)},
		"classId":         {fmt.Sprintf("%d", req.TargetCourse.ClassId)},
		"activePrimaryId": {fmt.Sprintf("%d", req.ActivityID)},
		"uid":             {fmt.Sprintf("%d", c.Uid)},
	}
}

func (c *ChaoxingClient) PreSign(req *SignRequest) error {
	resp, err := c.Client.PostForm(PreSignUrl, c.GenPreSignData(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *ChaoxingClient) GenNormalSignData(req *SignRequest) url.Values {
	return url.Values{
		// 固定参数
		"activeId": {fmt.Sprintf("%d", req.ActivityID)},
		"uid":      {fmt.Sprintf("%d", c.Uid)},
		"fid":      {fmt.Sprintf("%d", c.Fid)},
		"appType":  {"15"},
		// 位置签到参数
		"address":   {""},
		"ifTiJiao":  {"1"},
		"latitude":  {c.Latitude},
		"longitude": {c.Longitude},
		// 手势/签到码参数
		"signCode": {req.SignCode},
	}
}

// 移动端接口
// 0 通用签到接口 (普通签到，不提交图片的拍照签到，不提交位置的位置签到)
// 3,5 手势，签到码备用接口
func (c *ChaoxingClient) NormalSign(req *SignRequest) error {
	resp, err := c.Client.PostForm(NormalSignUrl, c.GenNormalSignData(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	text, _ := io.ReadAll(resp.Body)
	if !c.IsSignSuccess(string(text)) {
		// 试试手势签到/签到码签到
		return c.GestureAndCodeSign(req)
	}
	return nil
}

// 网页版接口
func (c *ChaoxingClient) GenGestureAndCodeSignData(req *SignRequest) url.Values {
	return url.Values{
		"activeId": {fmt.Sprintf("%d", req.ActivityID)},
		"classId":  {fmt.Sprintf("%d", req.TargetCourse.ClassId)},
		"courseId": {fmt.Sprintf("%d", req.TargetCourse.CourseId)},
	}
}

// 网页版接口
func (c *ChaoxingClient) GestureAndCodeSign(req *SignRequest) error {
	_, err := c.Client.PostForm(CodeAndGestureUrl, c.GenGestureAndCodeSignData(req))
	if err != nil {
		return err
	}
	return nil
}

func (c *ChaoxingClient) GenLocationSignData() url.Values {
	return url.Values{}
}

// 4. 位置签到
func (c *ChaoxingClient) LocationSign(req *SignRequest) error {
	return nil
}

func (c *ChaoxingClient) GenQRCodeSignData(req *SignRequest) url.Values {
	// url 是urlencoded的
	decodedString, err := url.QueryUnescape(req.QRCodeUrl)
	if err != nil {
		c.ReportError("解码出错:", err)
		return url.Values{}
	}
	// 解析 URL
	u, err := url.Parse(decodedString)
	if err != nil {
		c.ReportError("解析 URL 出错:", err)
		return url.Values{}
	}
	// 获取 enc 参数的值
	encValue := u.Query().Get("enc")
	activeId := u.Query().Get("c")
	return url.Values{
		"enc":      {encValue},
		"fid":      {"0"},
		"activeId": {activeId},
	}
}

// 2. 二维码签到
func (c *ChaoxingClient) QRCodeSign(req *SignRequest) error {
	resp, err := c.Client.PostForm(QRCodeSignUrl, c.GenQRCodeSignData(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	text, _ := io.ReadAll(resp.Body)
	if !c.IsSignSuccess(string(text)) {
		return fmt.Errorf("二维码签到失败，原因：%s", string(text))
	}
	return nil
}

// 无类型签到
func (c *ChaoxingClient) NoTypeSign(req *SignRequest) error {
	var err error
	if req.QRCodeUrl != "" {
		err = c.QRCodeSign(req)
	} else if req.SignCode != "" {
		err = c.NormalSign(req)
	} else {
		return fmt.Errorf("无法识别的签到类型")
	}
	return err
}

// 签到是否成功
func (c *ChaoxingClient) IsSignSuccess(text string) bool {
	return text == SIGN_SUCCESS || text == SIGN_DONE
}
