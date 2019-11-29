package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

//跨域中间件
func CorsMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		var filterHost = [...]string{"http://localhost.*"}

		//阻止不合法的域名访问
		var isAccess = false

		for _, v := range filterHost {
			match, _ := regexp.MatchString(v, origin)
			if match {
				isAccess = true
			}
		}

		if isAccess {
			c.Header("Access-Control-Allow-Origin", "http://localhost:2233")
			c.Header("Access-Control-Allow-Headers", "emaccesstk,Origin, X-Requested-With, Content-Type, Accept")
			c.Header("Access-Control-Allow-Methods", "GET, OPTIONS, POST, PUT, DELETE")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Set("content-type", "application/json")
		}

		//放行所有option方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		c.Next()
	}
}
