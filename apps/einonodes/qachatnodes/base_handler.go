package qachatnodes

import "errors"

type baseHandler struct {
}

func (b *baseHandler) checkState(state *State) error {
	if state == nil {
		return errors.New("state is nil")
	}
	if state.Ctx == nil {
		return errors.New("ctx is nil")
	}
	if state.SessionEntity == nil {
		return errors.New("sessionEntity is nil")
	}
	if state.QuestionEntity == nil {
		return errors.New("questionEntity is nil")
	}
	if state.ModelEntity == nil {
		return errors.New("modelEntity is nil")
	}
	if state.QuestionDbDatasetEntity == nil {
		return errors.New("questionDbDatasetEntity is nil")
	}
	return nil
}
