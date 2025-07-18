package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
)

type Controller struct {
	*Config
	Templates *template.Template
}

func NewController(cfg *Config, tmpl *template.Template) *Controller {
	return &Controller{
		Config:    cfg,
		Templates: tmpl,
	}
}

func (ctrl *Controller) Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			ctrl.renderTemplate(w, "home.html", map[string]interface{}{
				"BaseURL":        ctrl.AppBaseURL.String(),
				"Municipalities": ctrl.Municipalities,
				"Areas":          ctrl.Areas,
			})
		} else {
			ctrl.renderTemplate(w, "4xx.html", GetViewHttpError(http.StatusNotFound, "您請求的頁面不存在", ctrl.AppBaseURL, ctrl.AppBaseURL))
		}
	}
}

func (ctrl *Controller) AuthorizationLetter() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctrl.renderTemplate(w, "authorization-letter.html", map[string]interface{}{
			"BaseURL": ctrl.AppBaseURL.String(),
		})
	}
}

func (ctrl *Controller) ListMunicipalities() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		divisions := make(Divisions, 0)
		for _, m := range ctrl.Municipalities {
			if m.Division.Id == 0 {
				continue
			}
			divisions = append(divisions, m.Division)
		}

		writeJSON(w, http.StatusOK, RespListMunicipalities{
			Message: http.StatusText(http.StatusOK),
			Result: &ResultListMunicipalities{
				Divisions: divisions,
			},
		})
	}
}

func (ctrl *Controller) SearchRecallConstituency() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		var qp RequestQuerySearchRecallConstituency

		m := r.FormValue("municipality")
		mid, err := strconv.ParseUint(m, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "municipality error"})
			return
		}
		qp.MunicipalityId = mid

		if d := r.FormValue("district"); d != "" {
			val, err := strconv.ParseUint(d, 10, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"message": "district error"})
				return
			}
			qp.DistrictId = &val
		}

		if wd := r.FormValue("ward"); wd != "" {
			val, err := strconv.ParseUint(wd, 10, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"message": "ward error"})
				return
			}
			qp.WardId = &val
		}

		exists, divisions, legislators := ctrl.HasRecallLegislators(qp.MunicipalityId, qp.DistrictId, qp.WardId)
		if !exists {
			writeJSON(w, http.StatusNotFound, RespSearchRecallConstituency{
				Message: http.StatusText(http.StatusNotFound),
			})
			return
		}

		writeJSON(w, http.StatusOK, RespSearchRecallConstituency{
			Message: http.StatusText(http.StatusOK),
			Result: &ResultSearchRecallConstituency{
				Divisions:   divisions,
				Legislators: legislators,
			},
		})
	}
}

func (ctrl *Controller) Participate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := ctrl.GetRecallLegislator(r.PathValue("name"))
		if l == nil || l.RecallStatus != RecallStatusOngoing {
			http.Redirect(w, r, ctrl.AppBaseURL.String(), http.StatusMovedPermanently)
			return
		}

		address := r.URL.Query().Get("address")
		if address == "" {
			address = l.MunicipalityName
		}

		switch l.RecallStage {
		case 1, 2:
			ctrl.renderTemplate(w, "fill-form.html", map[string]interface{}{
				"BaseURL":          ctrl.AppBaseURL.String(),
				"PreviewURL":       l.ParticipateURL.JoinPath("preview").String(),
				"Address":          address,
				"TurnstileSiteKey": ctrl.TurnstileSiteKey,
				"Legislator":       l,
			})
		case 3, 4:
			ctrl.renderTemplate(w, "vote-reminder.html", map[string]interface{}{
				"BaseURL":    ctrl.AppBaseURL.String(),
				"Legislator": l,
			})
		default:
			http.Redirect(w, r, ctrl.AppBaseURL.String(), http.StatusMovedPermanently)
		}
	}
}

func (ctrl *Controller) ThankYou() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := ctrl.GetRecallLegislator(r.PathValue("name"))
		if l == nil || l.RecallStatus != RecallStatusOngoing {
			http.Redirect(w, r, ctrl.AppBaseURL.String(), http.StatusMovedPermanently)
			return
		}

		ctrl.renderTemplate(w, "thank-you.html", map[string]interface{}{
			"BaseURL":        ctrl.AppBaseURL.String(),
			"ParticipateURL": l.ParticipateURL,
			"CsoURL":         l.CsoURL,
		})
	}
}

func (ctrl *Controller) VerifyTurnstile(w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()
	token := r.FormValue("cf-turnstile-response")
	if token == "" {
		ctrl.renderTemplate(w, "4xx.html", GetViewHttpError(http.StatusBadRequest, "您的請求有誤，請回到首頁重新輸入。", ctrl.AppBaseURL, ctrl.AppBaseURL))
		return false
	}
	success, err := ctrl.VerifyTurnstileToken(token)
	if err != nil || !success {
		ctrl.renderTemplate(w, "4xx.html", GetViewHttpError(http.StatusForbidden, "驗證失敗，請回到首頁重新輸入", ctrl.AppBaseURL, ctrl.AppBaseURL))
		return false
	}
	return true
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (ctrl *Controller) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	if ctrl.AppEnv == AppEnvProduction {
		if err := ctrl.Templates.ExecuteTemplate(w, name, data); err != nil {
			http.Error(w, "Template rendering error", http.StatusInternalServerError)
		}
	} else {
		if t, err := template.ParseFiles("templates/tmpl.html", "templates/"+name); err != nil {
			http.Error(w, fmt.Errorf("Template parsing error: %v", err).Error(), http.StatusInternalServerError)
		} else if err := t.ExecuteTemplate(w, name, data); err != nil {
			http.Error(w, fmt.Errorf("Template rendering error: %v", err).Error(), http.StatusInternalServerError)
		}
	}
}

type SitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq"`
	Priority   string `xml:"priority"`
}

type SitemapURLSet struct {
	XMLName     xml.Name      `xml:"urlset"`
	Xmlns       string        `xml:"xmlns,attr"`
	SitemapURLs []*SitemapURL `xml:"url"`
}

func (ctrl *Controller) Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "v0.2"})
	}
}

func (ctrl *Controller) RobotsTxt() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/robots.txt")
		if err != nil {
			http.Error(w, "Template Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		tmpl.Execute(w, map[string]interface{}{
			"BaseURL":       ctrl.AppBaseURL.String(),
			"DisallowPaths": ctrl.DisallowPaths,
		})
	}
}

func (ctrl *Controller) Sitemap() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		date := "2025-03-02"
		urls := []*SitemapURL{
			{ctrl.AppBaseURL.String(), date, "daily", "1.0"},
			{ctrl.AppBaseURL.JoinPath("authorization-letter").String(), "2025-02-26", "yearly", "1.0"},
		}

		for _, l := range ctrl.RecallLegislators {
			if l.RecallStatus == "ONGOING" {
				urls = append(urls,
					&SitemapURL{l.ParticipateURL.String(), date, "weekly", "0.9"},
					&SitemapURL{l.ParticipateURL.JoinPath("thank-you").String(), date, "weekly", "0.8"},
				)
			}
		}

		sitemap := SitemapURLSet{
			Xmlns:       "http://www.sitemaps.org/schemas/sitemap/0.9",
			SitemapURLs: urls,
		}

		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		xml.NewEncoder(w).Encode(sitemap)
	}
}

func (ctrl *Controller) GetAsset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filePath := path.Join("assets", r.PathValue("filetype"), r.PathValue("filename"))
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		if ctrl.AppEnv == AppEnvProduction {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}

		http.ServeFile(w, r, filePath)
	}
}

func (ctrl *Controller) Preview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !ctrl.VerifyTurnstile(w, r) {
			ctrl.renderTemplate(w, "4xx.html", GetViewHttpError(http.StatusBadRequest, "不合法的請求", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		name := r.PathValue("name")
		l := ctrl.GetRecallLegislator(name)
		if l == nil || l.RecallStatus != RecallStatusOngoing || !l.IsPetitioning() {
			ctrl.renderTemplate(w, "4xx.html", GetViewHttpError(http.StatusConflict, "候選人不處於連署階段", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		up := RequestUriStageLegislator{Name: name, Stage: l.RecallStage}
		stage := strconv.FormatUint(up.Stage, 10)

		data := &PreviewData{
			BaseURL:          ctrl.AppBaseURL.String(),
			ParticipateURL:   l.ParticipateURL,
			RedirectURL:      l.ParticipateURL.JoinPath("thank-you").String(),
			PoliticianName:   up.Name,
			ConstituencyName: l.ConstituencyName,
			RecallStage:      up.Stage,
			ImagePrefix:      fmt.Sprintf("stage-%s-%s", stage, up.Name),
		}

		tmpfile := l.GetTmplFilename()
		ctrl.renderTemplate(w, tmpfile, data)
	}
}

func (ctrl *Controller) NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctrl.renderTemplate(w, "4xx.html", GetViewHttpError(http.StatusNotFound, "您請求的頁面不存在", ctrl.AppBaseURL, ctrl.AppBaseURL))
	}
}

type RequestQuerySearchRecallConstituency struct {
	MunicipalityId uint64
	DistrictId     *uint64
	WardId         *uint64
}

type RespSearchRecallConstituency struct {
	Message string                          `json:"message"`
	Result  *ResultSearchRecallConstituency `json:"result,omitempty"`
}

type ResultSearchRecallConstituency struct {
	Divisions   Divisions         `json:"divisions,omitempty"`
	Legislators RecallLegislators `json:"legislators,omitempty"`
}

type RespListMunicipalities struct {
	Message string                    `json:"message"`
	Result  *ResultListMunicipalities `json:"result,omitempty"`
}

type ResultListMunicipalities struct {
	Divisions Divisions `json:"divisions,omitempty"`
}
type RequestUriStageLegislator struct {
	Name  string
	Stage uint64
}

func (ctrl *Controller) PreviewOriginalLocalForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")

		l := ctrl.GetRecallLegislator(name)
		if l == nil {
			http.Redirect(w, r, ctrl.AppBaseURL.String(), http.StatusMovedPermanently)
			return
		}

		data := &PreviewData{
			BaseURL:          ctrl.AppBaseURL.String(),
			ParticipateURL:   l.ParticipateURL,
			RedirectURL:      l.ParticipateURL.JoinPath("thank-you").String(),
			PoliticianName:   name,
			ConstituencyName: l.ConstituencyName,
			RecallStage:      2,
		}

		tmpl := fmt.Sprintf("stage-2-%s.html", name)
		ctrl.renderTemplate(w, tmpl, data)
	}
}

type PreviewData struct {
	BaseURL          string
	ParticipateURL   *url.URL
	RedirectURL      string
	PoliticianName   string
	ConstituencyName string
	RecallStage      uint64
	ImagePrefix      string
}
