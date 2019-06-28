package wechat_payment

import (
	"encoding/xml"
	"fmt"
	"github.com/gin-gonic/gin"
	. "pay_service/module/comm"
	"strings"
	"time"
	"utils/data_conv/json_lib"
	"utils/data_conv/number_lib"
	"utils/data_conv/xml_lib"
	"utils/gin_check"
	"utils/http_lib"
	"utils/wechat"
	"utils/wechat/wechat_pay"
	"utils/wxpay"
)

type RetBase struct {
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
}

//支付付款返回信息
type RetMicroPay struct {
	RetBase
	Openid        string `xml:"openid"`         //用户标识
	TradeType     string `xml:"trade_type"`     //交易类型
	BankType      string `xml:"bank_type"`      //付款银行
	TransactionId string `xml:"transaction_id"` //微信订单号
	OutTradeNo    string `xml:"out_trade_no"`   //商户订单号
	TimeEnd       string `xml:"time_end"`       //支付完成时间
	TotalFee      int    `xml:"total_fee"`      //支付金额
	CashFee       int    `xml:"cash_fee"`       //现金支付金额
}

//查询订单返回信息
type RetQueryTrade struct {
	RetBase
	Openid          string `xml:"openid" json:"openid"`                      //用户标识
	TradeType       string `xml:"trade_type" json:"trade_type"`              //交易类型
	TradeStatus     string `xml:"trade_state" json:"trade_state"`            //交易状态
	BankType        string `xml:"bank_type" json:"bank_type"`                //付款银行
	TransactionId   string `xml:"transaction_id" json:"transaction_id"`      //微信订单号
	OutTradeNo      string `xml:"out_trade_no" json:"out_trade_no"`          //商户订单号
	TimeEnd         string `xml:"time_end" json:"time_end"`                  //支付完成时间
	TradeStatusDesc string `xml:"trade_state_desc" json:"trade_status_desc"` //对当前查询订单状态的描述和下一步操作的指引
	TotalFee        int    `xml:"total_fee" json:"total_fee"`                //标价金额
	CashFee         int    `xml:"cash_fee" json:"cash_fee"`                  //现金支付金额
}

//获取支付二维码返回信息
type RetPayCode struct {
	RetBase
	CodeUrl  string `json:"code_url"`  //二维码代码
	PrepayId string `json:"prepay_id"` //微信订单号
}

//退款返回信息
type RetRefund struct {
	RetBase
	TransactionId string `json:"transaction_id"` //微信订单号
	OutTradeNo    string `json:"out_trade_no"`   //商户订单号
	OutRefundNo   string `json:"out_refund_no"`  //商户退款单号
	RefundId      string `json:"refund_id"`      //微信退款单号
	TotalFee      int    `json:"total_fee"`      //标价金额
	RefundFee     int    `json:"refund_fee"`     //退款金额
	CashFee       int    `json:"cash_fee"`       //现金支付金额
}

//查询退款返回信息
type RetQueryRefund struct {
	RetBase
	TransactionId string `xml:"transaction_id" json:"transaction_id"` //微信支付订单号
	OutTradeNo    string `xml:"out_trade_no" json:"out_trade_no"`     //商户订单号
	RefundId      string `xml:"refund_id" json:"refund_id"`           //微信退款单号
	OutRefundNo   string `xml:"out_refund_no" json:"out_refund_no"`   //商户退款订单号
	TotalFee      int    `xml:"total_fee" json:"total_fee"`           //订单金额
}

//支付结果异步通知信息
type RetPaymentNotifyInfo struct {
	RetBase
	AppId          string `xml:"appid" json:"appid"`                       //公众账号ID
	MchId          string `xml:"mch_id" json:"mch_id"`                     //商户号
	OpenId         string `xml:"openid" json:"openid"`                     //用户标识
	TradeType      string `xml:"trade_type" json:"trade_type"`             //交易类型
	TotalFee       int    `xml:"total_fee" json:"total_fee"`               //订单金额
	CashFee        int    `xml:"cash_fee" json:"cash_fee"`                 //现金支付金额
	CouponFee      int    `xml:"coupon_fee" json:"coupon_fee"`             //总代金券金额
	TransactionId  string `xml:"transaction_id" json:"transaction_id"`     //微信支付订单号
	OutTradeNo     string `xml:"out_trade_no" json:"out_trade_no"`         //商户订单号
	TimeEnd        string `xml:"time_end" json:"time_end"`                 //支付完成时间
	TradeStateDesc string `xml:"trade_state_desc" json:"trade_state_desc"` //订单状态描述
	TradeState     string `xml:"trade_state" json:"trade_state"`           //订单状态
}

