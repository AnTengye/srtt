package baidu

// 输入参数
// 请求方式： 可使用 GET 或 POST 方式，如使用 POST 方式，Content-Type 请指定为：application/x-www-form-urlencoded
// 字符编码：统一采用 UTF-8 编码格式
// query 长度：为保证翻译质量，请将单次请求长度控制在 6000 bytes以内（汉字约为输入参数 2000 个）
//
// 字段名	类型	是否必填	描述	备注
// q	string	是	请求翻译query	UTF-8编码
// from	string	是	翻译源语言	可设置为auto
// to	string	是	翻译目标语言	不可设置为auto
// appid	string	是	APPID	可在管理控制台查看
// salt	string	是	随机数	可为字母或数字的字符串
// sign	string	是	签名	appid+q+salt+密钥的MD5值
type BaiduRequest struct {
	Q     string `json:"q"`
	From  string `json:"from"`
	To    string `json:"to"`
	Appid string `json:"appid"`
	Salt  string `json:"salt"`
	Sign  string `json:"sign"`
}

//输出参数
//返回的结果是json格式，包含以下字段：
//
//
//字段名	类型	描述	备注
//from	string	源语言	返回用户指定的语言，或者自动检测出的语种（源语言设为 auto 时）
//to	string	目标语言	返回用户指定的目标语言
//trans_result	array	翻译结果	返回翻译结果，包括 src 和 dst 字段
//trans_result.*.src	string	原文	接入举例中的“apple”
//trans_result.*dst	string	译文	接入举例中的“苹果”
//error_code	integer	错误码	仅当出现错误时显示

type BaiduResponse struct {
	From        string `json:"from"`
	To          string `json:"to"`
	TransResult []struct {
		Src string `json:"src"`
		Dst string `json:"dst"`
	} `json:"trans_result"`
	Error_code string `json:"error_code"`
}

// 自动检测	auto	中文	zh	英语	en
// 粤语	yue	文言文	wyw	日语	jp
// 韩语	kor	法语	fra	西班牙语	spa
// 泰语	th	阿拉伯语	ara	俄语	ru
// 葡萄牙语	pt	德语	de	意大利语	it
// 希腊语	el	荷兰语	nl	波兰语	pl
// 保加利亚语	bul	爱沙尼亚语	est	丹麦语	dan
// 芬兰语	fin	捷克语	cs	罗马尼亚语	rom
// 斯洛文尼亚语	slo	瑞典语	swe	匈牙利语	hu
// 繁体中文	cht	越南语	vie
var langMap = map[string]string{
	"auto": "auto",
	"zh":   "zh",
	"en":   "en",
	"ja":   "jp",
	"kor":  "kor",
	"fra":  "fra",
	"spa":  "spa",
	"ru":   "ru",
	"ara":  "ara",
	"th":   "th",
	"pt":   "pt",
	"de":   "de",
	"it":   "it",
	"el":   "el",
	"nl":   "nl",
	"pl":   "pl",
	"bul":  "bul",
	"est":  "est",
	"dan":  "dan",
	"fin":  "fin",
	"cs":   "cs",
	"rom":  "rom",
	"slo":  "slo",
	"swe":  "swe",
	"hu":   "hu",
	"cht":  "cht",
	"vie":  "vie",
}
