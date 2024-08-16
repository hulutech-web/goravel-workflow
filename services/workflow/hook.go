package workflow

type Hook interface {
	//通知发起人
	NotifySendOne(id uint) error
	//通知下一审核人
	NotifyNextAuditor(id uint) error
}
