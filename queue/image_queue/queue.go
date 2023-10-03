package imagequeue

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/capdale/was/model"
	rpcservice "github.com/capdale/was/rpc"
	rpc_protocol "github.com/capdale/was/rpc/proto"
	"github.com/google/uuid"
)

type database interface {
	PopImageQueues(n int) (*[]model.ImageQueue, error)
	RecoverImageQueue(index uint) error
	DeleteImageQueues(index uint) error
}

type ImageQueue struct {
	Duration               time.Duration
	DB                     database
	imageClassifierSevices *[]*rpcservice.ImageClassify
	once                   sync.Once
}

// change to imageSub (divide)

func New(d database, t time.Duration, imageClassifierSevices *[]*rpcservice.ImageClassify) *ImageQueue {
	return &ImageQueue{
		Duration:               t,
		DB:                     d,
		imageClassifierSevices: imageClassifierSevices,
	}
}

func (q *ImageQueue) Run(ctx *context.Context) {
	q.once.Do(func() {
		go q.mainRoutine(ctx)
	})
}

func (q *ImageQueue) mainRoutine(ctx *context.Context) {
	ch := make(chan *model.ImageQueue, len(*q.imageClassifierSevices))
	ticker := time.NewTicker(q.Duration)
	maxChannelN := len(*q.imageClassifierSevices)
	defer func() {
		ticker.Stop()
	}()

	for _, imageService := range *q.imageClassifierSevices {
		go q.subRoutine(ctx, ch, imageService)
	}

	for {
		select {
		case <-ticker.C:
			images, err := q.DB.PopImageQueues(maxChannelN)
			// fatal error, need alert
			if err != nil {
				fmt.Printf("Image Queue Error: %s", err.Error())
				break
			}
			for _, image := range *images {
				ch <- &image
			}
		case <-(*ctx).Done():
			return
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
				if err := q.DB.RecoverImageQueue(image.ID); err != nil {
					// log recover failed, err
					fmt.Println("warning") // TODO: change to log
				}
				break
			}
			reply, err := (*service.Client).ClassifyImage(*ctx, &rpc_protocol.ImageClassifierRequest{Image: *imageBytes})
			if err != nil {
				// log rpc server works bad, critical err
				if err := q.DB.RecoverImageQueue(image.ID); err != nil {
					// log recover failed, err
					fmt.Println("warning") // TODO: change to log
				}
				break
			}
			if err := q.DB.DeleteImageQueues(image.ID); err != nil {
				// log delete permanent error
				fmt.Println("delete data error") // TODO: change to log
			}
			fmt.Println(reply.ClassIndex)
			// move to external storage
			// pub, classify event to subs
			// ...

			// final task, remove file from storage
			err = q.RemoveImageFile(&image.UUID)
			if err != nil {
				// log this, not fatal, warning
				fmt.Println("remove image file error") // TODO: chane to log
			}
		}
	}
}

func (q *ImageQueue) getImageAsByte(imagePath *uuid.UUID) (*[]byte, error) {
	bytes, err := os.ReadFile(path.Join("./secret", imagePath.String()))
	return &bytes, err
}

func (q *ImageQueue) RemoveImageFile(imagePath *uuid.UUID) error {
	return os.Remove(imagePath.String())
}
