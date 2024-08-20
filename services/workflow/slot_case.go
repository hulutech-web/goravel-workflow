package workflow

type slotCaseProxy interface {
	Register()
	Uninstall()
}

type SlotCaseService interface {
}

type slotCaseServiceImpl struct {
	slotCase slotCaseProxy
}

func NewSlotCaseService(slotCase slotCaseProxy) SlotCaseService {
	return &slotCaseServiceImpl{slotCase: slotCase}
}

func (s *slotCaseServiceImpl) Register() {
}

func (s *slotCaseServiceImpl) Uninstall() {
}

type MockSlotCaseService interface {
}

type mockSlotCaseServiceImpl struct {
	slotCase slotCaseProxy
}

func NewMockSlotCaseService(slotCase slotCaseProxy) SlotCaseService {
	return &slotCaseServiceImpl{slotCase: slotCase}
}

func (s *mockSlotCaseServiceImpl) Register() {
}

func (s *mockSlotCaseServiceImpl) Uninstall() {
}
