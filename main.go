package main

import (
	"JGLSite/api"
	"JGLSite/test"
	"JGLSite/utils"
	"github.com/JGLTechnologies/SimpleFiles"
	"github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

var store *persist.MemoryStore
var Projects []*api.Project

const port string = "81"

func main() {
	godotenv.Load("/var/www/.env")
	gin.SetMode(gin.ReleaseMode)
	r := multitemplate.NewRenderer()
	r.AddFromFiles("home", "go web files/home.html", "go web files/base.html")
	r.AddFromFiles("projects", "go web files/projects.html", "go web files/base.html")
	r.AddFromFiles("client-error", "go web files/client_error.html")
	r.AddFromFiles("contact", "go web files/contact.html", "go web files/base.html")
	r.AddFromFiles("status", "go web files/status.html")
	r.AddFromFiles("contact-thank-you", "go web files/thank-you.html")
	r.AddFromFiles("contact-limit", "go web files/limit.html")
	r.AddFromFiles("contact-captcha", "go web files/captcha.html")
	r.AddFromFiles("contact-bl", "go web files/bl.html")
	r.AddFromFiles("error", "go web files/error.html")
	r.AddFromFiles("bmi-home", "go web files/bmi/build/index.html")
	r.AddFromFiles("go-module", "go web files/go_module.html")
	router := gin.New()
	router.HTMLRender = r
	store = persist.NewMemoryStore(time.Hour)
	router.Use(gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		err = strings.Split(err.(error).Error(), "runtime/debug.Stack()")[0]
		if utils.StartsWith(c.Request.URL.String(), "/api") {
			c.AbortWithStatusJSON(500, gin.H{"error": err})
		} else {
			c.HTML(500, "error", gin.H{"error": err})
			c.AbortWithStatus(500)
		}
	}))
	router.Use(gin.Logger())
	router.HandleMethodNotAllowed = true
	f, _ := SimpleFiles.New("cloudflare_ips.txt")
	s, _ := f.ReadString()
	ips := strings.Split(s, "\n")
	ips = append(ips, "127.0.0.1")
	router.SetTrustedProxies(ips)
	router.ForwardedByClientIP = true
	router.RemoteIPHeaders = []string{"X-Forwarded-For"}

	router.GET("/", cache.CacheByRequestPath(store, time.Minute*10), home)
	router.GET("/bot", func(c *gin.Context) {
		c.String(200, "Bot webpage is coming soon.")
	})
	router.GET("/home", cache.CacheByRequestPath(store, time.Hour*24), home)
	router.GET("/projects", cache.CacheByRequestPath(store, time.Minute), projects)
	router.GET("/contact", cache.CacheByRequestPath(store, time.Hour*24), contact)
	router.GET("/logo.png", cache.CacheByRequestPath(store, time.Hour*24), logo)
	router.GET("/favicon.ico", cache.CacheByRequestPath(store, time.Hour*24), favicon)

	goGroup := router.Group("/go")
	{
		goGroup.GET("/:module", func(c *gin.Context) {
			c.HTML(200, "go-module", gin.H{"module": c.Param("module")})
		})
	}

	testGroup := router.Group("/test")
	{
		testGroup.GET("/bmi", cache.CacheByRequestPath(store, time.Hour*24), test.BMIHome)
		testGroup.GET("/bmi/static/main.js", cache.CacheByRequestPath(store, time.Hour*24), test.BMIJS)
	}

	apiGroup := router.Group("/api")
	{
		apiGroup.GET("/bot/status", cache.CacheByRequestPath(store, time.Minute), api.BotStatus)
		apiGroup.GET("/bot/info", cache.CacheByRequestPath(store, time.Hour), api.BotInfo)
		apiGroup.GET("/downloads", cache.CacheByRequestPath(store, time.Minute*10), api.Downloads)
		apiGroup.POST("/contact", utils.GetMW(time.Second, 1), api.Contact)
	}

	router.NoRoute(noRoute)
	router.NoMethod(noMethod)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	done := make(chan bool)
	go func() {
		p, _ := api.Projects()
		Projects = p
		done <- true
		close(done)
		time.Sleep(time.Minute * 10)
		for {
			p, _ := api.Projects()
			Projects = p
			time.Sleep(time.Minute * 10)
		}
	}()
	<-done
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}

func favicon(c *gin.Context) {
	c.File("static/favicon.ico")
}

func logo(c *gin.Context) {
	c.File("static/logo.png")
}

func home(c *gin.Context) {
	c.HTML(200, "home", gin.H{})
}

func contact(c *gin.Context) {
	c.HTML(200, "contact", gin.H{})
}

func projects(c *gin.Context) {
	data := gin.H{}
	var html template.HTML
	if len(Projects) < 1 {
		data["projects"] = template.HTML("<p>Projects could not be loaded.</p>")
	} else {
		for _, v := range Projects {
			if v.Name == "JGL-Site" || v.Private {
				continue
			} else {
				if v.Downloads != "" {
					html += template.HTML("<p class=\"lead fw-normal text-muted mb-0\">\n<br/>\n<span style='color: var(--bs-dark);'>" + v.Name + ":</span>\n<br/><span style=\"position: relative; left: 10px;\">Description: " + v.Description + "</span>\n<br/><span style='position: relative; left: 10px;'>Downloads: " + v.Downloads + " </span>\n<br/><span style='position: relative; left: 10px; top: 7px;'>Github URL: <a\nhref=https://github.com/JGLTechnologies/" + v.Name + " >click</a></span>\n</p>")
				} else {
					html += template.HTML("<p class=\"lead fw-normal text-muted mb-0\">\n<br/>\n<span style='color: var(--bs-dark);'>" + v.Name + ":</span>\n<br/><span style=\"position: relative; left: 10px;\">Description: " + v.Description + "</span>\n<br/><span style='position: relative; left: 10px; top: 7px;'>Github URL: <a\nhref=https://github.com/JGLTechnologies/" + v.Name + " >click</a></span>\n</p>")
				}
			}
		}
		data["projects"] = html
	}
	c.HTML(200, "projects", data)
}

func noRoute(c *gin.Context) {
	if utils.StartsWith(c.Request.URL.String(), "/api") {
		c.JSON(404, gin.H{"error": "Not Found"})
	} else {
		c.HTML(404, "status", gin.H{
			"code":    "404",
			"message": "The page you requested does not exist.",
		})
	}
}

func noMethod(c *gin.Context) {
	if utils.StartsWith(c.Request.URL.String(), "/api") {
		c.JSON(405, gin.H{"error": "Method Not Allowed"})
	} else {
		c.HTML(405, "status", gin.H{
			"code":    "405",
			"message": "The method you used is not allowed.",
		})
	}
}
