package cmd

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/jon4hz/deadshot/internal/logging"

	"github.com/spf13/cobra"
)

var last bool

var logCmd = &cobra.Command{
	Use:   "logs",
	Short: "Helper command to deal with logs",
	Long:  "Helper command to deal with logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !last {
			return errors.New("please set at least one flag")
		}
		return logs(last)
	},
}

func init() {
	logCmd.Flags().BoolVarP(&last, "last", "l", false, "Print the last log")
}

func logs(last bool) error {
	if last {
		file, err := logging.GetLastLog()
		if err != nil {
			switch err.(type) {
			case *fs.PathError:
				return nil
			default:
				return err
			}
		}
		fmt.Println(file)
	}
	return nil
}
