package chao

import (
	"encoding/json"
	"fmt"
)

const (
	CourseInfoUrl = "http://mooc1-api.chaoxing.com/mycourse/backclazzdata?view=json&rss=1"
)

type CourseInfoResponse struct {
	Result      int           `json:"result"`
	Msg         string        `json:"msg"`
	ChannelList []ChannelList `json:"channelList"`
}

type ChannelList struct {
	Cfid     int    `json:"cfid"`
	Norder   int    `json:"norder"`
	CataName string `json:"cataName"`
	Cataid   string `json:"cataid"`
	Id       int    `json:"id"`
	Cpi      int    `json:"cpi"`
	//Key      int                `json:"key"`
	Content ChannelListContent `json:"content"`
	Topsign int                `json:"topsign"`
}

type ChannelListContent struct {
	Studentcount int                      `json:"studentcount"`
	Chatid       string                   `json:"chatid"`
	IsFiled      int                      `json:"isFiled"`
	Isthirdaq    int                      `json:"isthirdaq"`
	Isstart      bool                     `json:"isstart"`
	Isretire     int                      `json:"isretire"`
	Name         string                   `json:"name"`   // 班级名称
	Course       ChannelListContentCourse `json:"course"` // 课程信息
	Roletype     int                      `json:"roletype"`
	Id           int                      `json:"id"` // 签到用到的班级ID
	State        int                      `json:"state"`
	Cpi          int                      `json:"cpi"`
	Bbsid        string                   `json:"bbsid"`
	IsSquare     int                      `json:"isSquare"`
}

// 一般情况下，一个Channel只有一个课程
type ChannelListContentCourse struct {
	Data []ChannelListContentCourseData `json:"data"`
}

type ChannelListContentCourseData struct {
	BelongSchoolId     string `json:"belongSchoolId"`
	Coursestate        int    `json:"coursestate"`
	Teacherfactor      string `json:"teacherfactor"`
	IsCourseSquare     int    `json:"isCourseSquare"`
	CourseSquareUrl    string `json:"courseSquareUrl"`
	Imageurl           string `json:"imageurl"`
	Name               string `json:"name"` // 课程名称
	DefaultShowCatalog int    `json:"defaultShowCatalog"`
	Id                 int    `json:"id"` // 签到用到的课程ID
	AppData            int    `json:"appData"`
}

// 获取课程ID和班级ID列表

type Course struct {
	CourseId   int
	CourseName string
	ClassId    int
	ClassName  string
	ActivityID chan int // 待签到活动ID
}

type TrySignRequest struct {
	TargetCourse *Course
	ActivityID   int
}

// 获取课程信息并尝试 预签到 + 签到
// 预签到全部完成后 才可结束该函数
func (c *ChaoxingClient) GetCourseInfoAndTrySign() error {
	resp, err := c.Client.Get(CourseInfoUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取响应体CourseInfoResponse
	courseInfoResponse := &CourseInfoResponse{}
	if err := json.NewDecoder(resp.Body).Decode(courseInfoResponse); err != nil {
		return err
	}
	if courseInfoResponse.Result != 1 {
		return fmt.Errorf("获取课程信息失败，原因：%s", courseInfoResponse.Msg)
	}

	// 读取参数保存到Client里
	for _, channel := range courseInfoResponse.ChannelList {
		for _, course := range channel.Content.Course.Data {
			newCourse := &Course{
				CourseId:   course.Id,
				CourseName: course.Name,
				ClassId:    channel.Content.Id,
				ClassName:  channel.Content.Name,
			}
			newSignRequest := &SignRequest{
				ChaoxingClient: c,
				TargetCourse:   newCourse,
			}
			TrySignProcessorInst.SumitReq(newSignRequest)
		}
	}
	return nil
}
