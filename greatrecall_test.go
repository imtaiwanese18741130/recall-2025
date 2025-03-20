package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type GreatRecallTestSuite struct {
	suite.Suite
	apiKey string
	url    string
}

func (s *GreatRecallTestSuite) SetupSuite() {
	s.url = os.Getenv("GREATRECALL_API_URL")
	if s.url == "" {
		s.T().Skip("GREATRECALL_API_URL not defined")
	}
	s.apiKey = os.Getenv("GREATRECALL_API_KEY")
	if s.apiKey == "" {
		s.T().Skip("GREATRECALL_API_KEY not defined")
	}
}

func TestGreatRecallSuite(t *testing.T) {
	suite.Run(t, new(GreatRecallTestSuite))
}

func (s *GreatRecallTestSuite) TestStandardAddress() {
	testCases := map[string]string{
		// 中華民國全國建築師公會: "台" -> "臺" + 半形數字 + "-" -> "之"
		"台北市信義區基隆路2段51號13樓-3": "臺北市信義區景聯里4鄰基隆路二段51號十三樓之３",

		// 美商甲骨文有限公司台灣分公司: 全形數字 + 錯誤的「鄰」
		"臺北市信義區興雅里２６鄰忠孝東路５段６８號２８樓": "臺北市信義區興雅里24鄰忠孝東路五段68號二十八樓",

		// 雅虎台灣: 國字數字
		"臺北市羅斯福路二段一百號十二樓": "臺北市中正區板溪里2鄰羅斯福路二段100號十二樓",

		// 祥程工程有限公司 (廢止): 國字數字 + 門牌整編前地址 + -> 現址應為 "2弄1號二樓"
		"臺北縣中和市國光街一零九巷二弄一之一號": "新北市中和區自強里7鄰國光街109巷2弄1之1號",

		// 薰衣草森林: 缺少「村」
		"新竹縣尖石鄉嘉樂129號1樓": "新竹縣尖石鄉嘉樂村11鄰嘉樂129號",
	}
	for input, expected := range testCases {
		s.Run(input, func() {
			actual, err := standardAddress(input)
			s.Require().NoError(err)
			s.Require().Equal(expected, actual)
		})
	}
}