//退款异步通知信息
type RetRefundNotifyInfo struct {
	RetBase
	AppId               string `xml:"appid" json:"appid"`                                 //公众账号ID
	MchId               string `xml:"mch_id" json:"mch_id"`                               //商户号
	TransactionId       string `xml:"transaction_id" json:"transaction_id"`               //微信支付订单号
	OutTradeNo          string `xml:"out_trade_no" json:"out_trade_no"`                   //商户订单号
	RefundId            string `xml:"refund_id" json:"refund_id"`                         //微信退款单号
	OutRefundNo         string `xml:"out_refund_no" json:"out_refund_no"`                 //商户退款订单号
	TotalFee            int    `xml:"total_fee" json:"total_fee"`                         //订单金额
	SettlementTotalFee  int    `xml:"settlement_total_fee" json:"settlement_total_fee"`   //应结订单金额(当该订单有使用非充值券时，返回此字段。应结订单金额=订单金额-非充值代金券金额，应结订单金额<=订单金额)
	RefundFee           int    `xml:"refund_fee" json:"refund_fee"`                       //申请退款金额
	SettlementRefundFee int    `xml:"settlement_refund_fee" json:"settlement_refund_fee"` //退款金额
	RefundStatus        string `xml:"refund_status" json:"refund_status"`                 //退款状态(SUCCESS-退款成功,CHANGE-退款异常,REFUNDCLOSE—退款关闭)
	SuccessTime         string `xml:"success_time" json:"success_time"`                   //退款成功时间
	RefundRecvAccout    string `xml:"refund_recv_accout" json:"refund_recv_accout"`       //退款入账账户
	RefundAccount       string `xml:"refund_account"`                                     //退款资金来源(REFUND_SOURCE_RECHARGE_FUNDS-可用余额退款/基本账户,REFUND_SOURCE_UNSETTLED_FUNDS-未结算资金退款)
	RefundRequestSource string `xml:"refund_request_source" json:"refund_request_source"` //退款发起来源(API-接口,VENDOR_PLATFORM-商户平台)
}

//const (
//	OAUTH2_URL   = "window.location.href='https://open.weixin.qq.com/connect/oauth2/authorize?"
//	OAUTH2_PARAM = "appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_base&state=%s#wechat_redirect'"
//)

//证书相对路径
var (
	certFile = "resource/apiclient_cert.pem"
	keyFile  = "resource/apiclient_key.pem"
)
var (
	wxApiSecret string
)

//模板
const (
	wxPaymentPage = `<!DOCTYPE HTML>
<html>
    <head>
        <meta http-equiv="Content-CategoryName" content="text/html; charset=utf-8">
        <meta content="width=device-width, initial-scale=0, maximum-scale=0, user-scalable=0" name="viewport" />
    </head>
    <style>
    </style>
    <body>
    </body>
<script>
    function onBridgeReady() {
        WeixinJSBridge.invoke(
                'getBrandWCPayRequest',
                {
                    "appId": "参数1",   //公众号名称，由商户传入
                    "timeStamp": "参数2",//时间戳，自1970年以来的秒数
                    "nonceStr": "参数3",//随机串
                    "package": "参数4",
                    "signType": "参数5",//微信签名方式:
                    "paySign": "参数6"   //微信签名
                },
                function (res) {
                    if (res.err_msg == "get_brand_wcpay_request:ok") {
                        Toast("支付成功");
                    } else if (res.err_msg == "get_brand_wcpay_request:cancel") {
                        Toast("支付取消");
                    } else {
                        Toast("支付失败");
                    }
                }
        );
        //Toast提示
        function Toast(msg) {
            setTimeout(function () {
                document.getElementById('toast-msg').innerHTML = msg;
                var toastTag = document.getElementById('toast-wrap');
                toastTag.className = toastTag.className + 'toastAnimate';
                setTimeout(function () {
                    toastTag.className = toastTag.className.replace('toastAnimate', '');
                }, 1500);
            }, 100);
        }
    }
    if (typeof WeixinJSBridge == "undefined"){
        if( document.addEventListener ){
            document.addEventListener('WeixinJSBridgeReady', onBridgeReady, false);
        }else if (document.attachEvent){
            document.attachEvent('WeixinJSBridgeReady', onBridgeReady);
            document.attachEvent('onWeixinJSBridgeReady', onBridgeReady);
        }
    }else{
        onBridgeReady();
    }
</script>
</html>`
)

var wxPay wechat_pay.WXPay //微信支付对像

