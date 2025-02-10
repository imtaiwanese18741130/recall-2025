package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	c := NewController(cfg)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*.html")
	r.SetTrustedProxies(cfg.AppTrustedProxies)

	r.NoRoute(c.NotFound())
	r.GET("/health/v1/ping", c.Ping())
	r.GET("/", c.Home())
	r.GET("/filter", c.SearchZone())
	r.GET("/:stage/:zone", c.FillForm())
	r.POST("/:stage/:zone/preview", c.VerifyTurnstile(), c.PreviewLocalForm())
	r.GET("/:stage/:zone/thank-you", c.ThankYou())
	r.GET("/preview/:stage/:zone", c.PreviewOriginalLocalForm())
	r.GET("/robots.txt", c.RobotsTxt())
	r.GET("/sitemap.xml", c.Sitemap())
	r.GET("/assets/:type/:file", c.GetAsset())

	r.Run(":" + cfg.AppPort)
}
