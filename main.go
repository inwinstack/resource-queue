package main

import (
	"fmt"
	_ "log"
	"net/http"
	"os"
	"os/signal"
	_ "runtime"
	_ "time"

	"github.com/gorilla/mux"
	"github.com/kjelly/resource-queue/httpHandler"
	"github.com/kjelly/resource-queue/worker"
)

func setRouter(router *mux.Router, handler httpHandler.Handler) {
	router.HandleFunc("/"+handler.Kind()+"/{request_id}", handler.SetPriority).Methods("POST")
	router.HandleFunc("/"+handler.Kind()+"/{request_id}", handler.GetJobs).Methods("GET").Queries("uid", "uid")
	router.HandleFunc("/"+handler.Kind()+"/", handler.AddJob).Methods("POST")
	router.HandleFunc("/"+handler.Kind()+"/", handler.GetJobs).Methods("GET").Queries("owner_id", "owner_id")
	router.HandleFunc("/"+handler.Kind()+"/", handler.GetJobs).Methods("GET")
	router.HandleFunc("/"+handler.Kind()+"/{request_id}/test", handler.Test).Methods("GET")
}

func main() {
	fmt.Printf("Start\n")
	vHandler := httpHandler.InitVMHandler()
	VMWorker := worker.InitVMWorker(vHandler.GetQueue())
	go VMWorker.Run()

	router := mux.NewRouter()
	setRouter(router, vHandler)
	srv := &http.Server{Addr: ":8080", Handler: router}
	go srv.ListenAndServe()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	signal.Reset(os.Interrupt)

	fmt.Println("Got signal:", s)

	VMWorker.Stop()
	srv.Shutdown(nil)

}
