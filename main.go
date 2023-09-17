package main

import (
	"context"
	"fmt"
	"signfunc/chao"
	"signfunc/libs"

	"github.com/tencentyun/scf-go-lib/cloudfunction"
)

type SignEvent struct {
	UserInfo     []libs.User `json:"userinfo"`               // 用户信息
	ActivityType string      `json:"activitytype,omitempty"` // 签到类型
	ActivityID   int         `json:"activityid,omitempty"`   // 签到活动ID
	CourseID     int         `json:"courseid,omitempty"`     // 课程ID
	ClassID      int         `json:"classid,omitempty"`      // 班级ID
	QRCodeURL    string      `json:"qrcodeurl,omitempty"`    // 二维码签到
	SignCode     string      `json:"signcode,omitempty"`     // 手势/签到码签到

	// UserInfo = 普通签到/位置签到
	// CoureseID + ClassID + ActivityID = （网页版）手势/签到码签到
	// ActivityID + SignCode = （移动端）手势/签到码签到
	// QRCodeURL = 二维码签到
}

// SignJob -> SignJobDispatcher -> SignJobPool -> SignRequest -> SignPool -> Sign
func SignJobDispatcher(ctx context.Context, event SignEvent) (string, error) {
	var (
		err error
	)
	// 检查参数
	if len(event.UserInfo) == 0 {
		fmt.Printf("%+v\n", event)
		return "null userInfo", nil
	}
	for _, userInfo := range event.UserInfo {
		// 生成签到请求
		signReq := &chao.SignRequest{
			ActivityType: event.ActivityType,
			ActivityID:   event.ActivityID,
			QRCodeUrl:    event.QRCodeURL,
			SignCode:     event.SignCode,
		}
		// 生成签到任务
		signJob := &chao.Job{
			UserInfo: &userInfo,
			SignReq:  signReq,
		}
		// 将签到任务提交到协程池
		err = chao.SignJobDispatcherInst.SumitJob(signJob)
		if err != nil {
			return "error", err
		}
	}
	chao.MainWait.Wait()
	return "success", nil
}

func main() {
	// Make the handler available for Remote Procedure Call by Cloud Function
	cloudfunction.Start(SignJobDispatcher)
}
