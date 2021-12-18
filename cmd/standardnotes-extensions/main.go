package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.senan.xyz/standardnotes-extensions/pkg/controller"
)

func mustEnv(key string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	log.Fatalf("please provide a %q", key)
	return ""
}

func updateExtensions(fn func() error, wait time.Duration) {
	for {
		if err := fn(); err != nil {
			log.Printf("error updating extensions: %v", err)
		}
		time.Sleep(wait)
		log.Print("finished updating extensions")
	}
}

func main() {
	cfgListenAddr := mustEnv("SN_EXTS_LISTEN_ADDR")
	cfgUpdateIntervalMins := mustEnv("SN_EXTS_UPDATE_INTERVAL_MINS")
	ctrl := &controller.Controller{
		BaseURL:  mustEnv("SN_EXTS_BASE_URL"),
		ReposDir: mustEnv("SN_EXTS_REPOS_DIR"),
	}
	cfgUpdateInterval, err := strconv.Atoi(cfgUpdateIntervalMins)
	if err != nil {
		log.Fatalf("error parsing update interval: %v", err)
	}
	go updateExtensions(ctrl.UpdateExtensions, time.Duration(cfgUpdateInterval)*time.Minute)

	indexHandler, err := ctrl.ServeIndex()
	if err != nil {
		log.Fatalf("error creating index handler: %v", err)
	}
	webHandler, err := ctrl.ServeWeb()
	if err != nil {
		log.Fatalf("error creating web handler: %v", err)
	}

	r := mux.NewRouter()
	r.Handle("/", indexHandler).Methods(http.MethodGet)
	r.PathPrefix("/web/").Handler(webHandler).Methods(http.MethodGet)

	r.HandleFunc("/{id}/index.json", ctrl.ServeExtensionIndex).Methods(http.MethodGet, http.MethodOptions)
	r.PathPrefix("/{id}/{version}/").HandlerFunc(ctrl.ServeExtension).Methods(http.MethodGet, http.MethodOptions)

	// very lazy cors
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join([]string{http.MethodGet, http.MethodOptions}, ","))
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			next.ServeHTTP(w, req)
		})
	})

	server := http.Server{
		Addr:    cfgListenAddr,
		Handler: r,
	}
	log.Printf("listening on %q", cfgListenAddr)
	log.Fatalf("error starting server: %v", server.ListenAndServe())
}
