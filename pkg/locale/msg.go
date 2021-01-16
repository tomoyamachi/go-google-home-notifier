package locale

import (
	"fmt"
	"time"
)

type Locale interface {
	Code() string
	NotifyMessage(start time.Time, title string) string
}

var locales = map[string]Locale{
	Ja{}.Code(): Ja{},
	En{}.Code(): En{},
}

func GetLocale(l string) Locale {
	target, ok := locales[l]
	if !ok {
		return En{}
	}
	return target
}

type Ja struct{}

func (_ Ja) Code() string { return "ja" }
func (_ Ja) NotifyMessage(start time.Time, title string) string {
	return fmt.Sprintf("%s から %s", start.Format("01月02日の15:04"), title)
}

type En struct{}

func (_ En) Code() string { return "en" }
func (_ En) NotifyMessage(start time.Time, title string) string {
	return fmt.Sprintf("%s will start from %s", title, start.Format("2006/01/02 15:04"))
}
