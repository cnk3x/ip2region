package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func createQueryCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "query",
		Short: "查询IP地址",
		Args:  cobra.MinimumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			dbt, _ := c.Flags().GetString("type")
			s, err := createSearcher(c.Context(), dbt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return
			}
			defer s.Close()

			fmt.Fprintln(os.Stdout)
			fmt.Fprintln(os.Stdout)
			for _, ip := range args {
				if r, e := s.Search(c.Context(), ip); e != nil {
					fmt.Fprintln(os.Stderr, e.Error())
				} else {
					fmt.Fprintln(os.Stdout, r.String())
				}
			}
		},
	}

	c.Flags().StringP("type", "t", "mmdb", "数据库类型, xdb, mmdb")
	return c
}

//104.28.103.41 192.63.101.81 76.147.224.166 76.147.224.166 177.239.36.171 99.145.201.231 72.217.24.193 148.74.198.214 64.228.109.115 67.251.212.23 72.193.22.78 96.51.48.247 96.51.48.247 71.202.58.195 142.189.55.152 72.200.85.185 75.56.242.135 47.208.94.170 104.203.37.214 73.217.247.67 103.112.165.105 64.223.224.229 64.223.169.156 168.203.21.218 108.147.32.62 174.104.120.105
