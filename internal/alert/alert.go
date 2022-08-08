package alert

import "fmt"

type AlertType uint8

const (
	INFO AlertType = iota
	SUCCESS
	ERROR
	WARNING
)

var (
	alertMap map[AlertType]string = map[AlertType]string{
		INFO:    "資訊",
		SUCCESS: "成功",
		ERROR:   "失敗",
		WARNING: "警告",
	}
)

func (alertType AlertType) String() string {
	return alertMap[alertType]
}

type Alert struct {
	Type    AlertType `json:"type"`
	Message string    `json:"message"`
}

func (alert *Alert) String() string {
	return fmt.Sprintf("[訊息] 類型:%s 訊息:%s", alert.Type, alert.Message)
}

var (
	DataFormatError      = Alert{ERROR, "資料格式錯誤"}
	NotLoggedIn          = Alert{ERROR, "您尚未登入"}
	IllegalBehavior      = Alert{ERROR, "不允許的行為"}
	BadServer            = Alert{ERROR, "伺服器發生錯誤"}
	IllegalData          = Alert{ERROR, "發現非法資料"}
	NotFoundAccount      = Alert{ERROR, "不存在的使用者帳號"}
	WrongPassword        = Alert{ERROR, "密碼錯誤"}
	AccountAlreadyExists = Alert{ERROR, "使用者帳號已存在"}

	LoggedInSuccess  = Alert{SUCCESS, "登入成功"}
	LoggedOutSuccess = Alert{SUCCESS, "登出成功"}
	EditSuccess      = Alert{SUCCESS, "編輯成功"}
)
