package cimd

const (
	STX                      = 2
	ETX                      = 3
	TAB                      = 9
	COLON                    = 58
	LOGIN                    = "01"
	USER_IDENTITY            = "010"
	PASSWORD                 = "011"
	ALIVE                    = "40"
	LOGOUT                   = "02"
	SUBMIT_MSG               = "03"
	DST_ADDR                 = "021"
	ORIG_ADDR                = "023"
	USER_DATA                = "033"
	USER_DATA_BINARY         = "034"
	SUCCESSFUL_DELIVERY      = "4"
	DELIVER_STAT_REPORT_RESP = "73"
	DELIVER_STAT_REPORT_REQ  = "23"
	LOGIN_RESP               = "51"
	LOGOUT_RESP              = "52"
	GENERAL_ERROR_RESP       = "98"
	ERROR_CODE               = "900"
	INVALID_LOGIN            = "100"
	UNEXPECTED_OPERATION     = "1"
	ALIVE_RESP               = "90"
	SUBMIT_MSG_RESP          = "53"
	DELIVER_MESSAGE          = "20"
	DELIVER_MESSAGE_RESP     = "70"
	DST_ADDR_RESP            = "021"
	SVC_CENTER_RESP          = "060"
	STATUS_CODE              = "061"
	DISCHARGE_TIME           = "063"
	STATUS_REPORT_REQUEST    = "056"
)
