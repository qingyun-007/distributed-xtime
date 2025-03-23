package trigger

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xiaoxuxiansheng/xtimer/common/conf"
	"github.com/xiaoxuxiansheng/xtimer/common/model/vo"
	"github.com/xiaoxuxiansheng/xtimer/common/utils"
	"github.com/xiaoxuxiansheng/xtimer/pkg/concurrency"
	"github.com/xiaoxuxiansheng/xtimer/pkg/log"
	"github.com/xiaoxuxiansheng/xtimer/pkg/pool"
	"github.com/xiaoxuxiansheng/xtimer/pkg/redis"
	"github.com/xiaoxuxiansheng/xtimer/service/executor"
)

type Worker struct {
	task         taskService
	confProvider confProvider
	pool         pool.WorkerPool
	executor     *executor.Worker
	lockService  *redis.Client
}

func NewWorker(executor *executor.Worker, task *TaskService, lockService *redis.Client, confProvider *conf.TriggerAppConfProvider) *Worker {
	return &Worker{
		executor:     executor,
		task:         task,
		lockService:  lockService,
		pool:         pool.NewGoWorkerPool(confProvider.Get().WorkersNum),
		confProvider: confProvider,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.executor.Start(ctx)
}

func (w *Worker) Work(ctx context.Context, minuteBucketKey string, ack func()) error {
	// log.InfoContextf(ctx, "trigger_1 start: %v", time.Now())
	// defer func() {
	// 	log.InfoContextf(ctx, "trigger_1 end: %v", time.Now())
	// }()

	// 进行为时一分钟的 zrange 处理
	startTime, err := getStartMinute(minuteBucketKey)
	if err != nil {
		return err
	}

	conf := w.confProvider.Get()
	ticker := time.NewTicker(time.Duration(conf.ZRangeGapSeconds) * time.Second)
	defer ticker.Stop()

	endTime := startTime.Add(time.Minute)

	notifier := concurrency.NewSafeChan(int(time.Minute/(time.Duration(conf.ZRangeGapSeconds)*time.Second)) + 1)
	defer notifier.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// log.InfoContextf(ctx, "trigger_2 start: %v", time.Now())
		// defer func() {
		// 	log.InfoContextf(ctx, "trigger_2 end: %v", time.Now())
		// }()
		defer wg.Done()
		if err := w.handleBatch(ctx, minuteBucketKey, startTime, startTime.Add(time.Duration(conf.ZRangeGapSeconds)*time.Second)); err != nil {
			notifier.Put(err)
		}
	}()
	for range ticker.C {
		select {
		case e := <-notifier.GetChan():
			err, _ = e.(error)
			return err
		default:
		}

		if startTime = startTime.Add(time.Duration(conf.ZRangeGapSeconds) * time.Second); startTime.Equal(endTime) || startTime.After(endTime) {
			break
		}

		// log.InfoContextf(ctx, "start time: %v", startTime)

		wg.Add(1)
		go func(startTime time.Time) {
			// log.InfoContextf(ctx, "trigger_2 start: %v", time.Now())
			// defer func() {
			// 	log.InfoContextf(ctx, "trigger_2 end: %v", time.Now())
			// }()
			defer wg.Done()
			if err := w.handleBatch(ctx, minuteBucketKey, startTime, startTime.Add(time.Duration(conf.ZRangeGapSeconds)*time.Second)); err != nil {
				notifier.Put(err)
			}
		}(startTime)
	}

	wg.Wait()
	select {
	case e := <-notifier.GetChan():
		err, _ = e.(error)
		return err
	default:
	}

	ack()
	log.InfoContextf(ctx, "ack success, key: %s", minuteBucketKey)
	return nil
}

func (w *Worker) handleBatch(ctx context.Context, key string, start, end time.Time) error {
	bucket, err := getBucket(key)
	if err != nil {
		return err
	}

	tasks, err := w.task.GetTasksByTime(ctx, key, bucket, start, end)
	if err != nil {
		return err
	}

	timerIDs := make([]uint, 0, len(tasks))
	for _, task := range tasks {
		timerIDs = append(timerIDs, task.TimerID)
	}
	// log.InfoContextf(ctx, "key: %s, get tasks: %+v, start: %v, end: %v", key, timerIDs, start, end)
	for _, task := range tasks {
		task := task
		if err := w.pool.Submit(func() {
			// log.InfoContextf(ctx, "trigger_3 start: %v", time.Now())
			// defer func() {
			// 	log.InfoContextf(ctx, "trigger_3 end: %v", time.Now())
			// }()
			if err := w.executor.Work(ctx, utils.UnionTimerIDUnix(task.TimerID, task.RunTimer.UnixMilli())); err != nil {
				log.ErrorContextf(ctx, "executor work failed, err: %v", err)
			}
		}); err != nil {
			return err
		}
	}
	return nil
}

func getStartMinute(slice string) (time.Time, error) {
	timeBucket := strings.Split(slice, "_")
	if len(timeBucket) != 2 {
		return time.Time{}, fmt.Errorf("invalid format of msg key: %s", slice)
	}

	return utils.GetStartMinute(timeBucket[0])
}

func getBucket(slice string) (int, error) {
	timeBucket := strings.Split(slice, "_")
	if len(timeBucket) != 2 {
		return -1, fmt.Errorf("invalid format of msg key: %s", slice)
	}
	return strconv.Atoi(timeBucket[1])
}

type taskService interface {
	GetTasksByTime(ctx context.Context, key string, bucket int, start, end time.Time) ([]*vo.Task, error)
}

type confProvider interface {
	Get() *conf.TriggerAppConf
}
