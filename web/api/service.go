package api

import (
	_ "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

func ServiceList(c *gin.Context) {
	// TODO: report the list of service currently provisioned
	c.File("html/www/service-list.json")
}
