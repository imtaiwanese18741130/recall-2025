package main

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		panic("template parse error: " + err.Error())
	}

	ctrl := NewController(cfg, tmpl)
	r := NewRouteRegistrar()
	r.Set404Response(ctrl.NotFound())

	r.GET("/health/v1/ping", ctrl.Ping())
	r.GET("/robots.txt", ctrl.RobotsTxt())
	r.GET("/sitemap.xml", ctrl.Sitemap())
	r.GET("/assets/", ctrl.GetAsset())

	r.GET("/{$}", ctrl.Home())
	r.GET("/authorization-letter", ctrl.AuthorizationLetter())
	r.GET("/apis/municipalities", ctrl.ListMunicipalities())
	r.GET("/apis/constituencies", ctrl.SearchRecallConstituency())
	r.GET("/preview/{name}", ctrl.PreviewOriginalLocalForm())
	r.GET("/legislators/{name}", ctrl.Participate())
	r.POST("/legislators/{name}/preview", ctrl.Preview())
	r.GET("/legislators/{name}/thank-you", ctrl.ThankYou())

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      logRequest(r.GetMux()),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Listening on port %s", cfg.AppPort)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
