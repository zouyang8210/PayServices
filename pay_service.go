package main

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/url"
	"pay_service/module/alipay"
	. "pay_service/module/comm"
	"pay_service/module/wechat"
	"strings"
	"utils/alipay"
	"utils/data_conv/json_lib"
	"utils/data_conv/str_lib"
	"utils/file"
	"utils/gin_check"
	"utils/http_lib"
)

const OAUTH2_URL = "window.location.href='https://open.weixin.qq.com/connect/oauth2/authorize?"
const OAUTH2_PARAM = "appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_base&state=%s#wechat_redirect'"

//配置文件字段
const (
	WECHAT        = "WeChat"      //微信
	ALIPAY        = "AliPay"      //支付宝
	WX_APP_ID     = "wxAppId"     //微信公众号
	WX_MCH_ID     = "wxMchId"     //微信商户号
	WX_APP_SECRET = "wxAppSecret" //微信app密钥
	WX_API_SECRET = "wxApiSecret" //微信api密钥
	ALIPAY_APP_ID = "aliPayAppId" //支付宝AppId
)

//路径
const (
	CONF_PATH            = "conf/conf.txt"       //配置文件相对路径
	WX_RELATIVE_PATH     = "/payService/WeChat/" //微信接口相对路径
	ALIPAY_RELATIVE_PATH = "/payService/AliPay/" //支付宝接口相对路径
)

var service *gin.Engine

//微信变量
var (
	wxAppId            = EMPTY //微信appId
	wxMchId            = EMPTY //微信商户号
	wxAppSecret        = EMPTY //微信app密钥
	wxApiSecret        = EMPTY //微信api密钥
	wxPaymentNotify    = EMPTY //回调地址,用于OAUTH2
	wxMinProgramId     = EMPTY //小程序id
	wxMinProgramSecret = EMPTY //小程序密钥
)

//此页面返回到微信浏览器,来执行访问微信鉴权接口
const wxSkipPage = `<!DOCTYPE HTML>
<html>
<head>
<meta http-equiv="Content-CategoryName" content="text/html; charset=utf-8">
<meta content="width=device-width, initial-scale=0, maximum-scale=0, user-scalable=0" name="viewport" />
</head>
<script language="javascript">
执行脚本
</script>
<body>
</body>
</html>`

//支付宝公钥和私钥路径
const (
	ALIPAY_PUBLIC_KEY  = "resource/alipay_public.txt"  //支付宝平台公钥路径
	ALIPAY_PRIVATE_KEY = "resource/alipay_private.txt" //支付商户私钥路径
)

//微信证书路径
const (
	certFile = "resource/apiclient_cert.pem" //微信证书路径
	keyFile  = "resource/apiclient_key.pem"  //微信证书私钥路径
)

//支付宝变量
var (
	aliPayPublicKey  = EMPTY //支付宝商户私钥
	aliPayPrivateKey = EMPTY //支付宝平台公钥
	aliPayAppId      = EMPTY //支付宝appId
)

func main() {

	gin.SetMode(gin.DebugMode)

	service = gin.Default()
	service.Use(routerGateway)
	//微信支付接口
	service.POST(WxRelativePath("wxGetPayCode"), wechat_payment.WeChatGetPayCode)
	service.POST(WxRelativePath("wxMinProgramPay"), wechat_payment.WeChatMinProgramPay)
	service.POST(WxRelativePath("wxAppPay"), wechat_payment.WeChatAppPayment)
	service.GET(WxRelativePath("wxUnifyPay"), wechat_payment.WeChatUnifyPay)
	service.POST(WxRelativePath("wxMicroPay"), wechat_payment.WeChatMicroPay)
	service.POST(WxRelativePath("wxQueryTrade"), wechat_payment.WeChatQueryTrade)
	service.POST(WxRelativePath("wxRefund"), wechat_payment.WeChatRefund)
	service.POST(WxRelativePath("wxQueryRefund"), wechat_payment.WeChatQueryRefund)
	service.POST(WxRelativePath("wxPaymentNotifyVerify"), wechat_payment.WeChatPaymentNotifyVerify)
	service.POST(WxRelativePath("wxRefundNotifyDecode"), wechat_payment.WeChatRefundNotifyDecode)
	service.POST(WxRelativePath("wxReverse"), wechat_payment.WeChatReverse)
	//支付宝支付接口
	service.POST(AliPayRelativePath("aliPayMicroPay"), ali_payment.AliPayMicroPay)
	service.POST(AliPayRelativePath("aliPayRefund"), ali_payment.AliPayRefund)
	service.POST(AliPayRelativePath("aliPayQueryRefund"), ali_payment.AliPayQueryRefund)
	service.POST(AliPayRelativePath("AliPayVerifySign"), ali_payment.AliPayVerifySign)
	//微信,支付宝扫二合一码支付
	service.POST("/payService/unifyPayPage", unifyPayPage)

	service.Run(":8003") //启动服务
}

