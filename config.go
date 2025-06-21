package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	RecallStatusOngoing = "ONGOING"
	RecallStatusSuccess = "SUCCESS"
	RecallStatusFailed  = "FAILED"
)

const (
	AppEnvProduction = "production"
	AppEnvDev        = "dev"
)

type Config struct {
	AppEnv             string
	AppHostname        string
	AppPath            string
	AppPort            string
	AppTrustedProxies  []string
	AppBaseURL         *url.URL
	TurnstileSiteKey   string
	TurnstileSecretKey string
	DisallowPaths      []string

	RecallTerm uint64
	RecallLegislators
	RecallLegislatorMap map[uint64]RecallLegislators // uint64: ConstituencyId
	Areas
	Municipalities
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		AppEnv:             os.Getenv("APP_ENV"),
		AppHostname:        os.Getenv("APP_HOSTNAME"),
		AppPath:            os.Getenv("APP_PATH"),
		AppPort:            os.Getenv("APP_PORT"),
		AppTrustedProxies:  strings.Split(strings.ReplaceAll(os.Getenv("APP_TRUSTED_PROXIES"), " ", ""), ","),
		TurnstileSiteKey:   os.Getenv("TURNSTILE_SITE_KEY"),
		TurnstileSecretKey: os.Getenv("TURNSTILE_SECRET_KEY"),
		DisallowPaths:      []string{"/health/", "/apis/", "/assets/"},
		RecallTerm:         11,
	}

	if !strings.HasPrefix(cfg.AppPath, "/") {
		cfg.AppPath = "/" + cfg.AppPath
	}

	scheme := "https"
	if strings.HasPrefix(cfg.AppHostname, "localhost") {
		scheme = "http"
	}

	rootPath := ""
	if cfg.AppPath == "/" {
		rootPath = cfg.AppHostname
	} else {
		rootPath = cfg.AppHostname + cfg.AppPath
	}

	var err error

	cfg.AppBaseURL, err = url.ParseRequestURI(scheme + "://" + rootPath)
	if err != nil {
		return nil, err
	}

	cfg.RecallLegislators, cfg.RecallLegislatorMap, err = ReadConfigRecallLegislators(cfg.AppBaseURL)
	if err != nil {
		return nil, err
	}

	cfg.Areas = cfg.RecallLegislators.ToAreas()

	cfg.Municipalities, err = ReadConfigAdministrativeDivisions()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (r Config) GetRecallLegislator(name string) *RecallLegislator {
	for _, row := range r.RecallLegislators {
		if row.PoliticianName == name {
			return row
		}
	}

	return nil
}

func (r Config) HasRecallLegislators(municipalityId uint64, districtId, wardId *uint64) (bool, Divisions, RecallLegislators) {
	if !r.RecallLegislators.HasLegislatorInMunicipality(municipalityId) {
		return false, nil, nil
	}

	municipality := r.Municipalities[municipalityId]
	if districtId == nil {
		return true, municipality.Divisions, nil
	}

	dist := municipality.Districts[*districtId]
	matched := false
	for _, w := range dist.Wards {
		if _, exists := r.RecallLegislatorMap[w.ConstituencyId]; exists {
			matched = true
			break
		}
	}

	if !matched {
		return false, nil, nil
	}

	if wardId == nil {
		return true, dist.Divisions, nil
	}

	constituencyId := dist.Wards[*wardId].ConstituencyId
	if rls, exists := r.RecallLegislatorMap[constituencyId]; exists {
		return true, nil, rls
	}

	return false, nil, nil
}

func (r Config) VerifyTurnstileToken(token string) (bool, error) {
	verifyURL := "https://challenges.cloudflare.com/turnstile/v0/siteverify"

	data := map[string]string{
		"secret":   r.TurnstileSecretKey,
		"response": token,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, err
	}

	resp, err := http.Post(verifyURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	result := TurnstileSiteverifyResponse{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return false, err
	}

	if !result.Success {
		return false, fmt.Errorf("verification failed: %v", result.ErrorCodes)
	}

	return true, nil
}

type TurnstileSiteverifyResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
	Messages   []string `json:"messages"`
}

const (
	JSONConfigRecallLegislators       = "json-config/recall-legislators.json"
	JSONConfigAdministrativeDivisions = "json-config/administrative-divisions.json"
)

var weekdayStrings = [...]string{"日", "一", "二", "三", "四", "五", "六"}

// config: recall-legislator
func ReadConfigRecallLegislators(baseURL *url.URL) (RecallLegislators, map[uint64]RecallLegislators, error) {
	file, err := os.Open(JSONConfigRecallLegislators)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	rows := RecallLegislators{}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&rows); err != nil {
		return nil, nil, err
	}

	rlmap := map[uint64]RecallLegislators{}
	for _, r := range rows {
		r.ParticipateURL = baseURL.JoinPath("legislators", r.PoliticianName)
		r.ParticipateURLString = r.ParticipateURL.String()
		if r.SafetyCutoffDate != nil && *r.SafetyCutoffDate != "" {
			t, err := time.Parse("2006-01-02", *r.SafetyCutoffDate)
			if err != nil {
				return nil, nil, err
			}
			r.SafetyCutoffDateStr = fmt.Sprintf("%d 月 %d 日", t.Month(), t.Day())
		}

		if r.VotingDate != nil && *r.VotingDate != "" {
			t, err := time.Parse("2006-01-02", *r.VotingDate)
			if err != nil {
				return nil, nil, err
			}
			r.VotingDateWeekdayStr = fmt.Sprintf("%d 月 %d 日 (%s)", t.Month(), t.Day(), weekdayStrings[t.Weekday()])
			r.VotingDateStr = fmt.Sprintf("%d 月 %d 日", t.Month(), t.Day())
		}

		if _, exists := rlmap[r.ConstituencyId]; !exists {
			rlmap[r.ConstituencyId] = RecallLegislators{}
		}

		rlmap[r.ConstituencyId] = append(rlmap[r.ConstituencyId], r)
	}

	return rows, rlmap, nil
}

