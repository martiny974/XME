package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xpzouying/xiaohongshu-mcp/configs"
)

// ImageProcessor 图片处理器
type ImageProcessor struct {
	downloader *ImageDownloader
}

// NewImageProcessor 创建图片处理器
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		downloader: NewImageDownloader(configs.GetImagesPath()),
	}
}

// ProcessImages 处理图片列表，返回本地文件路径
// 支持两种输入格式：
// 1. URL格式 (http/https开头) - 自动下载到本地
// 2. 本地文件路径 - 验证后使用
func (p *ImageProcessor) ProcessImages(images []string) ([]string, error) {
	var localPaths []string
	var urlsToDownload []string
	var invalidPaths []string

	// 分离URL和本地路径
	for _, image := range images {
		if IsImageURL(image) {
			urlsToDownload = append(urlsToDownload, image)
		} else {
			// 验证本地路径
			if isValidLocalPath(image) {
				localPaths = append(localPaths, image)
			} else {
				invalidPaths = append(invalidPaths, image)
			}
		}
	}

	// 如果有无效的本地路径，返回错误
	if len(invalidPaths) > 0 {
		return nil, fmt.Errorf("invalid local file paths (file not found or not an image): %v", invalidPaths)
	}

	// 批量下载URL图片
	if len(urlsToDownload) > 0 {
		downloadedPaths, err := p.downloader.DownloadImages(urlsToDownload)
		if err != nil {
			return nil, fmt.Errorf("failed to download images: %w", err)
		}
		localPaths = append(localPaths, downloadedPaths...)
	}

	if len(localPaths) == 0 {
		return nil, fmt.Errorf("no valid images found")
	}

	return localPaths, nil
}

// isValidLocalPath 验证本地文件路径是否有效
func isValidLocalPath(path string) bool {
	// 检查路径是否为空
	if strings.TrimSpace(path) == "" {
		return false
	}

	// 清理路径（处理混合分隔符和特殊字符）
	cleanPath := filepath.Clean(path)
	
	// 检查文件是否存在
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return false
	}

	// 检查是否为文件（不是目录）
	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		return false
	}
	
	if fileInfo.IsDir() {
		return false
	}

	// 检查文件扩展名是否为图片格式
	ext := strings.ToLower(filepath.Ext(cleanPath))
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	
	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}

	return false
}
