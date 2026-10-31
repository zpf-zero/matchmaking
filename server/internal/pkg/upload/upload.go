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

/*
 * 一期图片存本地磁盘，不接对象存储，也不做 Storage 接口抽象。
 * 2026-10-26 已决策：第一版图片直接落 81.70.160.36 的 /data/matchmaking/server/uploads，
 * 后续只在这台机器部署，因此不再为「多机/换存储」预留抽象。
 *
 * 已知代价（已接受）：① 单机故障即丢全部图片 ② 与 API 抢同一台机的带宽 ③ 无 CDN
 * 必须配套：uploads/ 目录纳入定期备份，否则磁盘故障后库里只剩 URL，无法恢复。
 * 若将来真要换 OSS，改动集中在本文件（上层只调 SaveImage / Delete）。
 *
 * 【待配置】upload.base_url —— 清单第 3 项
 *   域名备案完成后，需同步改 config.prod.yaml 的 upload.base_url 为 HTTPS 域名，
 *   否则新上传图片仍返回 HTTP + IP 地址，微信侧会拦截。
 */

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
