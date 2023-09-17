package chao

import (
	"signfunc/libs"
	"sync"

	"github.com/panjf2000/ants/v2"
)

var (
	MainWait              = &sync.WaitGroup{}
	PreSignWait           = &sync.WaitGroup{}
	SignJobDispatcherInst *SignJobDispatcher
	SignProcessorInstance *SignProcessor
	TrySignProcessorInst  *TrySignProcessor
)

type TrySignProcessor struct {
	PreSignPool *ants.PoolWithFunc
}

func NewTrySignProcessor(preSignPool *ants.PoolWithFunc) *TrySignProcessor {
	return &TrySignProcessor{
		PreSignPool: preSignPool,
	}
}

func (s *TrySignProcessor) SumitReq(i interface{}) error {
	PreSignWait.Add(1)
	err := s.PreSignPool.Invoke(i)
	if err != nil {
		return err
	}
	return nil
}

// 对签到任务池的封装
type SignProcessor struct {
	SignPool *ants.PoolWithFunc
}

func NewSignProcessor(signPool *ants.PoolWithFunc) *SignProcessor {
	return &SignProcessor{
		SignPool: signPool,
	}
}

func (s *SignProcessor) SumitReq(i interface{}) error {
	MainWait.Add(1)
	err := s.SignPool.Invoke(i)
	if err != nil {
		return err
	}
	return nil
}

type SignJobDispatcher struct {
	SignJobPool *ants.PoolWithFunc
}

func NewSignJobDispatcher(signJobPool *ants.PoolWithFunc) *SignJobDispatcher {
	return &SignJobDispatcher{
		SignJobPool: signJobPool,
	}
}

func (s *SignJobDispatcher) SumitJob(i interface{}) error {
	MainWait.Add(1)
	err := s.SignJobPool.Invoke(i)
	if err != nil {
		return err
	}
	return nil
}

type SignRequest struct {
	ChaoxingClient *ChaoxingClient
	TargetCourse   *Course
	ActivityType   string // 签到类型, 实际上是OtherID
	ActivityID     int    // 签到活动ID
	QRCodeUrl      string // 二维码签到地址
	SignCode       string // 手势 签到码
}

type Job struct {
	UserInfo *libs.User
	SignReq  *SignRequest
}

func init() {
	var err error
	var jobPool *ants.PoolWithFunc
	jobPool, err = ants.NewPoolWithFunc(10, func(i interface{}) {
		defer MainWait.Done()
		var err error
		job := i.(*Job)
		chaoxingClient := NewChaoxingClient(job.UserInfo)
		if chaoxingClient == nil {
			chaoxingClient.ReportError("用户信息为空", nil)
			return
		}
		err = chaoxingClient.Login()
		if err != nil {
			chaoxingClient.ReportError("登录失败", err)
			return
		}
		err = chaoxingClient.GetUserInfo()
		if err != nil {
			chaoxingClient.ReportError("获取用户信息失败", err)
			return
		}
		// 预签到 + 普通签到 必须保证预签到后进行后续操作
		err = chaoxingClient.GetCourseInfoAndTrySign()
		if err != nil {
			chaoxingClient.ReportError("获取课程信息失败", err)
			return
		}
		// 等待预签到完成
		PreSignWait.Wait()
		// 开始签到
		job.SignReq.ChaoxingClient = chaoxingClient
		SignProcessorInstance.SumitReq(job.SignReq)

	}, ants.WithPreAlloc(true))
	if err != nil {
		panic(err)
	}

	// SignJob -> SignJobDispatcher -> SignJobPool -> PreSign => SignRequest => SignPool => Sign
	var signPool *ants.PoolWithFunc
	signPool, err = ants.NewPoolWithFunc(10, func(i interface{}) {
		defer MainWait.Done()
		// ActivityType 活动类型
		// -1. 预签到
		// 0. 普通签到
		// 2. 二维码签到
		// 3. 手势签到
		// 4. 位置签到
		// 5. 签到码

		req := i.(*SignRequest)
		chaoxingClient := req.ChaoxingClient
		switch req.ActivityType {
		case "-1":
			err = chaoxingClient.PreSign(req)
		case "0":
			err = chaoxingClient.NormalSign(req)
		case "2":
			err = chaoxingClient.QRCodeSign(req)
		case "3":
			err = chaoxingClient.NormalSign(req)
		case "4":
			err = chaoxingClient.NormalSign(req)
		case "5":
			err = chaoxingClient.NormalSign(req)
		default:
			err = chaoxingClient.NoTypeSign(req)
		}
		if err != nil {
			chaoxingClient.ReportError("签到失败", err)
		}
	}, ants.WithPreAlloc(true))
	if err != nil {
		panic(err)
	}

	var trySignPool *ants.PoolWithFunc
	trySignPool, err = ants.NewPoolWithFunc(10, func(i interface{}) {
		defer PreSignWait.Done()
		req := i.(*SignRequest)
		chaoxingClient := req.ChaoxingClient
		chaoxingClient.TrySign(req)
	}, ants.WithPreAlloc(true))
	if err != nil {
		panic(err)
	}

	SignJobDispatcherInst = NewSignJobDispatcher(jobPool)
	SignProcessorInstance = NewSignProcessor(signPool)
	TrySignProcessorInst = NewTrySignProcessor(trySignPool)

}
