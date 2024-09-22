package xdb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cnk3x/ip2region"
	"github.com/cnk3x/ip2region/pkg/fileio"
	"github.com/cnk3x/ip2region/pkg/httpio"
	"github.com/cnk3x/ip2region/providers/xdb/internal/xdb"
)

var (
	dbDownloadUrl = "https://raw.gitmirror.com/adysec/IP_database/main/ip2region/ip2region.xdb"
)

type CachePolicy string

const (
	File    CachePolicy = "file"
	Content CachePolicy = "content"
	Index   CachePolicy = "index"
)

type Provider struct {
	xdb    *xdb.Searcher
	policy CachePolicy
	dbUrl  string
	dbFile string
}

type Options struct {
	DownloadUrl string
	Cache       CachePolicy
}

func Open(ctx context.Context, dbPath string, options *Options) (p ip2region.Provider, err error) {
	if options == nil {
		options = &Options{}
	}

	if options.DownloadUrl == "" {
		options.DownloadUrl = dbDownloadUrl
	}

	if options.Cache == "" {
		options.Cache = File
	}

	s := &Provider{dbUrl: options.DownloadUrl, dbFile: dbPath, policy: options.Cache}

	if err = fileio.CheckExist(dbPath, func() (err error) {
		slog.Info("地址库不存在，开始下载", "path", dbPath, "url", options.DownloadUrl)
		err = s.Update(ctx)
		return
	}); err != nil {
		return
	}

	return s, s.init()
}

func (d *Provider) init() (err error) {
	var s *xdb.Searcher
	switch d.policy {
	case File:
		s, err = xdb.NewWithFileOnly(d.dbFile)
	case Index:
		var vi []byte
		if vi, err = xdb.LoadVectorIndexFromFile(d.dbFile); err == nil {
			s, err = xdb.NewWithVectorIndex(d.dbFile, vi)
		}
	case Content:
		var buf []byte
		if buf, err = xdb.LoadContentFromFile(d.dbFile); err == nil {
			s, err = xdb.NewWithBuffer(buf)
		}
	default:
		err = fmt.Errorf("错误的缓存策略 `%s`", d.policy)
	}

	if err != nil {
		if e := errors.Unwrap(err); e != nil {
			err = e
		}
		return
	}

	d.xdb = s
	return
}

func (d *Provider) Update(ctx context.Context) (err error) {
	return httpio.New(d.dbUrl).Use(httpio.StatusOK, httpio.Progress(consoleProgress)).
		Do(ctx, httpio.Download(
			d.dbFile,
			fileio.UseTempFile,
			fileio.Overwrite,
			fileio.BeforeSave(d.Close),
			fileio.AfterSave(d.init),
		))
}

func (d *Provider) Search(_ context.Context, ip string, _ ...string) (result *ip2region.Result, err error) {
	if d.xdb == nil {
		err = errors.New("reader is nil")
		return
	}

	var ipInt uint32
	if ipInt, err = xdb.IP2Int(ip); err != nil {
		return
	}
	return d.search(ipInt)
}

func (d *Provider) search(ip uint32) (result *ip2region.Result, err error) {
	var r string
	if r, err = d.xdb.Search(ip); err != nil {
		return
	}

	// 国家|0|省/州|城市|网络运营商
	rs := strings.SplitN(r, "|", 5)
	if len(rs) != 5 {
		err = fmt.Errorf("无效查询结果: %s", r)
		return
	}

	for i := range rs {
		if rs[i] == "0" {
			rs[i] = ""
		} else if i > 0 && rs[i] == rs[i-1] {
			rs[i] = ""
		}
	}

	result = &ip2region.Result{
		IP:          xdb.Long2IP(ip),
		Country:     ip2region.NewName(rs[0], "", 0),
		Continent:   ip2region.NewName(rs[1], "", 0),
		Subdivision: ip2region.NewName(rs[2], "", 0),
		City:        ip2region.NewName(rs[3], "", 0),
		ISP:         rs[4],
	}
	return
}

func (d *Provider) Close() (err error) {
	if d.xdb != nil {
		d.xdb.Close()
		d.xdb = nil
	}
	return
}

func consoleProgress(p httpio.ProgressState) {
	slog.Info(
		fmt.Sprintf("地址库正在下载: %6.2f%% %17s %8s/s",
			p.Percent(),
			fmt.Sprintf("%s/%s", fileio.HumanBytes(p.Current), fileio.HumanBytes(p.Total)),
			fileio.HumanBytes(p.Speed),
		),
	)
	if p.Completed() {
		slog.Info("地址库下载完成", "大小", fileio.HumanBytes(p.Total), "耗时", fileio.HumanDuration(p.Elapsed()), "均速", fileio.HumanBytes(p.AvSpeed())+"/s")
	}
}
