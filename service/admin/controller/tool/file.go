package tool

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"tongcheng/library/ginx"
	"tongcheng/library/minio"
	"tongcheng/utils/constant"
)

func View(c *gin.Context) {
	// 从查询参数获取文件名
	filePath := c.Query("filePath")
	if filePath == "" {
		ginx.ResponseParamError(c)
		return
	}

	paths := strings.Split(filePath, "/")
	if len(paths) != 4 {
		ginx.ResponseParamError(c)
		return
	}

	// 从MinIO下载文件
	reader, err := minio.DownloadImage(c, filePath)
	if err != nil {
		ginx.ResponseError(c, constant.CodeFileNotFound)
		return
	}

	// 获取文件扩展名以确定内容类型
	ext := strings.ToLower(filepath.Ext(filePath))
	contentType := GetContentType(ext)

	// 设置响应头并返回文件内容
	filename := paths[len(paths)-1]
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s", filename))
	c.DataFromReader(http.StatusOK, -1, contentType, reader, nil)
}

func GetContentType(ext string) string {
	contentType := ""
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	case ".bmp":
		contentType = "image/bmp"
	default:
		contentType = "application/octet-stream"
	}

	return contentType
}
