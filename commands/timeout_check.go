package commands

import (
	"fmt"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/hulutech-web/goravel-workflow/models"
)

type TimeoutCheckCommand struct{}

func NewTimeoutCheckCommand() console.Command {
	return &TimeoutCheckCommand{}
}

func (r *TimeoutCheckCommand) Signature() string {
	return "workflow:timeout-check"
}

func (r *TimeoutCheckCommand) Description() string {
	return "Check and handle overdue approval tasks"
}

func (r *TimeoutCheckCommand) Extend() command.Extend {
	return command.Extend{}
}

func (r *TimeoutCheckCommand) Handle(ctx console.Context) error {
	query := facades.Orm().Query()

	var procs []models.Proc
	if err := query.Model(&models.Proc{}).
		Where("status=?", models.ProcStatusPending).
		Find(&procs); err != nil {
		return err
	}

	now := carbon.Now()
	for _, proc := range procs {
		var process models.Process
		if err := query.Model(&models.Process{}).
			Where("id=?", proc.ProcessID).
			First(&process); err != nil {
			continue
		}

		if process.LimitTime <= 0 {
			continue // No timeout set for this step
		}

		// Use Concurrence (the approval time) to calculate elapsed seconds
		elapsedSeconds := proc.Concurrence.DiffInSeconds(now)

		if elapsedSeconds >= int64(process.LimitTime) {
			proc.Status = models.ProcStatusRejected
			proc.Content = "超时未处理，系统自动驳回"
			if saveErr := query.Model(&models.Proc{}).Where("id=?", proc.ID).Save(&proc); saveErr != nil {
				fmt.Printf("Failed to timeout-reject proc %d: %v\n", proc.ID, saveErr)
				continue
			}

			var entry models.Entry
			if err := query.Model(&models.Entry{}).Where("id=?", proc.EntryID).First(&entry); err == nil {
				entry.Status = models.ProcStatusRejected
				query.Model(&models.Entry{}).Where("id=?", entry.ID).Save(&entry)
				fmt.Printf("Timed out entry %d at process %d (%ds/%ds)\n", entry.ID, proc.ProcessID, elapsedSeconds, process.LimitTime)
			}
		}
	}

	return nil
}