func Init(appId, mchId, appSecret, apiSecret string, cFile, kFile string, MinProgramId, MinProgramSecret string) {
	wxPay = wechat_pay.WXPay{AppId: appId, MchId: mchId, AppSecret: appSecret, ApiSecret: apiSecret, MinProgramId: MinProgramId,
		MinProgramSecret: MinProgramSecret}
	wxApiSecret = apiSecret
	certFile = cFile
	keyFile = kFile
	str := `<xml><return_code>SUCCESS</return_code><appid><![CDATA[wx121ec1cc9aafbe5b]]></appid><mch_id><![CDATA[1486136062]]></mch_id><nonce_str><![CDATA[f3d600f86cb57c38509cad76d968dfdf]]></nonce_str><req_info><![CDATA[zV0t9wutGE/5Ucs1gGsCfFV17Cu3kJrxdc8SbSEQR1uVNtpJ7q1Pgvhprn9FYDN96NbF/D4UK8kJk6aAIiDa9s77gxhSmVIKe78H9y2bY14rGD2JnTDqk7bym6oBhsgmO CYMlzougEVdxXMkn0AAWbytYJ0Y7VRM/JNLHPpkmEueIgp0yLOfiPMH//TVTIJ4cqttRyg65IMSl19/T1gwmspVLmY2U12a748NIwe5RVtP57KQs3aidD8BTKLpov7sPD7Qv4EpkI7qu/q9nspvr65EOCTRYnpQ9jR5uF0AYcZbySmTnXf/9c/zAy2tI26vL1LEGcs79eRRk8od82lSoXbGXsHG3k463eyxT7h9Tf3LhJvTMw7Gx8trxhpUbj1o7dCJEsgyC36bAR4vGe1HqQj3iHWeFuyUia5OBmKeR50ONycMdYT4qsSQv6W8Ic670Wgj0krkU1bvyOGXl7GMpevwgA0QG85BY8FlRZzs9JSXXIlShR Jiv6pzITwGmDNqQUfUTphbgqbNmpW4/hR0uXcQqLVb091xdSESFz/bjIFXjOi0qsq13mSee91rqLS3DIq2MJNcLoeJ9qI9KpSjn2fNXWF14kDWyUhGL1LuAN1NFQXtP3nu2 x5bAPEhlSdEB8y8GueStRlD RAyVYAVMQiw9OPJwnCKwOwO77jyYksWWeNFV/mZ unp7PRQNRLU6w8XlheEli2irw8l0DNHJtN3D/Iu4BaDmDFmkKCW6rqpymkcIOB5CPB V9RuP6a8lupcoBCaEAAKRrDauEBHjH4D 94VVTqePBT97a5m90zWSOMyzcGuAhgdmpfomrCUy3AO8dRBwLirpTIbt4XGR60DobPzTvmIjida6wzcgyirl5ymAqFm3nn8ik6zXMENE037sz1FM51XRak0i0kVjAApnQqwwhMLhxmNhdSjsveLslcfDoBUdcA82eM4K/3JkeGC 9R/zfKEq5Ca192IS87pZBcpL4LoGSHDQ2kRc3IQl6h/I ej9UYDfW3Nixd5A3nSi1d8zZz5Zp0Jd8UtFFb1Tvv AJjesDLAR0Nw1IEK wNF6SV9j1RkS7/5xJfItiHJDwYYyNBFGOP2xlA==]]></req_info></xml>`
	if info, err := wxDecodeRefundNotify(str); err == nil {
		fmt.Printf("%#v", info)
	} else {
		fmt.Printf("解密出错:%v", err)
	}

}

//获取商家支付码
func WeChatGetPayCode(c *gin.Context) {
	var ret RetPayCode
	if _, mapData, err := gin_check.CheckPostParameter(c, BODY, TRADE_NO, NOTIFY_URL, CLIENT_IP, FEE); err == nil {
		if info, err := wxPay.GetPayCode(mapData[BODY].(string), mapData[TRADE_NO].(string), mapData[NOTIFY_URL].(string),
			mapData[CLIENT_IP].(string), int(mapData[FEE].(float64))); err == nil {
			json_lib.ObjectToObject(&ret, info)
			ret.ErrCode, ret.ErrMsg = wechat.AnalysisWxReturn(info.RetBase, info.RetPublic)
			c.JSON(HTTP_SUCCESS, ret)
			return
		} else {
			gin_check.SimpleReturn(ERR_CALL_PARMENT, err.Error(), c)
			return
		}
	}
}

