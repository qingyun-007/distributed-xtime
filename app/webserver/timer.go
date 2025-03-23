package webserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/xiaoxuxiansheng/xtimer/common/model/vo"
	service "github.com/xiaoxuxiansheng/xtimer/service/webserver"

	"github.com/gin-gonic/gin"
)

type TimerApp struct {
	service timerService
}

func NewTimerApp(service *service.TimerService) *TimerApp {
	return &TimerApp{service: service}
}

// CreateTimer 创建定时器定义
// @Summary 创建定时器定义
// @Description 创建定时器定义
// @Tags 定时器接口
// @Accept application/json
// @Produce application/json
// @Param def body vo.Timer true "创建定时器定义"
// @Success 200 {object} vo.CreateTimerResp
// @Router /api/timer/v1/def [post]
func (t *TimerApp) CreateTimer(c *gin.Context) {
	var req vo.Timer
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[create timer] bind req failed, err: %v", err)))
		return
	}

	id, err := t.service.CreateTimer(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusOK, vo.NewCodeMsg(-1, err.Error()))
		return
	}
	c.JSON(http.StatusOK, vo.NewCreateTimerResp(id, vo.NewCodeMsgWithErr(nil)))
}

// GetAppTimers 获取 app 下的定时器
// @Summary 获取 app 下的定时器
// @Description 批量获取定时器定义
// @Tags 定时器接口
// @Accept application/json
// @Produce application/json
// @Param def body vo.GetAppTimersReq true "创建定时器定义"
// @Success 200 {object} vo.CreateTimerResp
// @Router /api/timer/v1/defs [post]
func (t *TimerApp) GetAppTimers(c *gin.Context) {
	var req vo.GetAppTimersReq
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[get app timers] bind req failed, err: %v", err)))
		return
	}

	timers, total, err := t.service.GetAppTimers(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusOK, vo.NewCodeMsg(-1, err.Error()))
		return
	}
	c.JSON(http.StatusOK, vo.NewGetTimersResp(timers, total, vo.NewCodeMsgWithErr(nil)))
}

func (t *TimerApp) GetTimersByName(c *gin.Context) {
	var req vo.GetTimersByNameReq
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[get timers by name] bind req failed, err: %v", err)))
		return
	}

	timers, total, err := t.service.GetTimersByName(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusOK, vo.NewCodeMsg(-1, err.Error()))
		return
	}
	c.JSON(http.StatusOK, vo.NewGetTimersResp(timers, total, vo.NewCodeMsgWithErr(nil)))
}

func (t *TimerApp) DeleteTimer(c *gin.Context) {
	var req vo.TimerReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[delete timer] bind req failed, err: %v", err)))
		return
	}

	if err := t.service.DeleteTimer(c.Request.Context(), req.App, req.ID); err != nil {
		c.JSON(http.StatusOK, vo.NewCodeMsg(-1, err.Error()))
		return
	}
	c.JSON(http.StatusOK, vo.NewCodeMsgWithErr(nil))
}

func (t *TimerApp) UpdateTimer(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}

func (t *TimerApp) GetTimer(c *gin.Context) {
	var req vo.TimerReq
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[get timer] bind req failed, err: %v", err)))
		return
	}

	timer, err := t.service.GetTimer(c.Request.Context(), req.ID)
	if err != nil {
		c.JSON(http.StatusOK, vo.NewCodeMsg(-1, err.Error()))
		return
	}
	c.JSON(http.StatusOK, vo.NewGetTimerResp(timer, vo.NewCodeMsgWithErr(nil)))
}

// EnableTimer 激活定时器
// @Summary 激活定时器
// @Description 激活定时器
// @Tags 定时器接口
// @Accept application/json
// @Produce application/json
// @Param def body vo.EnableTimerReq true "激活定时器请求"
// @Success 200 {object} vo.EnableTimerResp
// @Router /api/timer/v1/enable [post]
func (t *TimerApp) EnableTimer(c *gin.Context) {
	var req vo.TimerReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[enable timer] bind req failed, err: %v", err)))
		return
	}

	if err := t.service.EnableTimer(c.Request.Context(), req.App, req.ID); err != nil {
		c.JSON(http.StatusOK, vo.NewCodeMsg(-1, err.Error()))
		return
	}
	c.JSON(http.StatusOK, vo.NewCodeMsgWithErr(nil))
}

// UnableTimer 去激活定时器
// @Summary 去激活定时器
// @Description 去激活定时器
// @Tags 定时器接口
// @Accept application/json
// @Produce application/json
// @Param def body vo.UnableTimerReq true "去激活定时器请求"
// @Success 200 {object} vo.UnableTimerResp
// @Router /api/timer/v1/unable [post]
func (t *TimerApp) UnableTimer(c *gin.Context) {
	var req vo.TimerReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[enable timer] bind req failed, err:%v", err)))
		return
	}

	if err := t.service.UnableTimer(c.Request.Context(), req.App, req.ID); err != nil {
		c.JSON(http.StatusOK, vo.NewCodeMsg(-1, err.Error()))
		return
	}
	c.JSON(http.StatusOK, vo.NewCodeMsgWithErr(nil))
}

type timerService interface {
	CreateTimer(ctx context.Context, timer *vo.Timer) (uint, error)
	DeleteTimer(ctx context.Context, app string, id uint) error
	UpdateTimer(ctx context.Context, timer *vo.Timer) error
	GetTimer(ctx context.Context, id uint) (*vo.Timer, error)
	EnableTimer(ctx context.Context, app string, id uint) error
	UnableTimer(ctx context.Context, app string, id uint) error
	GetAppTimers(ctx context.Context, req *vo.GetAppTimersReq) ([]*vo.Timer, int64, error)
	GetTimersByName(ctx context.Context, req *vo.GetTimersByNameReq) ([]*vo.Timer, int64, error)
}
