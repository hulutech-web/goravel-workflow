package workflow

type FlowPlugin struct {
}

// Signature The name and signature of the seeder.
func (s *FlowPlugin) Signature() string {
	return "FlowPlugin"
}

// Run executes the seeder logic.
func (s *FlowPlugin) Run() error {
	return nil
}
