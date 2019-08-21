package ali_payment

import (
	"github.com/gin-gonic/gin"
	. "pay_service/module/comm"
	"sort"
	"utils/alipay"
	"utils/crypto"
	"utils/data_conv/json_lib"
	"utils/data_conv/number_lib"
	"utils/gin_check"
	"utils/http_lib"
)

var aliPay alipay.AliPayLib

//支付码交易返回
type RetAliPayMicroPay struct {
	ErrCode       int    `json:"err_code"`
	ErrMsg        string `json:"err_msg"`
	TradeNo       string `json:"trade_no"`       //支付宝订单号
	OutTradeNo    string `json:"out_trade_no"`   //商户订单号
	BuyerLogonId  string `json:"buyer_logon_id"` //买家支付定账号
	TotalAmount   string `json:"total_amount"`   //订单交易总金额
	ReceiptAmount string `json:"receipt_amount"` //实收金额
	EndTime       string `json:"gmt_payment"`    //交易支付时间
}

//支付宝退款信息
type RetAliPayRefund struct {
	ErrCode    int     `json:"err_code"`
	ErrMsg     string  `json:"err_msg"`
	TradeNo    string  `json:"trade_no"`     //支付宝订单号
	OutTradeNo string  `json:"out_trade_no"` //商户订单号
	RefundFee  float64 `json:"refund_fee"`   //退款金额
	EndTime    string  `json:"gmt_payment"`  //退款支付时间
}

//支付宝查询退款信息返回
type RetAliPayQueryRefund struct {
	ErrCode      int     `json:"err_code"`
	ErrMsg       string  `json:"err_msg"`
	TradeNo      string  `json:"trade_no"`      //支付宝订单号
	OutTradeNo   string  `json:"out_trade_no"`  //商户订单号
	TotalAmount  float64 `json:"total_amount"`  //该笔退款所对应的交易的订单金额
	RefundAmount float64 `json:"refund_amount"` //本次退款请求，对应的退款金额
}

//交易异步通知
type NotifyInfo struct {
	ErrCode       int     `json:"err_code"`
	ErrMsg        string  `json:"err_msg"`
	NotifyTime    string  `json:"notify_time"`    //通知时间
	NotifyType    string  `json:"notify_type"`    //通知类型
	NotifyId      string  `json:"notify_id"`      //通知校验ID
	TradeNo       string  `json:"trade_no"`       //支付宝交易号
	AppId         string  `json:"app_id"`         //开发者的app_id
	OutTradeNo    string  `json:"out_trade_no"`   //商户订单号
	BuyerLogonId  string  `json:"buyer_logon_id"` //买家支付宝账号
	TradeStatus   string  `json:"trade_status"`   //交易状态(WAIT_BUYER_PAY-交易创建,TRADE_CLOSED-关闭,TRADE_SUCCESS-完成,TRADE_FINISHED-交易结束,不可退款)
	TotalAmount   float64 `json:"total_amount"`   //订单金额
	ReceiptAmount float64 `json:"receipt_amount"` //实收金额
	RefundFee     float64 `json:"refund_fee,omitempty"`
}

var aliPayPublicKey string //支付宝平台公钥。是用签验平台返回和回调数据

func Init(appId, privateKey, publicKey string) {
	aliPay = alipay.AliPayLib{AppId: appId, PrivateKey: privateKey, PublicKey: publicKey}
	aliPayPublicKey = publicKey
}

//支付宝支付码交易
func AliPayMicroPay(c *gin.Context) {
	if jsonData, mapData, err := gin_check.CheckPostParameter(c, BODY, TRADE_NO, AUTH_CODE, TOTAL_FEE); err == nil {
		var reqData alipay.ScanDealInfo
		json_lib.JsonToObject(string(jsonData), &reqData)
		reqData.Subject = reqData.Body
		reqData.TotalFee = mapData[TOTAL_FEE].(float64) / float64(100)
		if info, err := aliPay.ScanCodePay(reqData); err == nil {
			var retInfo RetAliPayMicroPay
			json_lib.ObjectToObject(&retInfo, info)
			retInfo.ErrCode, retInfo.ErrMsg = aliPay.AnalysisReturn(info.RetAliPayBase)
			c.JSON(HTTP_SUCCESS, retInfo)
		} else {
			gin_check.SimpleReturn(ERR_CALL_PARMENT, err.Error(), c)
		}
	}
}

