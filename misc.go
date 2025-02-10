package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func isValidMobileNumber(mobile string) bool {
	return regexp.MustCompile(`^09\d{8}$`).MatchString(mobile)
}

func isValidIdNumber(text string) bool {
	if len(text) != 10 {
		return false
	}

	letterMap := map[string]int{
		"A": 10, "B": 11, "C": 12, "D": 13, "E": 14,
		"F": 15, "G": 16, "H": 17, "J": 18, "K": 19,
		"L": 20, "M": 21, "N": 22, "P": 23, "Q": 24,
		"R": 25, "S": 26, "T": 27, "U": 28, "V": 29,
		"X": 30, "Y": 31, "W": 32, "Z": 33, "I": 34,
		"O": 35,
	}

	firstLetter := string(text[0])
	if _, ok := letterMap[firstLetter]; !ok {
		return false
	}

	firstDigit := letterMap[firstLetter]
	digits := []int{
		firstDigit / 10,
		firstDigit % 10,
	}

	for i := 1; i < 10; i++ {
		num, err := strconv.Atoi(string(text[i]))
		if err != nil {
			return false
		}
		digits = append(digits, num)
	}

	checksum := digits[0] + digits[1]*9 + digits[2]*8 + digits[3]*7 + digits[4]*6 + digits[5]*5 + digits[6]*4 + digits[7]*3 + digits[8]*2 + digits[9] + digits[10]
	return checksum%10 == 0
}

var arabicToChinese = map[string]string{
	"1":  "一",
	"2":  "二",
	"3":  "三",
	"4":  "四",
	"5":  "五",
	"6":  "六",
	"7":  "七",
	"8":  "八",
	"9":  "九",
	"10": "十",
}

func sanitizeAddress(address string) string {
	address = strings.ReplaceAll(address, "-", "之")

	pattern := regexp.MustCompile(`(\d+)\s*(段|樓)`)
	converted := pattern.ReplaceAllStringFunc(address, func(match string) string {
		subMatch := pattern.FindStringSubmatch(match)
		if len(subMatch) < 3 {
			return match
		}

		arabic := subMatch[1]
		unit := subMatch[2]

		if chinese, ok := arabicToChinese[arabic]; ok {
			return chinese + unit
		}

		if num := convertLargeNumber(arabic); num != "" {
			return num + unit
		}

		return match
	})

	return converted
}

func convertLargeNumber(num string) string {
	if n := len(num); n == 1 {
		return arabicToChinese[num]
	} else if n == 2 {
		digits := strings.Split(num, "")
		result := ""
		if digits[0] == "1" {
			result += "十"
		} else {
			result += arabicToChinese[digits[0]] + "十"
		}
		if digits[1] != "0" {
			result += arabicToChinese[digits[1]]
		}
		return result
	}
	return ""
}

func PrintStructWithJSON(strc interface{}) {
	jsonData, err := json.MarshalIndent(strc, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonData))
}
