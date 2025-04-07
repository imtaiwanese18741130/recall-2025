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
	if err := ctrl.CalcDaysLeft(); err != nil {
		panic("calc days left error: " + err.Error())
	}
	go func() {
		loc, err := time.LoadLocation("Asia/Taipei")
		if err != nil {
			log.Println("failed to load location:", err)
			return
		}

		for {
			now := time.Now().In(loc)

			next := time.Date(
				now.Year(), now.Month(), now.Day()+1,
				0, 0, 0, 0, loc,
			)

			duration := time.Until(next)
			time.Sleep(duration)

			if err := ctrl.CalcDaysLeft(); err != nil {
				log.Println("CalcDaysLeft error:", err)
			}
		}
	}()

	mux := http.NewServeMux()

	mux.HandleFunc("/health/v1/ping", withRecovery(ctrl.Ping))
	mux.HandleFunc("/robots.txt", withRecovery(ctrl.RobotsTxt))
	mux.HandleFunc("/sitemap.xml", withRecovery(ctrl.Sitemap))
	mux.HandleFunc("/assets/", withRecovery(ctrl.GetAsset))

	mux.HandleFunc("/", withRecovery(ctrl.Home))
	mux.HandleFunc("/authorization-letter", withRecovery(ctrl.AuthorizationLetter))
	mux.HandleFunc("/apis/constituencies", withRecovery(ctrl.SearchRecallConstituency))
	mux.HandleFunc("/preview/stages/", withRecovery(ctrl.PreviewOriginalLocalForm))
	mux.HandleFunc("/legislators/", withRecovery(ctrl.LegislatorRouter))
	mux.HandleFunc("/mayor", withRecovery(ctrl.MParticipate))
	mux.HandleFunc("/mayor/preview", withRecovery(ctrl.MPreviewLocalForm))
	mux.HandleFunc("/mayor/thank-you", withRecovery(ctrl.MThankYou))

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      logRequest(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Listening on port %s", cfg.AppPort)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func withRecovery(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		h(w, r)
	}
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
