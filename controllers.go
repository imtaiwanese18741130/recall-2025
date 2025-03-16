package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	MayorName = "高虹安"
	MayorCity = "新竹市"
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
			"Municipalities": ctrl.Municipalities,
			"Areas":          ctrl.Areas,
		})
	}
}

func (ctrl Controller) AuthorizationLetter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "authorization-letter.html", gin.H{
			"BaseURL": ctrl.AppBaseURL.String(),
		})
	}
}

func (ctrl Controller) SearchRecallConstituency() gin.HandlerFunc {
	return func(c *gin.Context) {
		qp := RequestQuerySearchRecallConstituency{}
		if err := c.ShouldBindQuery(&qp); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, RespSearchRecallConstituency{http.StatusText(http.StatusBadRequest), nil})
			return
		}

		exists, divisions, legislators := ctrl.HasRecallLegislators(qp.MunicipalityId, qp.DistrictId, qp.WardId)
		if !exists {
			c.AbortWithStatusJSON(http.StatusNotFound, RespSearchRecallConstituency{
				Message: http.StatusText(http.StatusNotFound),
				Result:  nil,
			})
			return
		}

		c.JSON(http.StatusOK, RespSearchRecallConstituency{
			Message: http.StatusText(http.StatusOK),
			Result: &ResultSearchRecallConstituency{
				Divisions:   divisions,
				Legislators: legislators,
			},
		})
	}
}

type RequestQuerySearchRecallConstituency struct {
	MunicipalityId uint64  `form:"municipality" binding:"required,numeric"`
	DistrictId     *uint64 `form:"district" binding:"omitempty,numeric"`
	WardId         *uint64 `form:"ward" binding:"omitempty,numeric"`
}

type RespSearchRecallConstituency struct {
	Message string                          `json:"message"`
	Result  *ResultSearchRecallConstituency `json:"result,omitempty"`
}

type ResultSearchRecallConstituency struct {
	Divisions   Divisions         `json:"divisions,omitempty"`
	Legislators RecallLegislators `json:"legislators,omitempty"`
}