//统一支付
func unifyPayPage(c *gin.Context) {
	if _, mapData, err := gin_check.CheckPostParameter(c, BODY, TRADE_NO, NOTIFY_URL, TOTAL_FEE); err == nil {
		userAgent := c.GetHeader(USER_AGENT)
		fmt.Println("【统一支付USER AGENT】：", userAgent)
		if strings.Contains(userAgent, "MQQBrowser") {
			param := fmt.Sprintf("%s,%s,%s,%v", mapData[BODY], mapData[TRADE_NO], str_lib.UrlToUrlEncode(mapData[NOTIFY_URL].(string)), mapData[TOTAL_FEE])
			script := getOauth2Url(wxAppId, wxPaymentNotify, param)
			s := strings.Replace(wxSkipPage, "执行脚本", script, 1)
			c.Data(HTTP_SUCCESS, TEXT_HTML, []byte(s))
		} else if strings.Contains(userAgent, "UCBrowser") {
			var dealInfo alipay.DealBaseInfo
			json_lib.ObjectToObject(&dealInfo, mapData)
			dealInfo.Subject = dealInfo.Body
			dealInfo.TotalFee = dealInfo.TotalFee / 100
			dealInfo.NotifyUrl = mapData[NOTIFY_URL].(string)
			if payPage, err := ali_payment.AliH5Payment(dealInfo); err == nil {
				c.Data(HTTP_SUCCESS, TEXT_HTML, []byte(payPage))
			} else {
				gin_check.SimpleReturn(ERR_CALL_PARMENT, err.Error(), c)
			}
		}
	}
}

//oauth2跳转地址转码
func getOauth2Url(appId, redirectUri, param string) (url string) {
	redirectUri = str_lib.UrlToUrlEncode(redirectUri)
	url = OAUTH2_URL + fmt.Sprintf(OAUTH2_PARAM, appId, redirectUri, param)
	return
}

//读取资源文件
func readResourceFile() (err error) {
	var buff []byte
	//支付宝公钥
	buff, err = file.ReadFile(ALIPAY_PUBLIC_KEY)
	if err == nil {
		aliPayPublicKey = string(buff)
	}
	//支付宝私钥
	buff, err = file.ReadFile(ALIPAY_PRIVATE_KEY)
	if err == nil {
		aliPayPrivateKey = string(buff)
	}
	return
}

//微信支付相对路径组合
func WxRelativePath(interfaceName string) (path string) {
	path = WX_RELATIVE_PATH + interfaceName
	return
}

//支付宝支付相对路径组合
func AliPayRelativePath(interfaceName string) (path string) {
	path = ALIPAY_RELATIVE_PATH + interfaceName
	return
}

func init() {
	wxAppId = file.ReadConfig(WECHAT, WX_APP_ID, CONF_PATH)
	wxMchId = file.ReadConfig(WECHAT, WX_MCH_ID, CONF_PATH)
	wxAppSecret = file.ReadConfig(WECHAT, WX_APP_SECRET, CONF_PATH)
	wxApiSecret = file.ReadConfig(WECHAT, WX_API_SECRET, CONF_PATH)
	wxPaymentNotify = file.ReadConfig(WECHAT, "wxPaymentNotify", CONF_PATH)
	wxMinProgramId = file.ReadConfig(WECHAT, "wxMinProgramId", CONF_PATH)
	wxMinProgramSecret = file.ReadConfig(WECHAT, "wxMinProgramSecret", CONF_PATH)
	aliPayAppId = file.ReadConfig(ALIPAY, ALIPAY_APP_ID, CONF_PATH)
	if wxAppId == EMPTY || wxMchId == EMPTY || wxAppSecret == EMPTY || wxApiSecret == EMPTY {
		fmt.Println("read config file fail")
	}
	err := readResourceFile()
	if err != nil {
		fmt.Println("load resource file error:", err)
	}

	wechat_payment.Init(wxAppId, wxMchId, wxAppSecret, wxApiSecret, certFile, keyFile, wxMinProgramId, wxMinProgramSecret)
	ali_payment.Init(aliPayAppId, aliPayPrivateKey, aliPayPublicKey)
}

//路由网关
func routerGateway(c *gin.Context) {
	method := c.Request.Method
	//调试输出
	switch method {
	case "POST", "PATCH", "PUT":
		buffer, str, _ := http_lib.GetBody(c.Request)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(buffer))
		str, _ = url.QueryUnescape(str)
		fmt.Println("Input Parameter->", str)
	}
}
