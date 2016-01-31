package web

import (
	dsc "github.com/jeffjen/go-discovery/info"
	api "github.com/jeffjen/podd/web/api"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"

	"net/http"
)

var (
	router = gin.Default()
)

func init() {
	// Register static polymer assets
	router.StaticFS("/assets", api.Dir(http.Dir("html/bower_components/"), http.Dir("html/custom_components/")))

	// Register static html resources
	router.StaticFS("/css", http.Dir("html/www/css/"))
	router.StaticFS("/img", http.Dir("html/www/img/"))
	router.StaticFS("/js", http.Dir("html/www/js/"))

	// Index to application
	router.StaticFile("/", "html/www/app.html")

	// Report node metadata
	router.GET("/info", gin.WrapF(dsc.Info))

	// API for cluster
	var cluster = router.Group("/cluster")
	{
		cluster.GET("/list", api.ClusterList)
		cluster.POST("/:name", api.ClusterCreate)
		cluster.PUT("/:name", api.ClusterUpdate)
	}

	// API for service
	var service = router.Group("/service")
	{
		service.GET("/list", api.ServiceList)
	}
}

func RunAPIEndpoint(addr string) {
	log.Error(http.ListenAndServe(addr, router))
}
