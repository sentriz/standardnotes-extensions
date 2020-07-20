package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.senan.xyz/standardnotes-extensions/controller"
	"go.senan.xyz/standardnotes-extensions/definition"
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
		BaseURL:        mustEnv("SN_EXTS_BASE_URL"),
		ReposDir:       mustEnv("SN_EXTS_REPOS_DIR"),
		DefinitionsDir: mustEnv("SN_EXTS_DEFINITIONS_DIR"),
		Packages:       map[string]*definition.Package{},
	}
	cfgUpdateInterval, err := strconv.Atoi(cfgUpdateIntervalMins)
	if err != nil {
		log.Fatalf("error parsing update interval: %v", err)
	}
	go updatePackages(ctrl.UpdatePackages, time.Duration(cfgUpdateInterval)*time.Minute)
	//
	r := mux.NewRouter()
	r.HandleFunc("/index.json", ctrl.ServeIndex)
	r.HandleFunc("/{id}/index.json", ctrl.ServePackageIndex)
	r.PathPrefix("/{id}/{version}/").HandlerFunc(ctrl.ServePackage)
	withCORS := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"DNT", "User-Agent", "X-Requested-With", "If-Modified-Since", "Cache-Control", "Content-Type", "Range"}),
	)
	withLogging := handlers.LoggingHandler
	server := http.Server{
		Addr:    cfgListenAddr,
		Handler: withLogging(os.Stdout, withCORS(r)),
	}
	log.Printf("listening on %q", cfgListenAddr)
	log.Fatalf("error starting server: %v", server.ListenAndServe())
}
