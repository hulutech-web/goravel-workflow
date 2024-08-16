package workflow

type Hook interface {
	NotifyStartThis(id uint) error
	NotifyNextAuditor(id uint) error
}
