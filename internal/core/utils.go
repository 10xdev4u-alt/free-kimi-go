package core

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func UUID(separator bool) string {
	id := uuid.New().String()
	if !separator {
		return id
	}
	return id
}

func UnixTimestamp() int64 {
	return time.Now().Unix()
}

func Timestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func GenerateRandomString(length int, charset string) string {
	var letters []rune
	if charset == "numeric" {
		letters = []rune("0123456789")
	} else {
		letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	}

	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GenerateCookie() string {
	timestamp := UnixTimestamp()
	items := []string{
		fmt.Sprintf("Hm_lvt_358cae4815e85d48f7e8ab7f3680a74b=%d", timestamp-int64(rand.Intn(2592000))),
		fmt.Sprintf("_ga=GA1.1.%s.%d", GenerateRandomString(10, "numeric"), timestamp-int64(rand.Intn(2592000))),
		fmt.Sprintf("_ga_YXD8W70SZP=GS1.1.%d.1.1.%d.0.0.0", timestamp-int64(rand.Intn(2592000)), timestamp-int64(rand.Intn(2592000))),
		fmt.Sprintf("Hm_lpvt_358cae4815e85d48f7e8ab7f3680a74b=%d", timestamp-int64(rand.Intn(2592000))),
	}
	cookie := ""
	for i, item := range items {
		if i > 0 {
			cookie += "; "
		}
		cookie += item
	}
	return cookie
}

func MD5(value string) string {
	hash := md5.Sum([]byte(value))
	return hex.EncodeToString(hash[:])
}

const (
	FakeUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"
)

func GetFakeHeaders() map[string]string {
	return map[string]string{
		"Accept":             "*/*",
		"Accept-Encoding":    "gzip, deflate, br, zstd",
		"Accept-Language":    "zh-CN,zh;q=0.9",
		"Origin":             "https://kimi.moonshot.cn",
		"Cookie":             GenerateCookie(),
		"R-Timezone":         "Asia/Shanghai",
		"Sec-Ch-Ua":          `"Google Chrome";v="123", "Not:A-Brand";v="8", "Chromium";v="123"`,
		"Sec-Ch-Ua-Mobile":   "?0",
		"Sec-Ch-Ua-Platform": `"Windows"`,
		"Sec-Fetch-Dest":     "empty",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Site":     "same-origin",
		"User-Agent":         FakeUserAgent,
	}
}
