package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var targetBucket string
var s3Client *s3.Client
var pollyClient *polly.Client

func init() {
	targetBucket = os.Getenv("TARGET_BUCKET_NAME")

	if targetBucket == "" {
		log.Fatal("missing environment variable TARGET_BUCKET_NAME")
	}
	fmt.Println("target S3 bucket", targetBucket)

	cfg, err := config.LoadDefaultConfig(context.Background())

	if err != nil {
		log.Fatal("failed to load config ", err)
	}

	s3Client = s3.NewFromConfig(cfg)
	pollyClient = polly.NewFromConfig(cfg)

}

func handler(ctx context.Context, s3Event events.S3Event) {
	for _, record := range s3Event.Records {

		fmt.Println("file", record.S3.Object.Key, "uploaded to", record.S3.Bucket.Name)

		sourceBucketName := record.S3.Bucket.Name
		fileName := record.S3.Object.Key

		err := textToSpeech(sourceBucketName, fileName)

		if err != nil {
			log.Fatal("failed to process file ", record.S3.Object.Key, " in bucket ", record.S3.Bucket.Name, err)
		}
	}
}

func main() {
	lambda.Start(handler)
}

func textToSpeech(sourceBucketName, textFileName string) error {

	voiceID := types.VoiceIdAmy
	outputFormat := types.OutputFormatMp3

	result, err := s3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(sourceBucketName),
		Key:    aws.String(textFileName),
	})

	if err != nil {
		return err
	}

	fmt.Println("successfully read file", textFileName, "from s3 bucket", sourceBucketName)

	buffer := new(bytes.Buffer)
	buffer.ReadFrom(result.Body)
	text := buffer.String()

	output, err := pollyClient.SynthesizeSpeech(context.Background(), &polly.SynthesizeSpeechInput{
		Text:         aws.String(text),
		OutputFormat: outputFormat,
		VoiceId:      voiceID,
	})

	if err != nil {
		return err
	}

	fmt.Println("successfully converted text to speech using polly")

	var buf bytes.Buffer
	_, err = io.Copy(&buf, output.AudioStream)
	if err != nil {
		return err
	}

	outputFileName := strings.Split(textFileName, ".")[0] + ".mp3"

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Body:        bytes.NewReader(buf.Bytes()),
		Bucket:      aws.String(targetBucket),
		Key:         aws.String(outputFileName),
		ContentType: output.ContentType,
	})

	if err != nil {
		return err
	}

	fmt.Println("successfully uploaded output mp3 file", outputFileName, "to s3 bucket", targetBucket)

	return nil
}
