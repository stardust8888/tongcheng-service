package file

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"tongcheng/library/ginx"
	"tongcheng/library/httpclient/wxapi"
	"tongcheng/library/minio"
	"tongcheng/library/snowflake"
	"tongcheng/logs"
	"tongcheng/utils/constant"
)

// View 查看图片接口
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

// Upload 上传图片接口
func Upload(c *gin.Context) {
	userID := ginx.GetUserID(c)
	sense := c.Query("sense")
	if sense == "" {
		ginx.ResponseError(c, constant.CodeParamError)
		return
	}
	if !strings.HasPrefix(sense, "/") {
		sense = "/" + sense
	}

	// 从表单中获取文件
	file, err := c.FormFile("file")
	if err != nil {
		ginx.ResponseError(c, constant.CodeFileNotFound)
		return
	}

	// 检查文件大小（例如限制为3MB）
	if file.Size > 3*constant.MB {
		ginx.ResponseError(c, constant.CodeFileSizeExceeded)
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		ginx.ResponseError(c, constant.CodeFileCannotRead)
		return
	}
	defer func() {
		_ = src.Close()
	}()

	// 生成唯一文件名
	fileExt := strings.ToLower(filepath.Ext(file.Filename))
	filePath := fmt.Sprintf("%s/%d/%s%s", sense, userID, snowflake.GenerateString(), fileExt)

	// 确定内容类型
	contentType := file.Header.Get(constant.HeaderContentType)
	if contentType == "" {
		contentType = GetContentType(fileExt)
	}

	// 内容安全：写 MinIO 前对图片做微信 img_sec_check；命中违规拒绝上传
	// (图片字节已在手，此处检测最直接，无需回捞 MinIO；亦早于 child create/update)
	imgBytes, err := io.ReadAll(src)
	if err != nil {
		ginx.ResponseError(c, constant.CodeFileCannotRead)
		return
	}
	_, _ = src.Seek(0, io.SeekStart)
	risky, secErr := wxapi.GetClient().ImgSecCheck(c, imgBytes, "media"+fileExt)
	if secErr != nil {
		logs.Error(c, "img sec check fail-open", zap.Error(secErr))
	}
	if risky {
		ginx.ResponseNotify(c, "图片含违规内容，请重新选择")
		return
	}

	// 上传到MinIO
	err = minio.UploadImage(c, src, filePath, contentType, file.Size)
	if err != nil {
		ginx.ResponseError(c, constant.CodeFileUploadError)
		return
	}

	resp := map[string]any{
		"filePath": filePath,
	}

	ginx.ResponseSuccess(c, resp)
}

type DeleteFileRequest struct {
	FilePath string `json:"filePath" binding:"required"`
}

func Delete(c *gin.Context) {
	userID := ginx.GetUserID(c)

	req := DeleteFileRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		ginx.ResponseParamError(c)
		return
	}

	filePath := req.FilePath
	paths := strings.Split(filePath, "/")
	if len(paths) != 4 {
		ginx.ResponseParamError(c)
		return
	}

	fileUserID := paths[len(paths)-2]
	if fileUserID != strconv.Itoa(userID) {
		ginx.ResponseError(c, constant.CodeUnauthorized)
		return
	}

	err = minio.DeleteImage(c, filePath)
	if err != nil {
		logs.Error(c, "failed to delete image from MinIO", zap.Error(err), zap.String("filePath", filePath))
		ginx.ResponseServerError(c, err)
		return
	}

	ginx.ResponseSuccess(c, nil)
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
