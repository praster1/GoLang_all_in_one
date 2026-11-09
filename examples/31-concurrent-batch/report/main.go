package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func main() {
	results, err := batch.Run(context.Background(), []string{"go", "한글"}, batch.Config{Workers: 2, Attempts: 1}, func(ctx context.Context, job batch.Job) (string, error) { return strings.ToUpper(job.Input), ctx.Err() })
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	report := batch.NewReport(results)
	fmt.Printf("총계=%d 성공=%d 실패=%d 취소=%d 시도=%d\n", report.Summary.Total, report.Summary.Succeeded, report.Summary.Failed, report.Summary.Canceled, report.Summary.Attempts)
	for _, item := range report.Items {
		fmt.Printf("%d %s -> %s\n", item.Index, item.Input, item.Output)
	}
}
