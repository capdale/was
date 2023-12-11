package imagequeue

import (
	"context"
	"fmt"
	"math"
	"os"
	"path"
	"sync"
	"time"

	"github.com/capdale/was/config"
	"github.com/capdale/was/logger"
	"github.com/capdale/was/model"
	rpcservice "github.com/capdale/was/rpc"
	rpc_protocol "github.com/capdale/was/rpc/proto"
	"github.com/capdale/was/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
)

type database interface {
	PopImageQueues(n int) (*[]model.ImageQueue, error)
	RecoverImageQueue(index uint) error
	DeleteImageQueues(index uint) error
	PutCollectionFromImageQueue(m *model.ImageQueue, index int64) error
}

type ImageQueue struct {
	logger                 *zap.Logger
	Duration               time.Duration
	DB                     database
	imageClassifierSevices *[]*rpcservice.ImageClassify
	S3                     *s3.BasicBucket
	once                   sync.Once
}

// change to imageSub (divide)

func New(d database, t time.Duration, imageClassifierSevices *[]*rpcservice.ImageClassify, isProduction bool, config *config.ImageQueue, s3 *s3.BasicBucket) *ImageQueue {
	queueLogger := logger.New(&lumberjack.Logger{
		Filename:   config.Log.Path,
		MaxSize:    config.Log.MaxSize,
		MaxBackups: config.Log.MaxBackups,
		MaxAge:     config.Log.MaxAge,
	}, isProduction, config.Log.Console)
	return &ImageQueue{
		logger:                 queueLogger,
		Duration:               t,
		DB:                     d,
		imageClassifierSevices: imageClassifierSevices,
		S3:                     s3,
	}
}

func (q *ImageQueue) Run(ctx *context.Context) {
	q.once.Do(func() {
		go q.mainRoutine(ctx)
	})
}

func (q *ImageQueue) mainRoutine(ctx *context.Context) {
	ch := make(chan *model.ImageQueue, len(*q.imageClassifierSevices))
	maxChannelN := len(*q.imageClassifierSevices)

	nSecond := int(math.Round(q.Duration.Seconds()))
	nthSleep := 0

	for _, imageService := range *q.imageClassifierSevices {
		go q.subRoutine(ctx, ch, imageService)
	}

	for {
		select {
		case <-(*ctx).Done():
			return
		default:
			images, err := q.DB.PopImageQueues(maxChannelN)
			// fatal error, need alert
			if err != nil {
				q.logger.Error(err.Error(), zap.String("worker", "master"))
				break
			}
			q.logger.Info(fmt.Sprintf("query %d images", len(*images)), zap.String("worker", "master"), zap.String("event", "query"))
			if len(*images) == 0 {
				ntimes := 1 << nthSleep
				q.logger.Info(fmt.Sprintf("query sleep for %d(s), %d(th) sleep", nSecond*ntimes, nthSleep), zap.String("worker", "master"), zap.String("event", "query"))
				time.Sleep(q.Duration * time.Duration(ntimes))
				if nthSleep < 1 {
					nthSleep += 1
				}
				break
			}
			for _, image := range *images {
				ch <- &image
			}
		}
	}
}

func (q *ImageQueue) subRoutine(ctx *context.Context, ch chan *model.ImageQueue, service *rpcservice.ImageClassify) {
	for {
		select {
		case <-(*ctx).Done():
			return
		case image := <-ch:
			imageBytes, err := q.getImageAsByte(&image.UUID)
			if err != nil {
				// log get image failed, err, need to trace in 10 hours
				q.logger.Error(fmt.Sprintf("read file (%s): %s", &image.UUID, err.Error()), zap.String("worker", "slave"), zap.String("event", "file"))
				if err := q.DB.RecoverImageQueue(image.ID); err != nil {
					// log recover failed, err
					q.logger.Error(fmt.Sprintf("recover queue (%s): %s", &image.UUID, err.Error()), zap.String("worker", "slave"), zap.String("event", "recover"))
				}
				break
			}
			reply, err := (*service.Client).ClassifyImage(*ctx, &rpc_protocol.ImageClassifierRequest{Image: *imageBytes})
			if err != nil {
				// log rpc server works bad, critical err
				q.logger.Error(fmt.Sprintf("rpc error (%s): %s", &image.UUID, err.Error()), zap.String("worker", "slave"), zap.String("event", "rpc"))
				if err := q.DB.RecoverImageQueue(image.ID); err != nil {
					// log recover failed, err
					q.logger.Error(fmt.Sprintf("recover queue (%s): %s", &image.UUID, err.Error()), zap.String("worker", "slave"), zap.String("event", "recover")) // TODO: change to log
				}
				break
			}

			q.logger.Info(fmt.Sprintf("rpc (%s): (%d)", image.UUID, reply.ClassIndex), zap.String("worker", "slave"), zap.String("event", "rpc"))

			err = q.S3.Upload(image.UUID.String(), imageBytes)
			if err != nil {
				if err := q.DB.RecoverImageQueue(image.ID); err != nil {
					// log recover failed, err
					q.logger.Error(fmt.Sprintf("recover queue (%s): %s", &image.UUID, err.Error()), zap.String("worker", "slave"), zap.String("event", "recover")) // TODO: change to log
				}
			}
			err = q.DB.PutCollectionFromImageQueue(image, reply.ClassIndex)
			if err != nil {
				q.logger.Error(fmt.Sprintf("Put Collection error (%s): %s, but process goes", &image.UUID, err.Error()), zap.String("worker", "slave"), zap.String("event", "collection"))
			}

			if err := q.DB.DeleteImageQueues(image.ID); err != nil {
				// log delete permanent error, but context goes, no break
				q.logger.Error(fmt.Sprintf("delete queue (%s): %s, but process goes", &image.UUID, err.Error()), zap.String("worker", "slave"), zap.String("event", "delete"))
			}
			// final task, remove file from storage
			err = q.RemoveImageFile(&image.UUID)
			if err != nil {
				// log this, not fatal, warning
				q.logger.Error(fmt.Sprintf("remove file (%s): %s, but process goes", &image.UUID, err.Error()), zap.String("worker", "slave"), zap.String("event", "file")) // TODO: chane to log
			}
		}
	}
}

func (q *ImageQueue) getImageAsByte(imagePath *uuid.UUID) (*[]byte, error) {
	bytes, err := os.ReadFile(path.Join("./secret", imagePath.String()))
	return &bytes, err
}

func (q *ImageQueue) RemoveImageFile(imagePath *uuid.UUID) error {
	return os.Remove(path.Join("./secret", imagePath.String()))
}
