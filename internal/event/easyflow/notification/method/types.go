package method

import (
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/enotify/notify"
)

type NotifyParams struct {
	Rules      []Rule
	Order      order.Order
	Tasks      []model.Task
	WantResult map[string]interface{}
}

type NotifierIntegration interface {
	Builder(title string, users []user.User, params NotifyParams) []notify.NotifierWrap
}
