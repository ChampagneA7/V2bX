package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/InazumaV/V2bX/common/readiness"
	"github.com/spf13/cobra"
)

func init() {
	command.AddCommand(&cobra.Command{
		Use:   "check-ready PID",
		Short: "Check whether the selected local process initialized its nodes",
		Run: func(_ *cobra.Command, args []string) {
			if len(args) != 1 {
				fmt.Fprintln(os.Stderr, "check-ready requires a PID")
				os.Exit(1)
			}
			pid, err := strconv.Atoi(args[0])
			if err == nil {
				err = readiness.Check(pid)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, "V2bX is not ready:", err)
				os.Exit(1)
			}
			fmt.Println("V2bX nodes ready")
		},
	})
}
