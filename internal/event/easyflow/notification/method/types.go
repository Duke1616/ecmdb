package method

import (
	"github.com/Bunny3th/easy-workflow/workflow/model"
	"github.com/Duke1616/ecmdb/internal/order"
	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/Duke1616/enotify/notify"
)

type NotifierIntegration interface {
	Builder(rules []Rule, order order.Order, startUser string, users []user.User,
		tasks []model.Task) []notify.NotifierWrap
}