//微信小程序支付
func WeChatMinProgramPay(c *gin.Context) {
	var retInfo wechat.RetMinProgramPay
	if _, mapData, err := gin_check.CheckPostParameter(c, BODY, TRADE_NO, NOTIFY_URL, CODE, FEE); err == nil {
		retInfo = wxPay.MinProgramPlaceOrder(mapData[BODY].(string), mapData[TRADE_NO].(string), mapData[NOTIFY_URL].(string),
			mapData[CODE].(string), int(mapData[FEE].(float64)))
		if retInfo.ErrCode != 0 {
			retInfo.ErrCode = ERR_CALL_PARMENT
		}
		c.JSON(HTTP_SUCCESS, retInfo)
	}
}

//微信APP支付
func WeChatAppPayment(c *gin.Context) {
	var retInfo wechat.RetAppPay
	if _, mapData, err := gin_check.CheckPostParameter(c, BODY, TRADE_NO, NOTIFY_URL, FEE); err == nil {
		retInfo = wxPay.AppPlaceOrder(mapData[BODY].(string), mapData[TRADE_NO].(string), mapData[NOTIFY_URL].(string),
			int(mapData[FEE].(float64)))
		if retInfo.ErrCode != 0 {
			retInfo.ErrCode = ERR_CALL_PARMENT
		}
		c.JSON(HTTP_SUCCESS, retInfo)
	}
}

//微信统一支付
func WeChatUnifyPay(c *gin.Context) {
	state := c.Query(STATE)
	code := c.Query(CODE)
	if state != EMPTY && code != EMPTY {
		params := strings.Split(state, ",")
		var fee int
		number_lib.StrToInt(params[3], &fee)
		info := wxPay.PublicPlaceOrder(params[0], params[1], params[2], code, fee)
		sFile := strings.Replace(wxPaymentPage, "参数1", info.AppId, 1)
		sFile = strings.Replace(sFile, "参数2", info.TimeStamp, 1)
		sFile = strings.Replace(sFile, "参数3", info.NonceStr, 1)
		sFile = strings.Replace(sFile, "参数4", info.Package, 1)
		sFile = strings.Replace(sFile, "参数5", info.SignType, 1)
		sFile = strings.Replace(sFile, "参数6", info.PaySign, 1)
		c.Data(HTTP_SUCCESS, TEXT_HTML, []byte(sFile))
	} else {
		gin_check.SimpleReturn(ERR_LACK_PARAM, "缺少参数:state 或 code", c)
	}
}

//微信支付码支付
func WeChatMicroPay(c *gin.Context) {
	var retInfo RetMicroPay
	var info wechat.RetMicroPay
	if _, mapData, err := gin_check.CheckPostParameter(c, BODY, TRADE_NO, AUTH_CODE, NOTIFY_URL, TOTAL_FEE); err == nil {
		tradeNo := mapData[TRADE_NO].(string)
		if info, err = wxPay.MicroPay(mapData[BODY].(string), tradeNo, mapData[NOTIFY_URL].(string),
			mapData[AUTH_CODE].(string), int(mapData[TOTAL_FEE].(float64))); err == nil {
			json_lib.ObjectToObject(&retInfo, info)
			retInfo.ErrCode, retInfo.ErrMsg = wechat.AnalysisWxReturn(info.RetBase, info.RetPublic)
			if info.ResultCode == weixin.SUCCESS {
				go wxQueryMicroTrade(tradeNo, mapData[NOTIFY_URL].(string))
			}
			c.JSON(HTTP_SUCCESS, retInfo)
		} else {
			gin_check.SimpleReturn(ERR_CALL_PARMENT, err.Error(), c)
		}
	}
}

//查询微信订单状态
func WeChatQueryTrade(c *gin.Context) {
	var retInfo RetQueryTrade
	var info wechat.RetQuery
	if _, mapData, err := gin_check.CheckPostParameter(c, TRADE_NO); err == nil {
		if info, err = wxPay.QueryOrder(mapData[TRADE_NO].(string)); err == nil {
			json_lib.ObjectToObject(&retInfo, info)
			retInfo.ErrCode, retInfo.ErrMsg = wechat.AnalysisWxReturn(info.RetBase, info.RetPublic)
			c.JSON(HTTP_SUCCESS, retInfo)
		} else {
			gin_check.SimpleReturn(ERR_CALL_PARMENT, err.Error(), c)
		}
	}
}

