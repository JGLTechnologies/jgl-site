package main

import (
	"JGLSite/api"
	"JGLSite/test"
	"JGLSite/utils"
	"fmt"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := multitemplate.NewRenderer()
	r.AddFromFiles("home", "go web files/home.html", "go web files/base.html")
	r.AddFromFiles("contact", "go web files/contact.html", "go web files/base.html")
	r.AddFromFiles("404", "go web files/404.html")
	r.AddFromFiles("bmi-home", "go web files/bmi/index.html")
	r.AddFromFiles("bmi-calc", "go web files/bmi/bmi.html")
	r.AddFromFiles("bmi-invalid", "go web files/bmi/invalid.html")
	r.AddFromFiles("contact-thank-you", "go web files/thank-you.html")
	r.AddFromFiles("contact-limit", "go web files/limit.html")
	r.AddFromFiles("contact-captcha", "go web files/captcha.html")
	r.AddFromFiles("contact-bl", "go web files/bl.html")
	r.AddFromFiles("contact-error", "go web files/error.html")

	server := gin.New()
	server.HTMLRender = r
	store := persistence.NewInMemoryStore(time.Hour)

	server.Use(gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		c.HTML(500, "contact-error", gin.H{"error": fmt.Sprintf("%s", err)})
		c.AbortWithStatus(500)
	}))
	server.Use(utils.LoggerWithConfig(gin.LoggerConfig{}))
	server.SetTrustedProxies([]string{"192.168.1.252", "127.0.0.1", "192.168.1.1"})

	server.GET("/", cache.CachePageWithoutQuery(store, time.Hour, home))
	server.GET("/home", cache.CachePageWithoutQuery(store, time.Hour, home))
	server.GET("/contact", cache.CachePageWithoutQuery(store, time.Hour, contact))
	server.GET("/bot", func(c *gin.Context) {
		c.String(200, "JGL Bot documentation is coming soon.")
	})

	testGroup := server.Group("/test")
	{
		testGroup.GET("/bmi", test.BMIHome)
		testGroup.GET("/bmi/calc", test.BMICalc)
	}

	apiGroup := server.Group("/api")
	{
		apiGroup.GET("/bot/status", cache.CachePageWithoutQuery(store, time.Minute, api.BotStatus))
		apiGroup.GET("/bot/info", cache.CachePageWithoutQuery(store, time.Hour, api.BotInfo))
		apiGroup.GET("/dpys", cache.CachePageWithoutQuery(store, time.Minute*10, api.DPYS))
		apiGroup.GET("/aiohttplimiter", cache.CachePageWithoutQuery(store, time.Minute*10, api.AIOHTTPRateLimiter))
		apiGroup.POST("/contact", utils.GetMW(1, 1), api.Contact)
	}

	server.NoRoute(noRoute)
	server.NoMethod(noRoute)
	log.Fatal(server.Run(":81"))
}

func home(c *gin.Context) {
	c.HTML(200, "home", gin.H{})
}

func contact(c *gin.Context) {
	c.HTML(200, "contact", gin.H{})
}

func noRoute(c *gin.Context) {
	c.HTML(404, "404", gin.H{})
}
