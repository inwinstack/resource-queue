package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	_ "runtime"
	_ "time"

	"github.com/gorilla/mux"
	"github.com/kjelly/resource-queue/httpHandler"
	"github.com/kjelly/resource-queue/worker"
	"github.com/onrik/logrus/filename"

	log "github.com/sirupsen/logrus"
	ini "gopkg.in/ini.v1"
)

func setRouter(router *mux.Router, handler httpHandler.Handler) {
	router.HandleFunc("/"+handler.Kind()+"/{request_id}", handler.UpdateProperty).Methods("POST")
	router.HandleFunc("/"+handler.Kind()+"/{request_id}", handler.GetJobs).Methods("GET").Queries("owner_id", "owner_id").Queries("status", "status")
	router.HandleFunc("/"+handler.Kind()+"/{request_id}", handler.GetJobs).Methods("GET").Queries("owner_id", "owner_id")
	router.HandleFunc("/"+handler.Kind()+"/{request_id}", handler.GetJob).Methods("GET")
	router.HandleFunc("/"+handler.Kind()+"/{request_id}", handler.DeleteJob).Methods("DELETE")
	router.HandleFunc("/"+handler.Kind()+"/", handler.AddJob).Methods("POST")
	router.HandleFunc("/"+handler.Kind()+"/", handler.GetJobs).Methods("GET").Queries("owner_id", "owner_id")
	router.HandleFunc("/"+handler.Kind()+"/", handler.GetJobs).Methods("GET")
	router.HandleFunc("/"+handler.Kind()+"/{request_id}/test", handler.Test).Methods("GET")
}

func main() {
	log.SetLevel(log.DebugLevel)
	filenameHook := filename.NewHook()
	filenameHook.Field = "line" // Customize source field name
	log.AddHook(filenameHook)

	var INIPath string
	flag.StringVar(&INIPath, "path", "queue.ini", "config path")
	flag.Parse()
	cfg, err := ini.Load(INIPath)
	if err != nil {
		log.Warnf("Failed to read config. (%s)", err)
		log.Warn("Use the default value to run.")
		cfg = ini.Empty()
	}

	log.Debug("Start")
	databaseType := cfg.Section("database").Key("type").MustString("sqlite3")
	databaseURI := cfg.Section("database").Key("uri").MustString("test.db")
	vHandler := httpHandler.InitVMHandler(databaseType, databaseURI)
	VMWorker := worker.InitVMWorker(vHandler.GetQueue())
	//go VMWorker.Run()

	router := mux.NewRouter()
	setRouter(router, vHandler)
	srv := &http.Server{Addr: ":8080", Handler: router}
	go srv.ListenAndServe()
	log.Debug("Listenering")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	signal.Reset(os.Interrupt)

	fmt.Println("Got signal:", s)

	VMWorker.Stop()
	srv.Shutdown(nil)

}
