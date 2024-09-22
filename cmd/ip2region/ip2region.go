package main

import (
	"context"
	"fmt"

	"github.com/cnk3x/ip2region"
	"github.com/cnk3x/ip2region/pkg/fileio"
	"github.com/cnk3x/ip2region/providers/mmdb"
	"github.com/cnk3x/ip2region/providers/xdb"
)

func createSearcher(ctx context.Context, dbt string) (ip2region.Provider, error) {
	switch dbt {
	case "mmdb":
		return mmdb.Open(ctx, fileio.DataFile("ip2region.mmdb"), nil)
	case "xdb":
		return xdb.Open(ctx, fileio.DataFile("ip2region.xdb"), &xdb.Options{Cache: xdb.Content})
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbt)
	}
}
