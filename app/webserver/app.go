package webserver

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	gs "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"github.com/xiaoxuxiansheng/xtimer/common/conf"
)

type Server struct {
	sync.Once
	engine *gin.Engine

	timerApp *TimerApp
	taskApp  *TaskApp

	timerRouter *gin.RouterGroup
	taskRouter  *gin.RouterGroup
	mockRouter  *gin.RouterGroup

	confProvider *conf.WebServerAppConfProvider
}

func NewServer(timer *TimerApp, task *TaskApp, confProvider *conf.WebServerAppConfProvider) *Server {
	s := Server{
		engine:       gin.Default(),
		timerApp:     timer,
		taskApp:      task,
		confProvider: confProvider,
	}

	s.engine.Use(CrosHandler())

	s.timerRouter = s.engine.Group("api/timer/v1")
	s.taskRouter = s.engine.Group("api/task/v1")
	s.mockRouter = s.engine.Group("api/mock/v1")
	s.RegisterBaseRouter()
	s.RegisterMockRouter()
	s.RegisterTimerRouter()
	s.RegisterTaskRouter()
	s.RegisterMonitorRouter()
	return &s
}

func (s *Server) Start() {
	s.Do(s.start)
}

func (s *Server) start() {
	conf := s.confProvider.Get()
	go func() {
		if err := s.engine.Run(fmt.Sprintf(":%d", conf.Port)); err != nil {
			panic(err)
		}
	}()
}

func (s *Server) RegisterBaseRouter() {
	s.engine.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))
}

func (s *Server) RegisterTimerRouter() {
	s.timerRouter.GET("/def", s.timerApp.GetTimer)
	s.timerRouter.POST("/def", s.timerApp.CreateTimer)
	s.timerRouter.DELETE("/def", s.timerApp.DeleteTimer)
	s.timerRouter.PATCH("/def", s.timerApp.UpdateTimer)

	s.timerRouter.GET("/defs", s.timerApp.GetAppTimers)
	s.timerRouter.GET("/defsByName", s.timerApp.GetTimersByName)

	s.timerRouter.POST("/enable", s.timerApp.EnableTimer)
	s.timerRouter.POST("/unable", s.timerApp.UnableTimer)
}

func (s *Server) RegisterTaskRouter() {
	s.taskRouter.GET("/records", s.taskApp.GetTasks)
}

func (s *Server) RegisterMockRouter() {
	s.mockRouter.Any("/mock", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, struct {
			Word string `json:"word"`
		}{
			Word: "hello world!",
		})
	})
}

func (s *Server) RegisterMonitorRouter() {
	s.engine.Any("/metrics", func(ctx *gin.Context) {
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
	})
}
