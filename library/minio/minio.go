package minio

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"

	"tongcheng/conf"
	"tongcheng/logs"
)

var client *minio.Client
var bucket string

// Init 初始化MinIO客户端
func Init() error {
	config := conf.GetConfiguration().MinIO

	// 初始化MinIO客户端
	minioClient, err := minio.New(config.Addr, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Username, config.Password, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize minio client: %v", err)
	}

	client = minioClient
	bucket = config.Bucket

	// 检查bucket是否存在，不存在则创建
	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v", err)
		}
	}

	return nil
}

// UploadImage 上传图片到MinIO
func UploadImage(ctx context.Context, file multipart.File, filePath string, contentType string, fileSize int64) error {
	// 上传文件
	_, err := client.PutObject(ctx, bucket, filePath, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		logs.Error(ctx, "Failed to upload image to MinIO", zap.Error(err))
		return err
	}

	return nil
}

// DownloadImage 从MinIO下载图片
func DownloadImage(ctx context.Context, filePath string) (io.Reader, error) {
	object, err := client.GetObject(ctx, bucket, filePath, minio.GetObjectOptions{})
	if err != nil {
		logs.Error(ctx, "Failed to download image from MinIO", zap.Error(err))
		return nil, err
	}

	// 检查对象是否存在
	_, err = object.Stat()
	if err != nil {
		logs.Error(ctx, "Image not found in MinIO", zap.Error(err))
		return nil, err
	}

	return object, nil
}

// GetFileURL 获取文件访问URL
func GetFileURL(ctx context.Context, filePath string) string {
	reqParams := make(url.Values)
	preURL, err := client.PresignedGetObject(ctx, bucket, filePath, time.Hour*24*7, reqParams) // 7天有效期
	if err != nil {
		logs.Error(ctx, "Failed to generate preURL", zap.Error(err))
		return ""
	}

	return preURL.String()
}

// DeleteImage 从MinIO删除图片
func DeleteImage(ctx context.Context, filePath string) error {
	err := client.RemoveObject(ctx, bucket, filePath, minio.RemoveObjectOptions{})
	if err != nil {
		logs.Error(ctx, "Failed to delete image from MinIO", zap.Error(err))
		return err
	}

	return nil
}
