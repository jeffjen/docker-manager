package web

import (
	api "github.com/jeffjen/docker-manager/web/api"
	dsc "github.com/jeffjen/go-discovery/info"

	log "github.com/Sirupsen/logrus"

	"net/http"
)

func init() {
	server := api.GetServeMux()

	server.HandleFunc("/info", dsc.Info)

	server.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("html/bower_components/"))))
	server.Handle("/", http.FileServer(http.Dir("html/")))

	server.HandleFunc("/cluster/list", api.ClusterList)
	server.HandleFunc("/cluster", api.ClusterCreate)
}

func RunAPIEndpoint(addr string) {
	server := api.GetServer()
	server.Addr = addr
	log.Error(server.ListenAndServe())
}
