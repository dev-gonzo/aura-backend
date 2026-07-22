package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type StorageUploadInput struct {
	Key           string
	Body          io.Reader
	ContentType   string
	CacheControl  string
	ContentLength int64
}

type StorageObject struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	URL    string `json:"url"`
}

type ObjectStorage interface {
	Upload(ctx context.Context, input StorageUploadInput) (StorageObject, error)
	Delete(ctx context.Context, key string) error
}

type SupabaseS3Storage struct {
	client        *s3.Client
	bucket        string
	publicBaseURL string
}

func (cfg AppConfig) HasSupabaseStorageConfig() bool {
	values := []string{
		cfg.SupabaseS3Endpoint,
		cfg.SupabaseS3Bucket,
		cfg.SupabaseS3AccessKey,
		cfg.SupabaseS3SecretKey,
		cfg.SupabasePublicURL,
	}

	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}

	return false
}

func (cfg AppConfig) ValidateSupabaseStorageConfig() error {
	if strings.TrimSpace(cfg.SupabaseS3Endpoint) == "" {
		return errors.New("SUPABASE_STORAGE_S3_ENDPOINT e obrigatorio")
	}

	if strings.TrimSpace(cfg.SupabaseS3Bucket) == "" {
		return errors.New("SUPABASE_STORAGE_BUCKET e obrigatorio")
	}

	if strings.TrimSpace(cfg.SupabaseS3AccessKey) == "" {
		return errors.New("SUPABASE_STORAGE_ACCESS_KEY_ID e obrigatorio")
	}

	if strings.TrimSpace(cfg.SupabaseS3SecretKey) == "" {
		return errors.New("SUPABASE_STORAGE_SECRET_ACCESS_KEY e obrigatorio")
	}

	return nil
}

func NewSupabaseS3Storage(ctx context.Context, cfg AppConfig) (*SupabaseS3Storage, error) {
	if err := cfg.ValidateSupabaseStorageConfig(); err != nil {
		return nil, err
	}

	creds := credentials.NewStaticCredentialsProvider(
		strings.TrimSpace(cfg.SupabaseS3AccessKey),
		strings.TrimSpace(cfg.SupabaseS3SecretKey),
		"",
	)

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(strings.TrimSpace(cfg.SupabaseS3Region)),
		awsconfig.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar configuracao AWS para o Supabase: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(strings.TrimSpace(cfg.SupabaseS3Endpoint))
		options.UsePathStyle = true
	})

	return &SupabaseS3Storage{
		client:        client,
		bucket:        strings.TrimSpace(cfg.SupabaseS3Bucket),
		publicBaseURL: strings.TrimSpace(cfg.SupabasePublicURL),
	}, nil
}

func (storage *SupabaseS3Storage) Upload(ctx context.Context, input StorageUploadInput) (StorageObject, error) {
	key := strings.TrimSpace(input.Key)
	if key == "" {
		return StorageObject{}, errors.New("a chave do objeto e obrigatoria")
	}

	if input.Body == nil {
		return StorageObject{}, errors.New("o corpo do arquivo e obrigatorio")
	}

	request := &s3.PutObjectInput{
		Bucket: aws.String(storage.bucket),
		Key:    aws.String(key),
		Body:   input.Body,
	}

	if strings.TrimSpace(input.ContentType) != "" {
		request.ContentType = aws.String(strings.TrimSpace(input.ContentType))
	}

	if strings.TrimSpace(input.CacheControl) != "" {
		request.CacheControl = aws.String(strings.TrimSpace(input.CacheControl))
	}

	if input.ContentLength > 0 {
		request.ContentLength = aws.Int64(input.ContentLength)
	}

	if _, err := storage.client.PutObject(ctx, request); err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) && apiError.ErrorCode() == "NoSuchBucket" {
			return StorageObject{}, fmt.Errorf(
				"bucket '%s' nao encontrado no Supabase Storage; crie esse bucket no painel do Supabase ou ajuste SUPABASE_STORAGE_BUCKET",
				storage.bucket,
			)
		}

		return StorageObject{}, fmt.Errorf("erro ao enviar arquivo para o Supabase Storage: %w", err)
	}

	return StorageObject{
		Bucket: storage.bucket,
		Key:    key,
		URL:    storage.PublicURL(key),
	}, nil
}

func (storage *SupabaseS3Storage) Delete(ctx context.Context, key string) error {
	normalizedKey := strings.TrimSpace(key)
	if normalizedKey == "" {
		return errors.New("a chave do objeto e obrigatoria")
	}

	if _, err := storage.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(storage.bucket),
		Key:    aws.String(normalizedKey),
	}); err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			if apiError.ErrorCode() == "NoSuchBucket" {
				return fmt.Errorf(
					"bucket '%s' nao encontrado no Supabase Storage; crie esse bucket no painel do Supabase ou ajuste SUPABASE_STORAGE_BUCKET",
					storage.bucket,
				)
			}

			if apiError.ErrorCode() == "NoSuchKey" {
				return nil
			}
		}

		return fmt.Errorf("erro ao excluir arquivo do Supabase Storage: %w", err)
	}

	return nil
}

func (storage *SupabaseS3Storage) PublicURL(key string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(storage.publicBaseURL), "/")
	normalizedKey := strings.TrimLeft(strings.TrimSpace(key), "/")
	if baseURL == "" || normalizedKey == "" {
		return ""
	}

	return baseURL + "/" + normalizedKey
}
