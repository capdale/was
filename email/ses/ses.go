package ses

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/capdale/was/config"
)

// aws ses to send email

type AwsSes struct {
	client *ses.Client
	domain string
	cache  fmtCache
}

type fmtCache struct {
	noReply string // no reply address caching
}

func New(sesconfig *config.Ses) (*AwsSes, error) {
	var credProvider aws.CredentialsProvider
	if sesconfig.Id == nil && sesconfig.Key == nil {
		credProvider = ec2rolecreds.New()
	} else {
		if sesconfig.Id == nil || sesconfig.Key == nil {
			return nil, config.ErrInvalidCredForm
		}
		credProvider = credentials.NewStaticCredentialsProvider(*sesconfig.Id, *sesconfig.Key, "")
	}
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(sesconfig.Region), awsConfig.WithCredentialsProvider(credProvider))
	if err != nil {
		return nil, err
	}
	sesClient := ses.NewFromConfig(cfg)
	domain := sesconfig.Domain
	return &AwsSes{
		client: sesClient,
		domain: domain,
		cache: fmtCache{
			noReply: fmt.Sprintf(noReplyFmt, domain),
		},
	}, nil
}

const noReplyFmt = "no-reply@%s"

type accountTicketPayload struct {
	VerifyLink string `json:"verifylink"`
}

func (a *AwsSes) SendTicketVerifyLink(ctx context.Context, email string, link string) error {
	p := &accountTicketPayload{
		VerifyLink: link,
	}

	pbytes, err := json.Marshal(p)
	if err != nil {
		return err
	}

	_, err = a.client.SendTemplatedEmail(ctx, &ses.SendTemplatedEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{email},
		},
		Source:       aws.String(a.cache.noReply),
		Template:     aws.String("AccountTicketTemplate"),
		TemplateData: aws.String(string(pbytes)),
	})
	return err
}
