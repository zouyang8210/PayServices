package comm

const (
	ERR_CODE          = "err_code" //
	ERR_MSG           = "err_msg"  //
	ERR_LACK_PARAM    = 1001       //缺少参数
	ERR_INVALID_PARAM = 1002       //参数无效
	ERR_CALL_PARMENT  = 1003       //调用失败
	ERR_VERIFY_SIGN   = 1004       //验签失败
	MSG_IVALID_PARAM  = "无效的参数"
	MSG_VERIFY_SIGN   = "验签失败"
)

const (
	EMPTY         = ""   //aa
	OK            = "OK" //aa
	TEXT_HTML     = "text/html"
	BODY          = "body"          //订单标题
	TRADE_NO      = "trade_no"      //商户订单号
	NOTIFY_URL    = "notify_url"    //回调地址
	CLIENT_IP     = "clientIp"      //客户端IP
	FEE           = "fee"           //付款金额
	CODE          = "code"          //
	USER_AGENT    = "User-Agent"    //
	STATE         = "state"         //状态
	AUTH_CODE     = "auth_code"     //授权码
	OUT_REFUND_NO = "out_refund_no" //退款订单号
	REFUND_FEE    = "refund_fee"    //退款金额
	TOTAL_FEE     = "total_fee"     //标价总金额
	NOTIFY_INFO   = "notify_info"   //通步通知信息
	HTTP_SUCCESS  = 200             //
)