//支付宝退款
func AliPayRefund(c *gin.Context) {
	if _, mapData, err := gin_check.CheckPostParameter(c, TRADE_NO, OUT_REFUND_NO, REFUND_FEE); err == nil {
		if info, err := aliPay.Refund(mapData[TRADE_NO].(string), mapData[OUT_REFUND_NO].(string), mapData[REFUND_FEE].(float64)/100); err == nil {
			var retInfo RetAliPayRefund
			json_lib.ObjectToObject(&retInfo, info)
			retInfo.ErrCode, retInfo.ErrMsg = aliPay.AnalysisReturn(info.RetAliPayBase)
			c.JSON(HTTP_SUCCESS, retInfo)
			return
		} else {
			gin_check.SimpleReturn(ERR_CALL_PARMENT, err.Error(), c)
		}
	}
}

func AliH5Payment(dealInfo alipay.DealBaseInfo) (respBody string, err error) {
	respBody, err = aliPay.H5Pay(dealInfo)
	return
}

//支付宝退款查询
func AliPayQueryRefund(c *gin.Context) {
	if _, mapData, err := gin_check.CheckPostParameter(c, TRADE_NO, OUT_REFUND_NO); err == nil {
		if info, err := aliPay.QueryRefund(mapData[TRADE_NO].(string), mapData[OUT_REFUND_NO].(string)); err == nil {
			var retInfo RetAliPayQueryRefund
			json_lib.ObjectToObject(&retInfo, info)
			retInfo.ErrCode, retInfo.ErrMsg = aliPay.AnalysisReturn(info.RetAliPayBase)
			c.JSON(HTTP_SUCCESS, retInfo)
			return
		} else {
			gin_check.SimpleReturn(ERR_CALL_PARMENT, err.Error(), c)
		}
	}
}

//支付宝验签
func AliPayVerifySign(c *gin.Context) {
	if body, err := c.GetRawData(); err == nil {
		if b, notifyInfo := VerifySign(string(body)); b {
			c.JSON(HTTP_SUCCESS, notifyInfo)
		} else {
			gin_check.SimpleReturn(ERR_VERIFY_SIGN, MSG_VERIFY_SIGN, c)
		}
	} else {
		gin_check.SimpleReturn(ERR_INVALID_PARAM, MSG_IVALID_PARAM, c)
	}
}

//支付宝异步通知验签
func VerifySign(body string) (ret bool, notifyInfo NotifyInfo) {
	data, err := http_lib.GetUrlParams("http://127.0.0.1?" + body)
	if err == nil {
		sign := data["sign"]
		delete(data, "sign")
		delete(data, "sign_type")
		var keys []string
		for k := range data {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var waitSign = ""
		for i := 0; i < len(keys); i++ {
			waitSign += keys[i] + "=" + data[keys[i]] + "&"
		}
		waitSign = waitSign[0 : len(waitSign)-1]
		ret, err = crypto.VerifyRas2Sign(waitSign, sign, aliPayPublicKey)
		if ret {
			json_lib.ObjectToObject(&notifyInfo, data)
			number_lib.StrToFloat(data["total_amount"], &notifyInfo.TotalAmount)
			number_lib.StrToFloat(data["receipt_amount"], &notifyInfo.ReceiptAmount)
			if data["refund_fee"] != "" {
				number_lib.StrToFloat(data["refund_fee"], &notifyInfo.RefundFee)
			}
		}
	}
	return
}
