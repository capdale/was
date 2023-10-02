package imagequeue

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/capdale/was/model"
	rpc_protocol "github.com/capdale/was/rpc/proto"
	"github.com/google/uuid"
)

type database interface {
	GetImageQueues(n int) (*[]*model.ImageQueue, error)
}

type ImageQueue struct {
	Duration               time.Duration
	DB                     database
	ImageClassifierClients *[]*rpc_protocol.ImageClassifyClient
	once                   sync.Once
}

// change to imageSub (divide)
type imageSub struct {
	ImageClassifierClient *rpc_protocol.ImageClassifyClient
}

func New(d database, t time.Duration) *ImageQueue {
	return &ImageQueue{
		Duration: t,
		DB:       d,
	}
}

func (q *ImageQueue) Run(ctx *context.Context) {
	q.once.Do(func() {
		go q.mainRoutine(ctx)
	})
}

func (q *ImageQueue) mainRoutine(ctx *context.Context) {
	ch := make(chan *model.ImageQueue, len(*q.ImageClassifierClients))
	ticker := time.NewTicker(q.Duration)
	maxChannelN := len(*q.ImageClassifierClients)
	defer func() {
		ticker.Stop()
	}()

	for _, imageClient := range *q.ImageClassifierClients {
		go q.subRoutine(ctx, ch, imageClient)
	}

	for {
		select {
		case <-ticker.C:
			images, err := q.DB.GetImageQueues(maxChannelN)
			// fatal error, need alert
			if err != nil {
				fmt.Printf("Image Queue Error: %s", err.Error())
				break
			}
			for _, image := range *images {
				ch <- image
			}
		case <-(*ctx).Done():
			return
		}
	}
}

func (q *ImageQueue) subRoutine(ctx *context.Context, ch chan *model.ImageQueue, client *rpc_protocol.ImageClassifyClient) {
	for {
		select {
		case <-(*ctx).Done():
			return
		case image := <-ch:
			imageBytes, err := q.getImageAsByte(&image.UUID)
			if err != nil {
				// log fatal, can't get image from local storage
				break
			}
			reply, err := (*client).ClassifyImage(*ctx, &rpc_protocol.ImageClassifierRequest{Image: *imageBytes})
			if err != nil {
				// log fatal rpc server works bad
				break
			}
			fmt.Println(reply.ClassIndex)
			// move to external storage
			// pub, classify event to subs
		}
	}
}

func (q *ImageQueue) getImageAsByte(imagePath *uuid.UUID) (*[]byte, error) {
	bytes, err := os.ReadFile(path.Join("./secret", imagePath.String()))
	return &bytes, err
}
