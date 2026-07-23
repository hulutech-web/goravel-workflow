package workflow

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/http"
	commands "github.com/hulutech-web/goravel-workflow/commands"
)

type Workflow struct {
	Context http.Context
}

func NewWorkflow(ctx http.Context) *Workflow {
	return &Workflow{
		Context: ctx,
	}
}

// RegisterCommands registers all workflow commands
func (w *Workflow) RegisterCommands() []console.Command {
	return []console.Command{
		commands.NewTimeoutCheckCommand(),
	}
}
