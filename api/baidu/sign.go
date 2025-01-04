package baidu

import (
	"crypto/md5"
	"fmt"
)

// 生成方法：
// Step1. 将请求参数中的 APPID(appid)， 翻译 query(q，注意为UTF-8编码)，随机数(salt)，以及平台分配的密钥(可在管理控制台查看) 按照 appid+q+salt+密钥的顺序拼接得到字符串 1。
// Step2. 对字符串 1 做 MD5 ，得到 32 位小写的 sign。
// 注：
// 1. 待翻译文本（q）参数需为 UTF-8 编码；
// 2. 在生成签名拼接 appid+q+salt+密钥 字符串时，q 不需要做 URL encode，在生成签名之后，发送 HTTP 请求之前才需要对要发送的待翻译文本字段 q 做 URL encode；
// 3.如遇到报 54001 签名错误，请检查您的签名生成方法是否正确，在对 sign 进行拼接和加密时，q 不需要做 URL encode，很多开发者遇到签名报错均是由于拼接 sign 前就做了 URL encode；
// 4.在生成签名后，发送 HTTP 请求时，如果将 query 拼接在URL上，需要对 query 做 URL encode。
func sign(appid string, q string, salt string, secret string) string {
	s1 := fmt.Sprintf("%s%s%s%s", appid, q, salt, secret)
	return fmt.Sprintf("%x", md5.Sum([]byte(s1)))
}
