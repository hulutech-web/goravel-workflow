package commands

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

type PublishWorkflow struct{}

func NewPublishWorkflow() *PublishWorkflow {
	return &PublishWorkflow{}
}
func (receiver *PublishWorkflow) Signature() string {
	return "workflow:publish"
}

// Description The console command description.
func (receiver *PublishWorkflow) Description() string {
	return "发布workflow资源"
}

// Extend The console command extend.
func (receiver *PublishWorkflow) Extend() command.Extend {
	return command.Extend{}
}

// Handle Execute the console command.
func (receiver *PublishWorkflow) Handle(ctx console.Context) error {
	return nil
}
