package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Controller struct {
	*Config
}

func NewController(cfg *Config) *Controller {
	return &Controller{
		Config: cfg,
	}
}

func (ctrl Controller) Home() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"BaseURL":        ctrl.AppBaseURL.String(),
			"Areas":          ctrl.Areas,
			"ZoneCandidates": template.JS(ctrl.ZoneCandidates),
		})
	}
}

func (ctrl Controller) SearchZone() gin.HandlerFunc {
	return func(c *gin.Context) {
		qp := RequestQuerySearchZone{}
		if err := c.ShouldBindQuery(&qp); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, RespSearchZone{http.StatusText(http.StatusBadRequest), nil})
			return
		}

		districts, exists := ctrl.AreaFilter[qp.Municipality]
		if !exists {
			c.AbortWithStatusJSON(http.StatusNotFound, RespSearchZone{http.StatusText(http.StatusNotFound), nil})
			return
		}

		if qp.District == nil {
			options := make([]string, len(districts))
			i := 0
			for k := range districts {
				options[i] = k
				i += 1
			}

			if options[0] == "" {
				if zoneCode, exists := districts[""][""]; !exists {
					c.AbortWithStatusJSON(http.StatusNotFound, RespSearchZone{http.StatusText(http.StatusNotFound), nil})
				} else {
					c.JSON(http.StatusOK, RespSearchZone{http.StatusText(http.StatusOK), &ResultSearchZone{[]string{}, zoneCode}})
				}
			} else {
				sort.Slice(options, func(i, j int) bool {
					return options[i] < options[j]
				})

				c.JSON(http.StatusOK, RespSearchZone{http.StatusText(http.StatusOK), &ResultSearchZone{options, ""}})
			}
			return
		}

		wards, exists := districts[*qp.District]
		if !exists {
			c.AbortWithStatusJSON(http.StatusNotFound, RespSearchZone{http.StatusText(http.StatusNotFound), nil})
			return
		}

		if qp.Ward == nil {
			options := make([]string, len(wards))
			i := 0
			for k := range wards {
				options[i] = k
				i += 1
			}

			if options[0] == "" {
				if zoneCode, exists := wards[""]; !exists {
					c.AbortWithStatusJSON(http.StatusNotFound, RespSearchZone{http.StatusText(http.StatusNotFound), nil})
				} else {
					c.JSON(http.StatusOK, RespSearchZone{http.StatusText(http.StatusOK), &ResultSearchZone{[]string{}, zoneCode}})
				}
			} else {
				sort.Slice(options, func(i, j int) bool {
					return options[i] < options[j]
				})

				c.JSON(http.StatusOK, RespSearchZone{http.StatusText(http.StatusOK), &ResultSearchZone{options, ""}})
			}
			return
		}

		zoneCode, exists := wards[*qp.Ward]
		if !exists {
			c.AbortWithStatusJSON(http.StatusNotFound, RespSearchZone{http.StatusText(http.StatusNotFound), nil})
			return
		}

		c.JSON(http.StatusOK, RespSearchZone{http.StatusText(http.StatusOK), &ResultSearchZone{[]string{}, zoneCode}})
	}
}

type RequestQuerySearchZone struct {
	Municipality string  `form:"municipality"`
	District     *string `form:"district"`
	Ward         *string `form:"ward"`
}

type RespSearchZone struct {
	Message string            `json:"message"`
	Result  *ResultSearchZone `json:"result,omitempty"`
}

type ResultSearchZone struct {
	Options  []string `json:"options"`
	ZoneCode string   `json:"zoneCode"`
}

func (ctrl Controller) FillForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		stage := c.Param("stage")
		if stage != "stage-1" && stage != "stage-2" {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		zone := c.Param("zone")
		z := ctrl.GetZone(zone)
		if z == nil {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		address := c.Query("address")
		if address == "" {
			address = z.Address
		}

		twentyYearsAgo := time.Now().AddDate(-20, 0, 0).Format("2006-01-02")

		c.HTML(http.StatusOK, "fill-form.html", gin.H{
			"ZoneCode":         z.ZoneCode,
			"ZoneName":         z.ZoneName,
			"Districts":        z.Districts,
			"CandidateName":    z.CandidateName,
			"Stage":            stage,
			"BaseURL":          ctrl.AppBaseURL.String(),
			"Address":          address,
			"TurnstileSiteKey": ctrl.TurnstileSiteKey,
			"MaxBirthDate":     twentyYearsAgo,
		})
	}
}

