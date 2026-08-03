package upload

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"hunjie/internal/config"
)

// 一期图片存本地磁盘，不接对象存储，也不做 Storage 接口抽象。
// 二期换 OSS 时改这一个文件即可。

var allowedScenes = map[string]bool{
	"avatar": true,
	"photo":  true,
	"chat":   true,
	"report": true,
}

// 按 magic number 判断真实类型，不信任扩展名
var allowedTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// SaveImage 保存图片，返回可访问的 URL
func SaveImage(fh *multipart.FileHeader, scene string) (string, error) {
	if !allowedScenes[scene] {
		return "", errors.New("不支持的上传场景")
	}

	maxSize := int64(config.C.Upload.MaxSizeMB) * 1024 * 1024
	if fh.Size > maxSize {
		return "", fmt.Errorf("图片不能超过 %dMB", config.C.Upload.MaxSizeMB)
	}
	if fh.Size == 0 {
		return "", errors.New("图片内容为空")
	}

	src, err := fh.Open()
	if err != nil {
		return "", fmt.Errorf("读取上传文件失败: %w", err)
	}
	defer src.Close()

	// 读前 512 字节判断真实类型
	head := make([]byte, 512)
	n, err := io.ReadFull(src, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("读取文件头失败: %w", err)
	}
	ext, ok := allowedTypes[strings.Split(http.DetectContentType(head[:n]), ";")[0]]
	if !ok {
		return "", errors.New("只支持 jpg / png / webp 格式的图片")
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("重置文件指针失败: %w", err)
	}

	// uploads/{scene}/YYYY/MM/DD/{uuid}.{ext}
	relDir := filepath.Join(scene, time.Now().Format("2006/01/02"))
	absDir := filepath.Join(config.C.Upload.Dir, relDir)
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return "", fmt.Errorf("创建上传目录失败: %w", err)
	}

	name := uuid.NewString() + ext
	dst, err := os.Create(filepath.Join(absDir, name))
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	relPath := filepath.ToSlash(filepath.Join(relDir, name))
	return strings.TrimRight(config.C.Upload.BaseURL, "/") + "/" + relPath, nil
}

// Delete 按 URL 删除本地文件。删不掉只当作没删，不影响业务
func Delete(fileURL string) {
	base := strings.TrimRight(config.C.Upload.BaseURL, "/")
	if !strings.HasPrefix(fileURL, base) {
		return
	}
	rel := strings.TrimPrefix(strings.TrimPrefix(fileURL, base), "/")
	if rel == "" || strings.Contains(rel, "..") {
		return
	}
	_ = os.Remove(filepath.Join(config.C.Upload.Dir, filepath.FromSlash(rel)))
}
