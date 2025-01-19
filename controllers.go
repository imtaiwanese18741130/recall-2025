package main

import (
	"fmt"
	"net/http"
	"os"
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
			"BaseURL": ctrl.AppBaseURL.String(),
		})
	}
}

func (ctrl Controller) FillForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		stage := c.Param("stage")
		if stage != "stage-1" && stage != "stage-2" {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL))
			return
		}

		zone := c.Param("zone")
		topic, addressPrefix := ctrl.GetZoneTopic(zone)
		if topic == "" {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL))
			return
		}

		c.HTML(http.StatusOK, "fill-form.html", gin.H{
			"Topic":            topic,
			"Zone":             zone,
			"Stage":            stage,
			"BaseURL":          ctrl.AppBaseURL.String(),
			"AddressPrefix":    addressPrefix,
			"TurnstileSiteKey": ctrl.TurnstileSiteKey,
		})
	}
}

func (ctrl Controller) PreviewLocalForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		stage := c.Param("stage")
		if stage != "stage-1" && stage != "stage-2" {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL))
			return
		}

		zone := c.Param("zone")
		topic, _ := ctrl.GetZoneTopic(zone)
		if topic == "" {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL))
			return
		}

		qp := RequestQueryPreview{}
		if err := c.ShouldBindWith(&qp, binding.Form); err != nil {
			fmt.Println(err)
			c.HTML(http.StatusBadRequest, "4xx.html", GetViewHttpError(http.StatusBadRequest, "您的請求有誤，請回到首頁重新輸入。", ctrl.AppBaseURL))
			return
		}

		redirectURL := ctrl.AppBaseURL.JoinPath("thank-you")
		query := redirectURL.Query()
		query.Add("stage", stage)
		query.Add("zone", zone)
		redirectURL.RawQuery = query.Encode()

		data, err := qp.ToPreviewData(ctrl.Config, topic, redirectURL.String())
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
	MobileNumber string `form:"mobile-number" binidng:"omitempty"`
}

func (r RequestQueryPreview) ToPreviewData(cfg *Config, topic, redirectURL string) (*PreviewData, error) {
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
	BaseURL    string
	Topic      string
	Name       string
	BirthYear  int
	BirthMonth int
	BirthDate  int
	Address    string
	IdNumber   struct {
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
	MobileNumber string
	RedirectURL  string
}

func (ctrl Controller) PreviewOriginalLocalForm() gin.HandlerFunc {
	return func(c *gin.Context) {
		stage := c.Param("stage")
		if stage != "stage-1" && stage != "stage-2" {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL))
			return
		}

		zone := c.Param("zone")
		if !ctrl.HasZone(zone) {
			c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL))
			return
		}

		tmpfile := "preview-" + stage + "-" + zone + ".html"
		c.HTML(http.StatusOK, tmpfile, gin.H{"BaseURL": ctrl.AppBaseURL.String()})
	}
}

func (ctrl Controller) VerifyTurnstile() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.PostForm("cf-turnstile-response")
		if token == "" {
			c.HTML(http.StatusBadRequest, "4xx.html", GetViewHttpError(http.StatusBadRequest, "您的請求有誤，請回到首頁重新輸入。", ctrl.AppBaseURL))
			c.Abort()
			return
		}

		if success, err := ctrl.VerifyTurnstileToken(token); err != nil || !success {
			c.HTML(http.StatusForbidden, "4xx.html", GetViewHttpError(http.StatusForbidden, "驗證失敗，請回到首頁重新輸入", ctrl.AppBaseURL))
			c.Abort()
			return
		}

		c.Next()
	}
}

func (ctrl Controller) ThankYou() gin.HandlerFunc {
	return func(c *gin.Context) {
		recallFormURL := ctrl.AppBaseURL
		stage := c.Query("stage")
		if stage == "stage-1" || stage == "stage-2" {
			recallFormURL = recallFormURL.JoinPath(stage)
		} else {
			recallFormURL = nil
		}

		topic := ""
		zone := c.Query("zone")
		if ctrl.HasZone(zone) {
			recallFormURL = recallFormURL.JoinPath(zone)
			topic, _ = ctrl.GetZoneTopic(zone)
		} else {
			recallFormURL = nil
		}

		c.HTML(http.StatusOK, "thank-you.html", gin.H{
			"BaseURL":       ctrl.AppBaseURL.String(),
			"RecallFormURL": recallFormURL,
			"Topic":         topic,
		})
	}
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

		c.Header("Cache-Control", "public, max-age=3600")
		c.File(filePath)
	}
}

type RequestURIAsset struct {
	Type string `uri:"type" binding:"required"`
	File string `uri:"file" binding:"required"`
}

func (ctrl Controller) NotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "4xx.html", GetViewHttpError(http.StatusNotFound, "抱歉，我們無法找到您要的頁面。", ctrl.AppBaseURL))
	}
}

func (ctrl Controller) Ping() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "v0.0.1"})
	}
}