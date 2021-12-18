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

func updatePackages(fn func() error, wait time.Duration) {
	for {
		if err := fn(); err != nil {
			log.Printf("error updating packages: %v", err)
		}
		time.Sleep(wait)
		log.Print("finished updating packages")
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
	go updatePackages(ctrl.UpdatePackages, time.Duration(cfgUpdateInterval)*time.Minute)

	r := mux.NewRouter()
	r.HandleFunc("/index.json", ctrl.ServeIndex).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/{id}/index.json", ctrl.ServePackageIndex).Methods(http.MethodGet, http.MethodOptions)
	r.PathPrefix("/{id}/{version}/").HandlerFunc(ctrl.ServePackage).Methods(http.MethodGet, http.MethodOptions)

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