type RecallLegislators []*RecallLegislator

func (rs RecallLegislators) HasLegislatorInMunicipality(municipalityId uint64) bool {
	for _, r := range rs {
		if r.MunicipalityId == municipalityId {
			return true
		}
	}

	return false
}

func (rs *RecallLegislators) CalcDaysLeft(now time.Time) {
	for _, r := range *rs {
		r.CalcDaysLeft(now)
	}
}

type RecallLegislator struct {
	ConstituencyId        uint64   `json:"constituencyId"`
	MunicipalityId        uint64   `json:"municipalityId"`
	Term                  uint64   `json:"term"`
	MunicipalityName      string   `json:"municipalityName"`
	ConstituencyNum       uint64   `json:"constituencyNum"`
	PoliticianName        string   `json:"politicianName"`
	RecallStage           uint64   `json:"recallStage"`
	RecallStatus          string   `json:"recallStatus"`
	FormDeployed          bool     `json:"formDeployed"`
	CsoURL                string   `json:"csoURL"`
	CalendarURL           string   `json:"calendarURL"`
	HasCalendarMaintainer bool     `json:"hasCalendarMaintainer"`
	VotingDate            *string  `json:"votingDate"`
	VotingEventURL        *string  `json:"votingEventURL"`
	VotingDateStr         string   `json:"votingDateStr"`
	VotingDateWeekdayStr  string   `json:"votingDateWeekdayStr"`
	ByElectionDate        *string  `json:"byElectionDate"`
	ByElectionEventURL    *string  `json:"byElectionEventURL"`
	SafetyCutoffDate      *string  `json:"safetyCutoffDate"`
	EndedDate             *string  `json:"endedDate"`
	IsShortage            bool     `json:"isShortage"`
	ConstituencyName      string   `json:"constituencyName"`
	ParticipateURL        *url.URL `json:"-"`
	ParticipateURLString  string   `json:"participateURL"`
	DaysLeft              int      `json:"daysLeft"`
	SafetyCutoffDateStr   string   `json:"safetyCutoffDateStr"`
}

func (r *RecallLegislator) CalcDaysLeft(now time.Time) {
	if r.EndedDate == nil || *r.EndedDate == "" {
		r.DaysLeft = 0
		return
	}

	cutoff, err := time.Parse("2006-01-02", *r.EndedDate)
	if err != nil {
		r.DaysLeft = 0
		return
	}

	r.DaysLeft = int(cutoff.Sub(now).Hours() / 24)
}

func (r RecallLegislator) IsPetitioning() bool {
	if r.RecallStage == 1 || r.RecallStage == 2 {
		return true
	}

	return false
}

func (r RecallLegislator) GetTmplFilename() string {
	switch r.RecallStage {
	case 1, 2:
		return fmt.Sprintf("stage-%d-%s.html", r.RecallStage, r.PoliticianName)
	}

	return ""
}

func (rs RecallLegislators) ToAreas() Areas {
	areas := Areas{}
	for _, r := range rs {
		matched := false
		for _, a := range areas {
			if a.MunicipalityId == r.MunicipalityId {
				matched = true
				a.RecallLegislators = append(a.RecallLegislators, r)
				break
			}
		}

		if !matched {
			areas = append(areas, &Area{r.MunicipalityId, &r.MunicipalityName, RecallLegislators{r}})
		}
	}

	return areas
}

type Areas []*Area

type Area struct {
	MunicipalityId    uint64
	MunicipalityName  *string
	RecallLegislators RecallLegislators
}

// config: administrative-divisions
func ReadConfigAdministrativeDivisions() (Municipalities, error) {
	file, err := os.Open(JSONConfigAdministrativeDivisions)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	rows := Municipalities{}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&rows); err != nil {
		return nil, err
	}

	for _, m := range rows {
		m.Divisions = Divisions{}
		for _, d := range m.Districts {
			d.Divisions = Divisions{}
			for _, w := range d.Wards {
				d.Divisions = append(d.Divisions, w.Division)
			}

			sort.Slice(d.Divisions, func(i, j int) bool {
				return d.Divisions[i].Name < d.Divisions[j].Name
			})

			m.Divisions = append(m.Divisions, d.Division)
		}

		sort.Slice(m.Divisions, func(i, j int) bool {
			return m.Divisions[i].Name < m.Divisions[j].Name
		})
	}

	return rows, nil
}

type Municipalities []*Municipality

type Division struct {
	Id   uint64 `json:"id"`
	Name string `json:"n"`
}

type Divisions []*Division

type Municipality struct {
	*Division
	Districts map[uint64]*District `json:"ds"`
	Divisions `json:"-"`
}

type District struct {
	*Division
	Wards     map[uint64]*Ward `json:"ws"`
	Divisions `json:"-"`
}

type Ward struct {
	*Division
	ConstituencyId uint64 `json:"cid"`
}
