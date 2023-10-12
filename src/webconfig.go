package main

import "github.com/gofiber/fiber/v2"

var WebConfig = fiber.Map{
	"SiteName":                "short url",
	"SiteDescription":         "길이가 긴 URL을 짧게 줄여보세요.",
	"UrlInputPlaceHolder":     "줄이고 싶은 URL을 입력해주세요",
	"UrlButton":               "URL 검색",
	"InputBlinkError":         "내용을 입력해 주세요",
	"AddUrlTitle":             "새로운 URL 추가",
	"AddUrlDescription":       "필요한 내용을 작성해 주세요.",
	"KeyInputPlaceHolder":     "원하는 URL의 별칭을 입력해주세요",
	"AdminPwInputPlaceHolder": "관리용 패스워드를 입력해 주세요.",
	"AddUrlButton":            "URL 추가",
	"ApiErrorAlert":           "처리과정에서 오류가 발생하였습니다.",
	"DelUrlTitle":             "URL 관리",
	"DelUrlDescription":       "URL이 정상적으로 등록되었습니다. 이 페이지에서는 등록된 URL을 삭제할 수 있습니다.",
	"UrlLinkPlaceholder":      "새로운 URL",
	"UrlLink":                 "(접속하기)",
	"DelKeyInputPlaceHolder":  "삭제할 URL의 별칭을 입력해 주세요.",
	"UrlDelButton":            "URL 삭제",
	"HomeButton":              "돌아가기",
	"UrlDelSuccess":           "URL이 정상적으로 삭제되었습니다.",
	"Err404Title":             "404 Not Found",
	"Err404Description":       "이 페이지는 존재하지 않습니다.",
	"Err404ExtraDescrpition":  "URL이 삭제되었거나 이동되었을 수 있습니다.",
}

func ExtraWebConfig(extra fiber.Map) fiber.Map {
	newMap := map[string]interface{}{}

	for k, v := range WebConfig {
		newMap[k] = v
	}
	for k, v := range extra {
		newMap[k] = v
	}

	return newMap
}