func (ctrl Controller) PreviewLocalForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		stage := c.Param("stage")
		if stage != "stage-1" && stage != "stage-2" {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		zone := c.Param("zone")
		z := ctrl.GetZone(zone)
		if z == nil {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		qp := RequestQueryPreview{}
		if err := c.ShouldBindWith(&qp, binding.Form); err != nil {
			c.HTML(http.StatusBadRequest, "4xx.html", GetViewHttpError(http.StatusBadRequest, "您的請求有誤，請回到首頁重新輸入。", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		redirectURL := ctrl.AppBaseURL.JoinPath(stage, zone, "thank-you")
		data, err := qp.ToPreviewData(ctrl.Config, stage, zone, z.GetTopic(), redirectURL.String())
		if err != nil {
			c.HTML(http.StatusBadRequest, "4xx.html", ViewHttp4xxError{
				HttpStatusCode: http.StatusBadRequest,
				ErrorMessage:   err.Error(),
				ReturnURL:      ctrl.AppBaseURL.String(),
			})
			return
		}

		tmpfile := "preview-" + stage + "-" + zone + ".html"
		c.HTML(http.StatusOK, tmpfile, data)
	}
}

type RequestQueryPreview struct {
	Name         string `form:"name" binding:"required"`
	IdNumber     string `form:"id-number" binding:"required"`
	BirthDate    string `form:"birth-date" binding:"required"`
	Address      string `form:"address" binding:"required"`
	MobileNumber string `form:"mobile-number" binidng:"required"`
}

func (r RequestQueryPreview) ToPreviewData(cfg *Config, stage, zone, topic, redirectURL string) (*PreviewData, error) {
	if !isValidIdNumber(r.IdNumber) {
		return nil, fmt.Errorf("身份證輸入錯誤")
	}

	t, err := time.Parse("2006-01-02", r.BirthDate)
	if err != nil {
		return nil, fmt.Errorf("生日輸入錯誤")
	}

	if r.MobileNumber != "" {
		if !isValidMobileNumber(r.MobileNumber) {
			return nil, fmt.Errorf("手機號碼輸入錯誤")
		}
	}

	birthYear, birthMonth, birthDate := t.Date()
	birthYear = birthYear - 1911

	data := &PreviewData{
		BaseURL:      cfg.AppBaseURL.String(),
		Stage:        stage,
		Zone:         zone,
		Topic:        topic,
		Name:         r.Name,
		BirthYear:    birthYear,
		BirthMonth:   int(birthMonth),
		BirthDate:    birthDate,
		Address:      sanitizeAddress(r.Address),
		MobileNumber: r.MobileNumber,
		RedirectURL:  redirectURL,
	}

	for i := 0; i < len(r.IdNumber); i += 1 {
		switch i {
		case 0:
			data.IdNumber.D0 = string(r.IdNumber[i])
		case 1:
			data.IdNumber.D1 = string(r.IdNumber[i])
		case 2:
			data.IdNumber.D2 = string(r.IdNumber[i])
		case 3:
			data.IdNumber.D3 = string(r.IdNumber[i])
		case 4:
			data.IdNumber.D4 = string(r.IdNumber[i])
		case 5:
			data.IdNumber.D5 = string(r.IdNumber[i])
		case 6:
			data.IdNumber.D6 = string(r.IdNumber[i])
		case 7:
			data.IdNumber.D7 = string(r.IdNumber[i])
		case 8:
			data.IdNumber.D8 = string(r.IdNumber[i])
		case 9:
			data.IdNumber.D9 = string(r.IdNumber[i])
		}
	}

	return data, nil
}

type PreviewData struct {
	BaseURL      string
	Stage        string
	Zone         string
	Topic        string
	Name         string
	BirthYear    int
	BirthMonth   int
	BirthDate    int
	Address      string
	IdNumber     IdNumber
	MobileNumber string
	RedirectURL  string
}

type IdNumber struct {
	D0 string
	D1 string
	D2 string
	D3 string
	D4 string
	D5 string
	D6 string
	D7 string
	D8 string
	D9 string
}

func (ctrl Controller) ThankYou() gin.HandlerFunc {
	return func(c *gin.Context) {
		stage := c.Param("stage")
		if stage != "stage-1" && stage != "stage-2" {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		zone := c.Param("zone")
		z := ctrl.GetZone(zone)
		if z == nil {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		topic := z.GetTopic()
		recallFormURL := ctrl.AppBaseURL.JoinPath(stage, zone)

		c.HTML(http.StatusOK, "thank-you.html", gin.H{
			"BaseURL":       ctrl.AppBaseURL.String(),
			"RecallFormURL": recallFormURL.String(),
			"Topic":         topic,
		})
	}
}

func (ctrl Controller) PreviewOriginalLocalForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		stage := c.Param("stage")
		if stage != "stage-1" && stage != "stage-2" {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		zone := c.Param("zone")
		z := ctrl.GetZone(zone)
		if z == nil {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		redirectURL := ctrl.AppBaseURL.JoinPath(stage, zone, "thank-you")

		tmpfile := "preview-" + stage + "-" + zone + ".html"
		c.HTML(http.StatusOK, tmpfile, gin.H{
			"BaseURL":      ctrl.AppBaseURL.String(),
			"Stage":        stage,
			"Zone":         zone,
			"Topic":        z.GetTopic(),
			"RedirectURL":  redirectURL.String(),
			"Name":         "邱吉爾",
			"BirthYear":    63,
			"BirthMonth":   11,
			"BirthDate":    30,
			"Address":      "某某市某某區某某里某某路三段 123 號七樓一段超長的地址一段超長的地址一段超長的地址一段超長的地址一段超長的地址",
			"IdNumber":     IdNumber{"A", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
			"MobileNumber": "0987654321",
		})
	}
}

func (ctrl Controller) VerifyTurnstile() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.PostForm("cf-turnstile-response")
		if token == "" {
			c.HTML(http.StatusBadRequest, "4xx.html", GetViewHttpError(http.StatusBadRequest, "您的請求有誤，請回到首頁重新輸入。", ctrl.AppBaseURL, ctrl.AppBaseURL))
			c.Abort()
			return
		}

		if success, err := ctrl.VerifyTurnstileToken(token); err != nil || !success {
			c.HTML(http.StatusForbidden, "4xx.html", GetViewHttpError(http.StatusForbidden, "驗證失敗，請回到首頁重新輸入", ctrl.AppBaseURL, ctrl.AppBaseURL))
			c.Abort()
			return
		}

		c.Next()
	}
}

func (ctrl Controller) RobotsTxt() gin.HandlerFunc {
	return func(c *gin.Context) {
		tmpl, err := template.ParseFiles("templates/robots.txt")
		if err != nil {
			c.String(http.StatusInternalServerError, "Template Error")
			return
		}

		data := gin.H{
			"BaseURL":       ctrl.AppBaseURL.String(),
			"DisallowPaths": ctrl.DisallowPaths,
		}

		c.Header("Content-Type", "text/plain; charset=utf-8")
		if err := tmpl.Execute(c.Writer, data); err != nil {
			c.String(http.StatusInternalServerError, "Render Error")
		}
	}
}

func (ctrl Controller) Sitemap() gin.HandlerFunc {
	return func(c *gin.Context) {
		urls := []*SitemapURL{
			&SitemapURL{ctrl.AppBaseURL.String(), "2025-02-02", "daily", "1.0"},
		}

		for _, z := range ctrl.Zones {
			if z.Deployed {
				urls = append(urls, &SitemapURL{ctrl.AppBaseURL.JoinPath("stage-1", z.ZoneCode).String(), "2025-02-02", "monthly", "0.8"})
				urls = append(urls, &SitemapURL{ctrl.AppBaseURL.JoinPath("stage-1", z.ZoneCode, "thank-you").String(), "2025-02-05", "weekly", "0.8"})
			}
		}

		sitemap := SitemapURLSet{
			Xmlns:       "http://www.sitemaps.org/schemas/sitemap/0.9",
			SitemapURLs: urls,
		}

		c.Header("Content-Type", "application/xml; charset=utf-8")
		c.XML(http.StatusOK, sitemap)
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

func (ctrl Controller) GetAsset() gin.HandlerFunc {
	return func(c *gin.Context) {
		up := RequestURIAsset{}
		if err := c.ShouldBindUri(&up); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		filePath := fmt.Sprintf("./assets/%s/%s", up.Type, up.File)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		if ctrl.AppEnv == "production" {
			c.Header("Cache-Control", "public, max-age=3600")
		} else {
			c.Header("Cache-Control", "no-cache")
		}
		c.File(filePath)
	}
}

type RequestURIAsset struct {
	Type string `uri:"type" binding:"required"`
	File string `uri:"file" binding:"required"`
}

func (ctrl Controller) NotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL, ctrl.AppBaseURL))
	}
}

func (ctrl Controller) Ping() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "v0.0.1"})
	}
}
