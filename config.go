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
)

const (
	AreaNameKeelung       = "基隆市"
	AreaNameTaipei        = "臺北市"
	AreaNameNewTaipei     = "新北市"
	AreaNameTaoyuan       = "桃園市"
	AreaNameHsinchuCity   = "新竹市"
	AreaNameHsinchuCounty = "新竹縣"
	AreaNameMiaoli        = "苗栗縣"
	AreaNameTaichung      = "臺中市"
	AreaNameChanghua      = "彰化縣"
	AreaNameNantou        = "南投縣"
	AreaNameYunlin        = "雲林縣"
	AreaNameHualien       = "花蓮縣"
	AreaNameTaitung       = "臺東縣"
	AreaNameKinmen        = "金門縣"
	AreaNameLienchiang    = "連江縣"
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
	Zones              map[string]*Zone
	Areas              []*Area
	AreaFilter         map[string]map[string]map[string]string
	ZoneCandidates     []byte
	DisallowPaths      []string
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
		DisallowPaths:      []string{"/health/", "/filter", "/assets/"},
	}

	var baseURL *url.URL
	if !strings.HasPrefix(cfg.AppPath, "/") {
		cfg.AppPath = "/" + cfg.AppPath
	}

	scheme := "https"
	if cfg.AppHostname == "localhost" {
		scheme = "http"
	}

	rootPath := ""
	if cfg.AppPath == "/" {
		rootPath = cfg.AppHostname
	} else {
		rootPath = cfg.AppHostname + cfg.AppPath
	}

	baseURL, err := url.ParseRequestURI(scheme + "://" + rootPath)
	if err != nil {
		return nil, err
	}

	cfg.AppBaseURL = baseURL
	cfg.Zones = map[string]*Zone{
		"keelung-1": &Zone{"keelung-1", ZoneCandidate{"基隆市選區", "林沛祥"}, "基隆", "基隆市", true, 1},

		"newtaipei-1":  &Zone{"newtaipei-1", ZoneCandidate{"新北市第一選區", "洪孟楷"}, "石門、三芝等 6 區", "新北市", true, 1},
		"newtaipei-7":  &Zone{"newtaipei-7", ZoneCandidate{"新北市第七選區", "葉元之"}, "板橋", "新北市板橋區", true, 7},
		"newtaipei-8":  &Zone{"newtaipei-8", ZoneCandidate{"新北市第八選區", "張智倫"}, "中和", "新北市中和區", true, 8},
		"newtaipei-9":  &Zone{"newtaipei-9", ZoneCandidate{"新北市第九選區", "林德福"}, "永和、中和", "新北市", true, 9},
		"newtaipei-11": &Zone{"newtaipei-11", ZoneCandidate{"新北市第十一選區", "羅明才"}, "新店、深坑等 5 區", "新北市", true, 11},
		"newtaipei-12": &Zone{"newtaipei-12", ZoneCandidate{"新北市第十二選區", "廖先翔"}, "汐止、金山等 7 區", "新北市", true, 12},

		"taipei-3": &Zone{"taipei-3", ZoneCandidate{"臺北市第三選區", "王鴻薇"}, "中山、北松山", "臺北市", true, 3},
		"taipei-4": &Zone{"taipei-4", ZoneCandidate{"臺北市第四選區", "李彥秀"}, "內湖、南港", "臺北市", true, 4},
		"taipei-6": &Zone{"taipei-6", ZoneCandidate{"臺北市第六選區", "羅智強"}, "大安", "臺北市大安區", true, 6},
		"taipei-7": &Zone{"taipei-7", ZoneCandidate{"臺北市第七選區", "徐巧芯"}, "信義、南松山", "臺北市", true, 7},
		"taipei-8": &Zone{"taipei-8", ZoneCandidate{"臺北市第八選區", "賴士葆"}, "文山、南中正", "臺北市", true, 8},

		"taoyuan-1": &Zone{"taoyuan-1", ZoneCandidate{"桃園市第一選區", "牛煦庭"}, "蘆竹、龜山、桃園", "桃園市", true, 1},
		"taoyuan-2": &Zone{"taoyuan-2", ZoneCandidate{"桃園市第二選區", "涂權吉"}, "大園、觀音等 4 區", "桃園市", true, 2},
		"taoyuan-3": &Zone{"taoyuan-3", ZoneCandidate{"桃園市第三選區", "魯明哲"}, "中壢", "桃園市中壢區", true, 3},
		"taoyuan-4": &Zone{"taoyuan-4", ZoneCandidate{"桃園市第四選區", "萬美玲"}, "桃園", "桃園市桃園區", true, 4},
		"taoyuan-5": &Zone{"taoyuan-5", ZoneCandidate{"桃園市第五選區", "呂玉玲"}, "平鎮、龍潭", "桃園市", true, 5},
		"taoyuan-6": &Zone{"taoyuan-6", ZoneCandidate{"桃園市第六選區", "邱若華"}, "八德、大溪等 4 區", "桃園市", true, 6},

		"hsinchucity-1": &Zone{"hsinchucity-1", ZoneCandidate{"新竹市選區", "鄭正鈐"}, "新竹", "新竹市", true, 1},

		"hsinchucounty-1": &Zone{"hsinchucounty-1", ZoneCandidate{"新竹縣第一選區", "徐欣瑩"}, "新豐、湖口等 7 區", "新竹縣", true, 1},
		"hsinchucounty-2": &Zone{"hsinchucounty-2", ZoneCandidate{"新竹縣第二選區", "林思銘"}, "竹東、寶山等 7 區", "新竹縣", true, 2},

		"miaoli-1": &Zone{"miaoli-1", ZoneCandidate{"苗栗縣第一選區", "陳超明"}, "竹南、後龍等 8 區", "苗栗縣", true, 1},
		"miaoli-2": &Zone{"miaoli-2", ZoneCandidate{"苗栗縣第二選區", "邱鎮軍"}, "頭份、三灣等 10 區", "苗栗縣", true, 2},

		"taichung-2": &Zone{"taichung-2", ZoneCandidate{"臺中市第二選區", "顏寬恒"}, "沙鹿、霧峰等 5 區", "臺中市", true, 2},
		"taichung-3": &Zone{"taichung-3", ZoneCandidate{"臺中市第三選區", "楊瓊瓔"}, "大雅、潭子等 4 區", "臺中市", true, 3},
		"taichung-4": &Zone{"taichung-4", ZoneCandidate{"臺中市第四選區", "廖偉翔"}, "西屯、南屯", "臺中市", true, 4},
		"taichung-5": &Zone{"taichung-5", ZoneCandidate{"臺中市第五選區", "黃健豪"}, "北屯、北區", "臺中市", true, 5},
		"taichung-6": &Zone{"taichung-6", ZoneCandidate{"臺中市第六選區", "羅廷瑋"}, "中、西、東、南", "臺中市", true, 6},
		"taichung-8": &Zone{"taichung-8", ZoneCandidate{"臺中市第八選區", "江啓臣"}, "豐原、石岡等 5 區", "臺中市", true, 8},

		"changhua-3": &Zone{"changhua-3", ZoneCandidate{"彰化縣第三選區", "謝衣鳯"}, "第五、第七、第八", "彰化縣", true, 3},

		"nantou-1": &Zone{"nantou-1", ZoneCandidate{"南投縣第一選區", "馬文君"}, "埔里、草屯等 6 區", "南投縣", true, 1},
		"nantou-2": &Zone{"nantou-2", ZoneCandidate{"南投縣第二選區", "游顥"}, "南投、名間等 7 區", "南投縣", true, 2},

		"yunlin-1": &Zone{"yunlin-1", ZoneCandidate{"雲林縣第一選區", "丁學忠"}, "第三、第五、第六", "雲林縣", true, 1},

		"hualien-1": &Zone{"hualien-1", ZoneCandidate{"花蓮縣選區", "傅崐萁"}, "花蓮", "花蓮縣", true, 1},

		"taitung-1": &Zone{"taitung-1", ZoneCandidate{"臺東縣選區", "黃建賓"}, "臺東", "臺東縣", true, 1},

		"kinmen-1": &Zone{"kinmen-1", ZoneCandidate{"金門縣選區", "陳玉珍"}, "金門", "金門縣", true, 1},

		"lienchiang-1": &Zone{"lienchiang-1", ZoneCandidate{"連江縣選區", "陳雪生"}, "連江", "連江縣", false, 1},
	}

	cfg.Areas = cfg.ToAreas()

	if data, err := os.ReadFile("assets/candidate-areas-mapping.json"); err != nil {
		return nil, err
	} else if err := json.Unmarshal(data, &cfg.AreaFilter); err != nil {
		return nil, err
	}

	zoneCandidates := map[string]*ZoneCandidate{}
	for code, z := range cfg.Zones {
		zoneCandidates[code] = &z.ZoneCandidate
	}

	data, err := json.Marshal(zoneCandidates)
	if err != nil {
		return nil, err
	}

	cfg.ZoneCandidates = data

	return cfg, nil
}

