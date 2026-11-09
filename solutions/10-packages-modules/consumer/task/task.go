package task

type Task struct {
	Title string
	Done  bool
}

func (t Task) Description() string {
	if t.Done {
		return t.Title + " (완료)"
	}
	return t.Title + " (진행)"
}
