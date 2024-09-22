package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func createUpdateCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "update",
		Short: "更新数据库",
		Run: func(c *cobra.Command, args []string) {
			dbt, _ := c.Flags().GetString("type")
			s, err := createSearcher(c.Context(), dbt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return
			}
			defer s.Close()
			err = s.Update(c.Context())
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		},
	}

	c.Flags().StringP("type", "t", "mmdb", "数据库类型, xdb, mmdb")
	return c
}
