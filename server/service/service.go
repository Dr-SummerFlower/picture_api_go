package service

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"summerflower.local/picture_api/server/logger"
)

var images = "resources/images/"

func GetRandom(c *gin.Context) {
	maxId := getMaxId()
	id := strconv.Itoa(rand.Intn(maxId) + 1)

	imagePath := images + id + ".jpg"

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "图片不存在"})
		return
	}

	etag, err := generateETag(imagePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法生成 ETag"})
		return
	}

	if match := c.GetHeader("If-None-Match"); match != "" {
		if match == etag {
			c.Status(http.StatusNotModified)
			return
		}
	}

	c.Header("Expires", "0")
	c.Header("Pragma", "no-cache")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("ETag", etag)
	c.Header("Content-Type", "image/jpeg")
	c.Header("Content-Disposition", "inline; filename=\""+url.QueryEscape(id+".jpg")+"\"")

	c.File(imagePath)
}

func GetImageById(c *gin.Context) {
	id := c.Param("id")
	imagePath := images + id + ".jpg"

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "图片不存在"})
		return
	}

	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Header("Content-Type", "image/jpeg")
	c.Header("Content-Disposition", "inline; filename=\""+url.QueryEscape(id+".jpg")+"\"")
	c.File(imagePath)
}

func getMaxId() int {
	maxID := 0
	err := filepath.Walk(images, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Log.Error("无法遍历文件", zap.Error(err))
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".jpg" {
			id, err := strconv.Atoi(info.Name()[:len(info.Name())-4])
			if err != nil {
				logger.Log.Error("无法解析文件名", zap.Error(err))
				return err
			}
			if id > maxID {
				maxID = id
			}
		}
		return nil
	})
	if err != nil || maxID == 0 {
		return maxID + 1
	} else {
		return maxID
	}
}

func generateETag(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Log.Error("无法打开文件", zap.Error(err))
		return "", err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Log.Error("无法关闭文件", zap.Error(err))
		}
	}(file)

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		logger.Log.Error("无法计算文件哈希值", zap.Error(err))
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
