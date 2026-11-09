package format

import "fmt"

func Title(name string, count int) string { return fmt.Sprintf("%s: %d개", name, count) }
