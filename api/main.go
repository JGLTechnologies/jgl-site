package api

import (
	"JGLSite/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req"
	"net/http"
	"time"
)

type postForm struct {
	Name    string `form:"name" binding:"required"`
	Email   string `form:"email" binding:"required"`
	Message string `form:"message" binding:"required"`
	Token   string `form:"token" binding:"required"`
}

func Contact(c *gin.Context) {
	formData := postForm{}
	if bindingErr := c.Bind(&formData); bindingErr != nil {
		c.HTML(400, "status", gin.H{"code": "400", "message": "Invalid request body"})
		return
	}
	name := formData.Name
	email := formData.Email
	message := formData.Message
	token := formData.Token
	if len(name) > 200 || len(email) > 254 || len(message) > 1020 {
		c.HTML(400, "status", gin.H{"code": "400", "message": "Invalid request body"})
		return
	}
	data := map[string]string{"name": name, "email": email, "message": message, "token": token, "ip": utils.GetIP(c)}
	client := http.Client{
		Timeout: time.Second * 5,
	}
	r := req.New()
	r.SetClient(&client)
	res, err := r.Post("https://jglbotapi.us/contact", req.BodyJSON(&data))
	if err != nil {
		c.HTML(500, "error", gin.H{"error": fmt.Sprintf("%s", err)})
	} else {
		var resJSON map[string]interface{}
		jsonErr := res.ToJSON(&resJSON)
		if jsonErr != nil {
			c.HTML(500, "error", gin.H{"error": fmt.Sprintf("%s", err)})
		} else {
			if res.Response().StatusCode == 200 {
				c.HTML(200, "contact-thank-you", gin.H{})
			} else if res.Response().StatusCode == 429 {
				c.HTML(429, "contact-limit", gin.H{"remaining": resJSON["remaining"]})
			} else if res.Response().StatusCode == 401 {
				c.HTML(401, "contact-captcha", gin.H{})
			} else if res.Response().StatusCode == 403 {
				c.HTML(403, "contact-bl", gin.H{})
			} else {
				c.HTML(500, "error", gin.H{"error": resJSON["error"]})
			}
		}
	}

}