func (r Config) HasZone(zone string) bool {
	_, exists := r.Zones[zone]
	return exists
}

func (r Config) GetZone(zone string) *Zone {
	z, exists := r.Zones[zone]
	if !exists {
		return nil
	}

	return z
}

func (r Config) ToAreas() []*Area {
	list := map[string][]*Zone{}
	for code, z := range r.Zones {
		pieces := strings.Split(code, "-")
		switch pieces[0] {
		case "taipei":
			if _, exists := list[AreaNameTaipei]; !exists {
				list[AreaNameTaipei] = []*Zone{}
			}
			list[AreaNameTaipei] = append(list[AreaNameTaipei], z)

		case "newtaipei":
			if _, exists := list[AreaNameNewTaipei]; !exists {
				list[AreaNameNewTaipei] = []*Zone{}
			}
			list[AreaNameNewTaipei] = append(list[AreaNameNewTaipei], z)

		case "keelung":
			if _, exists := list[AreaNameKeelung]; !exists {
				list[AreaNameKeelung] = []*Zone{}
			}
			list[AreaNameKeelung] = append(list[AreaNameKeelung], z)

		case "taoyuan":
			if _, exists := list[AreaNameTaoyuan]; !exists {
				list[AreaNameTaoyuan] = []*Zone{}
			}
			list[AreaNameTaoyuan] = append(list[AreaNameTaoyuan], z)

		case "hsinchucity":
			if _, exists := list[AreaNameHsinchuCity]; !exists {
				list[AreaNameHsinchuCity] = []*Zone{}
			}
			list[AreaNameHsinchuCity] = append(list[AreaNameHsinchuCity], z)

		case "hsinchucounty":
			if _, exists := list[AreaNameHsinchuCounty]; !exists {
				list[AreaNameHsinchuCounty] = []*Zone{}
			}
			list[AreaNameHsinchuCounty] = append(list[AreaNameHsinchuCounty], z)

		case "miaoli":
			if _, exists := list[AreaNameMiaoli]; !exists {
				list[AreaNameMiaoli] = []*Zone{}
			}
			list[AreaNameMiaoli] = append(list[AreaNameMiaoli], z)

		case "taichung":
			if _, exists := list[AreaNameTaichung]; !exists {
				list[AreaNameTaichung] = []*Zone{}
			}
			list[AreaNameTaichung] = append(list[AreaNameTaichung], z)

		case "changhua":
			if _, exists := list[AreaNameChanghua]; !exists {
				list[AreaNameChanghua] = []*Zone{}
			}
			list[AreaNameChanghua] = append(list[AreaNameChanghua], z)

		case "nantou":
			if _, exists := list[AreaNameNantou]; !exists {
				list[AreaNameNantou] = []*Zone{}
			}
			list[AreaNameNantou] = append(list[AreaNameNantou], z)

		case "yunlin":
			if _, exists := list[AreaNameYunlin]; !exists {
				list[AreaNameYunlin] = []*Zone{}
			}
			list[AreaNameYunlin] = append(list[AreaNameYunlin], z)

		case "hualien":
			if _, exists := list[AreaNameHualien]; !exists {
				list[AreaNameHualien] = []*Zone{}
			}
			list[AreaNameHualien] = append(list[AreaNameHualien], z)

		case "taitung":
			if _, exists := list[AreaNameTaitung]; !exists {
				list[AreaNameTaitung] = []*Zone{}
			}
			list[AreaNameTaitung] = append(list[AreaNameTaitung], z)

		case "kinmen":
			if _, exists := list[AreaNameKinmen]; !exists {
				list[AreaNameKinmen] = []*Zone{}
			}
			list[AreaNameKinmen] = append(list[AreaNameKinmen], z)

		case "lienchiang":
			if _, exists := list[AreaNameLienchiang]; !exists {
				list[AreaNameLienchiang] = []*Zone{}
			}
			list[AreaNameLienchiang] = append(list[AreaNameLienchiang], z)
		}
	}

	for _, zones := range list {
		sort.Slice(zones, func(i, j int) bool {
			return zones[i].sort < zones[j].sort
		})
	}

	sortedKeys := []string{
		AreaNameKeelung,
		AreaNameTaipei,
		AreaNameNewTaipei,
		AreaNameTaoyuan,
		AreaNameHsinchuCity,
		AreaNameHsinchuCounty,
		AreaNameMiaoli,
		AreaNameTaichung,
		AreaNameChanghua,
		AreaNameNantou,
		AreaNameYunlin,
		AreaNameHualien,
		AreaNameTaitung,
		AreaNameKinmen,
		AreaNameLienchiang,
	}

	areas := []*Area{}
	for _, key := range sortedKeys {
		if zones, exists := list[key]; exists {
			areas = append(areas, &Area{key, zones})
		}
	}

	return areas
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

type Zone struct {
	ZoneCode string
	ZoneCandidate
	Districts string
	Address   string
	Deployed  bool
	sort      int
}

type ZoneCandidate struct {
	ZoneName      string `json:"zoneName"`
	CandidateName string `json:"candidateName"`
}

func (r Zone) GetTopic() string {
	return r.CandidateName + " - " + r.ZoneName + " (" + r.Districts + ") "
}

type Area struct {
	Name  string
	Zones []*Zone
}

type TurnstileSiteverifyResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
	Messages   []string `json:"messages"`
}
