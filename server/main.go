package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net"
	"net/http"
	"summerflower.local/picture_api/server/config"
	"summerflower.local/picture_api/server/logger"
	"summerflower.local/picture_api/server/route"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()

	logger.InitLogger()
	defer func(Log *zap.Logger) {
		err := Log.Sync()
		if err != nil {

		}
	}(logger.Log)
	app.Use(logger.GinLogger())

	app.Use(Cors())
	route.SetupRouter(app)

	app.NoRoute(NoResponse)

	var (
		ipv4 = ""
		ipv6 = ""
	)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Log.Warn("获取本机ip失败")
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipv4 = ipnet.IP.String()
			}
			if ipnet.IP.To16() != nil && ipnet.IP.To4() == nil {
				ipv6 = ipnet.IP.String()
			}
		}
	}

	logger.Log.Info("服务启动，你可以访问以下地址：")
	logger.Log.Info("http://localhost:" + config.Config.Server.Port)
	logger.Log.Info("http://" + ipv4 + ":" + config.Config.Server.Port)
	logger.Log.Info("http://[" + ipv6 + "]:" + config.Config.Server.Port)
	app.Run(config.Config.Server.Host + ":" + config.Config.Server.Port).Error()
}

func NoResponse(c *gin.Context) {
	c.String(http.StatusNotFound, "页面找不到啦")
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
