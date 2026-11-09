package report

import "fmt"

// Totaler describes the single behavior the report needs.
type Totaler interface{ Total() int }

func Summary(value Totaler) string { return fmt.Sprintf("합계: %d원", value.Total()) }
