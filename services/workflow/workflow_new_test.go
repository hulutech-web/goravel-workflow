package workflow

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUniqueSliceEmpty tests uniqueSlice with an empty slice.
func TestUniqueSliceEmpty(t *testing.T) {
	result := uniqueSlice([]int{})
	assert.Equal(t, []int{}, result)
}

// TestUniqueSliceNoDuplicates tests uniqueSlice with no duplicates.
func TestUniqueSliceNoDuplicates(t *testing.T) {
	result := uniqueSlice([]int{1, 2, 3})
	assert.Equal(t, []int{1, 2, 3}, result)
}

// TestUniqueSliceAllSame tests uniqueSlice where all elements are identical.
func TestUniqueSliceAllSame(t *testing.T) {
	result := uniqueSlice([]int{5, 5, 5, 5})
	assert.Equal(t, []int{5}, result)
}

// TestUniqueSliceWithZero tests uniqueSlice including zero values.
func TestUniqueSliceWithZero(t *testing.T) {
	result := uniqueSlice([]int{0, 1, 0, 2, 1, 0})
	assert.Equal(t, []int{0, 1, 2}, result)
}

// TestNewBaseWorkflowSingleton tests that NewBaseWorkflow returns the same instance.
func TestNewBaseWorkflowSingleton(t *testing.T) {
	w1 := NewBaseWorkflow()
	w2 := NewBaseWorkflow()
	assert.Same(t, w1, w2)
	assert.NotNil(t, w1.hooks)
}

// TestNilWorkflowNotifySendOne tests nil receiver handling for NotifySendOne.
func TestNilWorkflowNotifySendOne(t *testing.T) {
	var w *Workflow = nil
	err := w.NotifySendOne(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

// TestNilWorkflowNotifyNextAuditor tests nil receiver handling for NotifyNextAuditor.
func TestNilWorkflowNotifyNextAuditor(t *testing.T) {
	var w *Workflow = nil
	err := w.NotifyNextAuditor(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

// TestInvokeHooksWrongSignature tests that hooks with wrong signatures are safely ignored.
func TestInvokeHooksWrongSignature(t *testing.T) {
	w := NewBaseWorkflow()

	wrongHook := func(a uint, b string) {}
	wrongValue := reflect.ValueOf(wrongHook)
	w.RegisterHook("wrong_sig_test", wrongValue)

	// Should not panic
	w.invokeHooks("wrong_sig_test", 1)
}

// TestRegisterHookMultiple tests registering multiple hooks under the same name.
func TestRegisterHookMultiple(t *testing.T) {
	w := NewBaseWorkflow()

	hook1 := func(id uint) {}
	hook2 := func(id uint) {}
	w.RegisterHook("multi_hook_test", reflect.ValueOf(hook1))
	w.RegisterHook("multi_hook_test", reflect.ValueOf(hook2))

	assert.Len(t, w.hooks["multi_hook_test"], 2)
}

// TestRegisterHookNilMap tests RegisterHook when hooks map is nil.
func TestRegisterHookNilMap(t *testing.T) {
	w := &Workflow{hooks: nil}
	hookFn := func(id uint) {}
	w.RegisterHook("nil_map_test", reflect.ValueOf(hookFn))
	assert.NotNil(t, w.hooks)
	assert.Len(t, w.hooks["nil_map_test"], 1)
}

// TestConcTypeConstants verifies concurrency type constants.
func TestConcTypeConstants(t *testing.T) {
	assert.Equal(t, 0, ConcTypeSequential)
	assert.Equal(t, 1, ConcTypeConsensus)
	assert.Equal(t, 2, ConcTypeAny)
}

// TestAuditorConstants verifies auditor special value constants.
func TestAuditorConstants(t *testing.T) {
	assert.Equal(t, -1000, AuditorInitiator)
	assert.Equal(t, -1001, AuditorDirector)
	assert.Equal(t, -1002, AuditorManager)
	assert.Equal(t, -1003, AuditorFormField)
	assert.Equal(t, -1004, AuditorDynamicExpr)
}

// TestUniqueSliceMixedDuplicates tests uniqueSlice with mixed duplicates.
func TestUniqueSliceMixedDuplicates(t *testing.T) {
	result := uniqueSlice([]int{3, 1, 2, 3, 1, 4, 2})
	assert.Equal(t, []int{3, 1, 2, 4}, result)
}

// TestUniqueSliceSingleElement tests uniqueSlice with a single element.
func TestUniqueSliceSingleElement(t *testing.T) {
	result := uniqueSlice([]int{42})
	assert.Equal(t, []int{42}, result)
}

// TestInvokeHooksCorrectSignature tests that hooks with correct signature are called.
func TestInvokeHooksCorrectSignature(t *testing.T) {
	w := NewBaseWorkflow()
	called := false
	hookFn := func(id uint) {
		called = true
		assert.Equal(t, uint(99), id)
	}
	w.RegisterHook("correct_sig_test", reflect.ValueOf(hookFn))
	w.invokeHooks("correct_sig_test", 99)
	assert.True(t, called, "hook should have been called")
}

// TestInvokeHooksNotFound tests behavior when hook name doesn't exist.
func TestInvokeHooksNotFound(t *testing.T) {
	w := NewBaseWorkflow()
	// Should not panic
	w.invokeHooks("nonexistent_hook", 1)
}