func (ctrl Controller) Participate() gin.HandlerFunc {
	return func(c *gin.Context) {
		up := RequestUriStageLegislator{}
		if err := c.ShouldBindUri(&up); err != nil {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "您請求的頁面不存在", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		l := ctrl.GetRecallLegislator(up.Name)
		if l == nil {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "您請求的頁面不存在", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		if l.RecallStatus != RecallStatusOngoing {
			c.Redirect(http.StatusMovedPermanently, ctrl.AppBaseURL.String())
			return
		}

		address := c.Query("address")
		if address == "" {
			address = l.MunicipalityName
		}

		previewURL := l.ParticipateURL.JoinPath("preview")

		switch l.RecallStage {
		case 1, 2:
			c.HTML(http.StatusOK, "fill-form.html", gin.H{
				"BaseURL":          ctrl.AppBaseURL.String(),
				"PreviewURL":       previewURL.String(),
				"Address":          address,
				"TurnstileSiteKey": ctrl.TurnstileSiteKey,
				"Legislator":       l,
			})
		case 3, 4:
			c.HTML(http.StatusOK, "vote-reminder.html", gin.H{
				"BaseURL":    ctrl.AppBaseURL.String(),
				"Legislator": l,
			})
		default:
			c.Redirect(http.StatusMovedPermanently, ctrl.AppBaseURL.String())
		}
	}
}

type RequestUriStageLegislator struct {
	Name  string `uri:"name" binding:"required"`
	Stage uint64 `uri:"stage" binding:"omitempty,numeric,oneof=1 2 3 4"`
}

func (ctrl Controller) PreviewLocalForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		up := RequestUriStageLegislator{}
		if err := c.ShouldBindUri(&up); err != nil {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "您請求的頁面不存在", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		l := ctrl.GetRecallLegislator(up.Name)
		if l == nil {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "您請求的頁面不存在", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		if l.RecallStatus != RecallStatusOngoing {
			c.Redirect(http.StatusMovedPermanently, ctrl.AppBaseURL.String())
			return
		}

		if !l.IsPetitioning() {
			c.HTML(http.StatusConflict, "4xx.html", GetViewHttpError(http.StatusConflict, "候選人不處於連署階段", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		up.Stage = l.RecallStage

		qp := RequestQueryPreview{}
		if err := c.ShouldBindWith(&qp, binding.Form); err != nil {
			c.HTML(http.StatusBadRequest, "4xx.html", GetViewHttpError(http.StatusBadRequest, "您的請求有誤，請回到首頁重新輸入。", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		data, err := qp.ToPreviewData(ctrl.Config, &up, l)
		if err != nil {
			c.HTML(http.StatusBadRequest, "4xx.html", ViewHttp4xxError{
				HttpStatusCode: http.StatusBadRequest,
				ErrorMessage:   err.Error(),
				ReturnURL:      ctrl.AppBaseURL.String(),
			})
			return
		}

		tmpfile := l.GetTmplFilename()
		c.HTML(http.StatusOK, tmpfile, data)
	}
}

type RequestQueryPreview struct {
	Name         string `form:"name" binding:"required"`
	IdNumber     string `form:"id-number" binding:"required"`
	BirthYear    int    `form:"birth-year" binding:"required"`
	BirthMonth   int    `form:"birth-month" binding:"required"`
	BirthDay     int    `form:"birth-day" binding:"required"`
	Address      string `form:"address" binding:"required"`
	MobileNumber string `form:"mobile-number" binidng:"required"`
}

func (r RequestQueryPreview) ToPreviewData(cfg *Config, up *RequestUriStageLegislator, l *RecallLegislator) (*PreviewData, error) {
	if !isValidIdNumber(r.IdNumber) {
		return nil, fmt.Errorf("身份證輸入錯誤")
	}

	if l.RecallStage == 1 {
		if r.MobileNumber != "" {
			if !isValidMobileNumber(r.MobileNumber) {
				return nil, fmt.Errorf("手機號碼輸入錯誤")
			}
		}
	}

	stage := strconv.FormatUint(up.Stage, 10)
	redirectURL := l.ParticipateURL.JoinPath("thank-you")
	imagePrefix := fmt.Sprintf("stage-%s-%s", stage, up.Name)

	data := &PreviewData{
		BaseURL:          cfg.AppBaseURL.String(),
		ParticipateURL:   l.ParticipateURL,
		RedirectURL:      redirectURL.String(),
		PoliticianName:   up.Name,
		ConstituencyName: l.ConstituencyName,
		RecallStage:      up.Stage,
		ImagePrefix:      imagePrefix,
		Name:             r.Name,
		BirthYear:        r.BirthYear,
		BirthMonth:       r.BirthMonth,
		BirthDate:        r.BirthDay,
		MobileNumber:     r.MobileNumber,
		Address:          sanitizeAddress(r.Address),
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
	BaseURL          string
	ParticipateURL   *url.URL
	RedirectURL      string
	PoliticianName   string
	ConstituencyName string
	RecallStage      uint64
	ImagePrefix      string
	Name             string
	IdNumber         IdNumber
	BirthYear        int
	BirthMonth       int
	BirthDate        int
	MobileNumber     string
	Address          string
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
		up := RequestUriStageLegislator{}
		if err := c.ShouldBindUri(&up); err != nil {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "您請求的頁面不存在", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		l := ctrl.GetRecallLegislator(up.Name)
		if l == nil {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "您請求的頁面不存在", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		if l.RecallStatus != RecallStatusOngoing {
			c.Redirect(http.StatusMovedPermanently, ctrl.AppBaseURL.String())
			return
		}

		c.HTML(http.StatusOK, "thank-you.html", gin.H{
			"BaseURL":        ctrl.AppBaseURL.String(),
			"ParticipateURL": l.ParticipateURL,
			"CalendarURL":    l.CalendarURL,
			"CsoURL":         l.CsoURL,
		})
	}
}

func (ctrl Controller) PreviewOriginalLocalForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		up := RequestUriStageLegislator{}
		if err := c.ShouldBindUri(&up); err != nil {
			c.Redirect(http.StatusMovedPermanently, ctrl.AppBaseURL.String())
			return
		}

		var data *PreviewData
		if up.Name != MayorName {
			l := ctrl.GetRecallLegislator(up.Name)
			if l == nil {
				c.Redirect(http.StatusMovedPermanently, ctrl.AppBaseURL.String())
				return
			}

			data = &PreviewData{
				BaseURL:          ctrl.AppBaseURL.String(),
				ParticipateURL:   l.ParticipateURL,
				RedirectURL:      l.ParticipateURL.JoinPath("thank-you").String(),
				PoliticianName:   up.Name,
				ConstituencyName: l.ConstituencyName,
				RecallStage:      up.Stage,
				Name:             "邱吉爾",
				IdNumber:         IdNumber{"A", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
				BirthYear:        888,
				BirthMonth:       11,
				BirthDate:        30,
				MobileNumber:     "0987654321",
				Address:          "某某市某某區某某里某某路三段 123 號七樓一段超長的地址一段超長的地址一段超長的地址一段超長的地址一段超長的地址一段超長的地址",
			}
		} else {
			participateURL := ctrl.AppBaseURL.JoinPath("mayor")
			data = &PreviewData{
				BaseURL:          ctrl.AppBaseURL.String(),
				ParticipateURL:   participateURL,
				RedirectURL:      participateURL.JoinPath("thank-you").String(),
				PoliticianName:   up.Name,
				ConstituencyName: MayorCity,
				RecallStage:      up.Stage,
				Name:             "邱吉爾",
				IdNumber:         IdNumber{"A", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
				BirthYear:        888,
				BirthMonth:       11,
				BirthDate:        30,
				MobileNumber:     "0987654321",
				Address:          "某某市某某區某某里某某路三段 123 號七樓一段超長的地址一段超長的地址一段超長的地址一段超長的地址一段超長的地址",
			}
		}

		switch up.Stage {
		case 2:
			tmpfile := fmt.Sprintf("stage-2-%s.html", up.Name)
			c.HTML(http.StatusOK, tmpfile, data)
			return
		default:
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "您請求的頁面不存在", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}
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

		c.Header("Content-Type", "text/plain; charset=utf-9")
		if err := tmpl.Execute(c.Writer, data); err != nil {
			c.String(http.StatusInternalServerError, "Render Error")
		}
	}
}

func (ctrl Controller) Sitemap() gin.HandlerFunc {
	return func(c *gin.Context) {
		date := "2025-03-02"
		urls := []*SitemapURL{
			&SitemapURL{ctrl.AppBaseURL.String(), date, "daily", "1.0"},
			&SitemapURL{ctrl.AppBaseURL.JoinPath("authorization-letter").String(), "2025-02-26", "yearly", "1.0"},
			&SitemapURL{ctrl.AppBaseURL.JoinPath("mayor").String(), "2025-03-12", "weekly", "0.9"},
			&SitemapURL{ctrl.AppBaseURL.JoinPath("mayor", "thank-you").String(), "2025-03-12", "weekly", "0.9"},
		}

		for _, l := range ctrl.RecallLegislators {
			legislatorURL := l.ParticipateURL
			if l.RecallStatus == "ONGOING" {
				urls = append(urls, &SitemapURL{legislatorURL.String(), date, "weekly", "0.9"})
				urls = append(urls, &SitemapURL{legislatorURL.JoinPath("thank-you").String(), date, "weekly", "0.8"})
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
			//if up.Type == "images" && strings.HasPrefix(up.File, "stage-2-") {
			//	c.Header("Cache-Control", "no-cache")
			//} else {
			//	c.Header("Cache-Control", "public, max-age=3600")
			//}
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
		c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "您請求的頁面不存在", ctrl.AppBaseURL, ctrl.AppBaseURL))
	}
}

func (ctrl Controller) Ping() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "v0.0.1"})
	}
}

// For mayor
func (ctrl Controller) MParticipate() gin.HandlerFunc {
	return func(c *gin.Context) {
		previewURL := ctrl.AppBaseURL.JoinPath("mayor", "preview")
		address := MayorCity

		c.HTML(http.StatusOK, "mayor-fill-form.html", gin.H{
			"BaseURL":          ctrl.AppBaseURL.String(),
			"PreviewURL":       previewURL.String(),
			"Address":          address,
			"TurnstileSiteKey": ctrl.TurnstileSiteKey,
		})
	}
}

func (ctrl Controller) MPreviewLocalForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		qp := RequestQueryPreview{}
		if err := c.ShouldBindWith(&qp, binding.Form); err != nil {
			c.HTML(http.StatusBadRequest, "4xx.html", GetViewHttpError(http.StatusBadRequest, "您的請求有誤，請回到首頁重新輸入。", ctrl.AppBaseURL, ctrl.AppBaseURL))
			return
		}

		data, err := qp.ToMayorPreviewData(ctrl.Config)
		if err != nil {
			c.HTML(http.StatusBadRequest, "4xx.html", ViewHttp4xxError{
				HttpStatusCode: http.StatusBadRequest,
				ErrorMessage:   err.Error(),
				ReturnURL:      ctrl.AppBaseURL.String(),
			})
			return
		}

		c.HTML(http.StatusOK, "stage-2-"+MayorName+".html", data)
	}
}

func (r RequestQueryPreview) ToMayorPreviewData(cfg *Config) (*PreviewData, error) {
	if !isValidIdNumber(r.IdNumber) {
		return nil, fmt.Errorf("身份證輸入錯誤")
	}

	redirectURL := cfg.AppBaseURL.JoinPath("mayor", "thank-you")
	imagePrefix := "stage-2-" + MayorName

	data := &PreviewData{
		BaseURL:          cfg.AppBaseURL.String(),
		ParticipateURL:   cfg.AppBaseURL.JoinPath("mayor"),
		RedirectURL:      redirectURL.String(),
		PoliticianName:   MayorName,
		ConstituencyName: MayorCity,
		RecallStage:      2,
		ImagePrefix:      imagePrefix,
		Name:             r.Name,
		BirthYear:        r.BirthYear,
		BirthMonth:       r.BirthMonth,
		BirthDate:        r.BirthDay,
		Address:          sanitizeAddress(r.Address),
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

func (ctrl Controller) MThankYou() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "mayor-thank-you.html", gin.H{
			"BaseURL":        ctrl.AppBaseURL,
			"ParticipateURL": ctrl.AppBaseURL.JoinPath("mayor"),
			"CsoURL":         "https://www.facebook.com/hc.thebigrecall",
		})
	}
}