//微信退款
func WeChatRefund(c *gin.Context) {
	var info wechat.RetRefund
	if _, mapData, err := gin_check.CheckPostParameter(c, TRADE_NO, OUT_REFUND_NO, REFUND_FEE, TOTAL_FEE, NOTIFY_URL); err == nil {
		var retInfo RetRefund
		if info, err = wxPay.Refund(mapData[TRADE_NO].(string), mapData[OUT_REFUND_NO].(string), mapData[NOTIFY_URL].(string),
			int(mapData[TOTAL_FEE].(float64)), int(mapData[REFUND_FEE].(float64)), certFile, keyFile); err == nil {
			json_lib.ObjectToObject(&retInfo, info)
			retInfo.ErrCode, retInfo.ErrMsg = wechat.AnalysisWxReturn(info.RetBase, info.RetPublic)
			c.JSON(HTTP_SUCCESS, retInfo)
		} else {
			gin_check.SimpleReturn(ERR_CALL_PARMENT, err.Error(), c)
		}
	}
}

//支付结果异步通知验签
func WeChatPaymentNotifyVerify(c *gin.Context) {
	if _, mapData, err := gin_check.CheckPostParameter(c, NOTIFY_INFO); err == nil {
		if b, retInfo := wxVerifyPaymentNotify(mapData[NOTIFY_INFO].(string)); b {
			c.JSON(HTTP_SUCCESS, retInfo)
		} else {
			gin_check.SimpleReturn(ERR_VERIFY_SIGN, MSG_VERIFY_SIGN, c)
			return
		}
	}
}

//退款订单异步通知解密
func WeChatRefundNotifyDecode(c *gin.Context) {
	if _, mapData, err := gin_check.CheckPostParameter(c, NOTIFY_INFO); err == nil {
		retInfo, err := wxDecodeRefundNotify(mapData[NOTIFY_INFO].(string))
		if err != nil {
			retInfo.ErrCode = -1
			retInfo.ErrMsg = err.Error()
		}
		c.JSON(HTTP_SUCCESS, retInfo)
	}
}

//退款订单查询
func WeChatQueryRefund(c *gin.Context) {
	var info wechat.RetQueryRefund
	if _, mapData, err := gin_check.CheckPostParameter(c, OUT_REFUND_NO); err == nil {
		var retInfo RetQueryRefund
		if info, err = wxPay.QueryRefund(mapData[OUT_REFUND_NO].(string), EMPTY); err == nil {
			json_lib.ObjectToObject(&retInfo, info)
			retInfo.ErrCode, retInfo.ErrMsg = wechat.AnalysisWxReturn(info.RetBase, info.RetPublic)
			c.JSON(HTTP_SUCCESS, retInfo)
		} else {
			gin_check.SimpleReturn(ERR_CALL_PARMENT, err.Error(), c)
		}
	}
}

func WeChatReverse(c *gin.Context) {
	if _, mapData, err := gin_check.CheckPostParameter(c, "out_trade_no"); err == nil {
		if resp, err := wxPay.Reverse(mapData["out_trade_no"].(string), "resource/apiclient_cert.pem", "resource/apiclient_key.pem"); err == nil {
			c.JSON(HTTP_SUCCESS, resp)
		} else {
			fmt.Println(err)
		}
	}
}

//查询微信支付码支付订单状态
func wxQueryMicroTrade(tradeNo string, notifyUrl string) {
	var info wechat.RetQuery
	for i := 0; i < 15; i++ {
		info, _ = wxPay.QueryOrder(tradeNo)
		if info.TradeStatus == weixin.SUCCESS {
			//提交订单状态
			var desc wechat.PaymentNotifyInfo
			json_lib.ObjectToObject(&desc, info)
			xmlBuffer, _ := xml.Marshal(desc)
			http_lib.HttpSubmit(http_lib.POST, notifyUrl, string(xmlBuffer), nil)
			break
		}
		time.Sleep(2 * time.Second)
	}
	return
}

func wxDecodeRefundNotify(xmlStr string) (retInfo RetRefundNotifyInfo, err error) {
	var info wechat.RefundNotifyInfo
	var buff []byte
	xml_lib.XmlToObject(xmlStr, &info)
	buff, err = wechat.DecodeRefundData(info.ReqInfo, wxApiSecret)
	fmt.Printf(info.ReqInfo)
	xml_lib.XmlToObject(string(buff), &info.RefundEncryptInfo)
	json_lib.ObjectToObject(&retInfo, info.RefundEncryptInfo)
	return
}

func wxVerifyPaymentNotify(xmlStr string) (b bool, retInfo RetPaymentNotifyInfo) {
	var info wechat.PaymentNotifyInfo
	xml_lib.XmlToObject(xmlStr, &info)
	fmt.Println(json_lib.ObjectToJson(info))
	if b = wechat.VerifySign(info, info.Sign, wxApiSecret); b {
		json_lib.ObjectToObject(&retInfo, info)
		//retInfo.ErrCode, retInfo.ErrMsg = wechat.AnalysisWxReturn(info.RetBase, info.RetPublic)
	}
	return
}
