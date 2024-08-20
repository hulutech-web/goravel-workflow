package workflow

type Hook interface {
	NotifySendOne(id uint) error
	NotifyNextAuditor(id uint) error
}
