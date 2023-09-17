package chao

import (
	"encoding/json"
	"fmt"
)

// ActivityType 活动类型
// 0. 普通签到
// 2. 二维码签到
// 3. 手势签到
// 4. 位置签到
// 5. 签到码
const (
	ActivityTypeNormalSign = "0"
	ActivityTypeQRCodeSign = "2"
	ActivityTypeGesture    = "3"
	ActivityTypeLocation   = "4"
	ActivityTypeSignCode   = "5"

	ActivityUrl = "https://mobilelearn.chaoxing.com/v2/apis/active/student/activelist?&showNotStartedActive=0&_=1663752482576"
)

type ActivityResponse struct {
	Result   int          `json:"result"`
	Msg      string       `json:"msg"`
	Data     ActivityData `json:"data"`
	ErrorMsg string       `json:"errorMsg"`
}

type ActivityData struct {
	ReadingDuration int                      `json:"readingDuration"`
	ActiveList      []ActivityDataActiveList `json:"activeList"`
}

type ActivityDataActiveList struct {
	UserStatus int    `json:"userStatus"`
	NameTwo    string `json:"nameTwo"`
	OtherId    string `json:"otherId"`
	GroupId    int    `json:"groupId"`
	Source     int    `json:"source"`
	IsLook     int    `json:"isLook"`
	Type       int    `json:"type"`
	ReleaseNum int    `json:"releaseNum"`
	AttendNum  int    `json:"attendNum"`
	ActiveType int    `json:"activeType"`
	Logo       string `json:"logo"`
	NameOne    string `json:"nameOne"` // 判断签到类型
	StartTime  int    `json:"startTime"`
	Id         int    `json:"id"` // 签到用的Activity ID
	//EndTime    int    `json:"endTime"`
	Status   int    `json:"status"`
	NameFour string `json:"nameFour"`
}

//activeType=2 且 status=1 就为未签到活动

func (c *ChaoxingClient) TrySign(req *SignRequest) {
	url := fmt.Sprintf("%s&fid=%d&courseId=%d&classId=%d",
		ActivityUrl, c.Fid, req.TargetCourse.CourseId, req.TargetCourse.ClassId)
	resp, err := c.Client.Get(url)
	if err != nil {
		c.ReportError("获取活动信息失败", err)
		return
	}
	defer resp.Body.Close()
	activityResponse := &ActivityResponse{}
	if err := json.NewDecoder(resp.Body).Decode(activityResponse); err != nil {
		c.ReportError(c.UserID+"获取活动信息失败", err)
		return
	}
	if activityResponse.Result != 1 {
		fmt.Printf(c.UserID+"获取活动信息失败，原因：%s", activityResponse.ErrorMsg)
		return
	}
	for _, activity := range activityResponse.Data.ActiveList {
		if activity.ActiveType == 2 && activity.Status == 1 {
			req.ActivityID = activity.Id
			// 阻塞 预签到
			err = c.PreSign(req)
			if err != nil {
				c.ReportError("预签到失败", err)
				return
			}
			// 非阻塞 普通签到
			req.ActivityID = activity.Id
			req.ActivityType = ActivityTypeNormalSign
			SignProcessorInstance.SumitReq(req)
		}
	}
}
