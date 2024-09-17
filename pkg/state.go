package pkg

// IsValidTransition checks if moving from current to new state is valid
func IsValidTransition(current, new TaskState) bool {
	transitions := map[TaskState][]TaskState{
		Pending:    {InProgress, Failed},
		InProgress: {Completed, Failed},
		Failed:     {},
		Completed:  {},
	}

	allowedTransitions, exists := transitions[current]
	if !exists {
		return false
	}

	for _, state := range allowedTransitions {
		if state == new {
			return true
		}
	}
	return false
}
