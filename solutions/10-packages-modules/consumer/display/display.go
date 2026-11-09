package display

type Describer interface{ Description() string }

func Line(value Describer) string { return "- " + value.Description() }
