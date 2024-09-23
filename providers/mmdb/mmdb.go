package mmdb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/cnk3x/ip2region"
	"github.com/cnk3x/ip2region/pkg/fileio"
	"github.com/cnk3x/ip2region/pkg/httpio"
	"github.com/oschwald/geoip2-golang"
)

var dbDownloadUrl = "https://raw.gitmirror.com/P3TERX/GeoLite.mmdb/download/GeoLite2-City.mmdb"

// var dbDownloadUrl = "https://raw.gitmirror.com/P3TERX/GeoLite.mmdb/download/GeoLite2-Country.mmdb"
// var dbDownloadUrl = "https://raw.gitmirror.com/adysec/IP_database/main/geolite/GeoLite2-City.mmdb"
// var dbDownloadUrl = "https://raw.gitmirror.com/adysec/IP_database/main/geolite/GeoLite2-Country.mmdb"

type Options struct {
	DownloadUrl string
}

func Default() (p ip2region.Provider, err error) {
	return Open(context.Background(), fileio.DataFile("ip2region.mmdb"), nil)
}

func Open(ctx context.Context, dbFile string, options *Options) (p ip2region.Provider, err error) {
	if options == nil {
		options = &Options{}
	}
	if options.DownloadUrl == "" {
		options.DownloadUrl = dbDownloadUrl
	}

	s := &Provider{dbUrl: options.DownloadUrl, dbFile: dbFile}

	if err = fileio.CheckExist(dbFile, func() (err error) {
		slog.Info("地址库不存在，开始下载", "path", dbFile, "url", options.DownloadUrl)
		err = s.Update(ctx)
		return
	}); err != nil {
		return
	}

	return s, s.init()
}

type Provider struct {
	r      *geoip2.Reader
	dbUrl  string
	dbFile string
}

func (d *Provider) init() (err error) {
	if d.r, err = geoip2.Open(d.dbFile); err != nil {
		return
	}
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

func (d *Provider) Search(_ context.Context, ip string, langs ...string) (out *ip2region.Result, err error) {
	if d.r == nil {
		err = errors.New("reader is nil")
		return
	}

	var r *geoip2.City
	if r, err = d.r.City(net.ParseIP(ip)); err != nil {
		return
	}

	out = &ip2region.Result{IP: ip}

	out.Continent = getName(r.Continent.Names, r.Continent.Code, r.Continent.GeoNameID, langs...)
	out.Country = getName(r.Country.Names, r.Country.IsoCode, r.Country.GeoNameID, langs...)

	if len(r.Subdivisions) > 0 {
		out.Subdivision = getName(r.Subdivisions[0].Names, r.Subdivisions[0].IsoCode, r.Subdivisions[0].GeoNameID, langs...)
	}

	out.City = getName(r.City.Names, "", r.City.GeoNameID, langs...)

	return
}

func (d *Provider) Close() (err error) {
	if d.r != nil {
		err = d.r.Close()
		d.r = nil
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

func getName(names map[string]string, code string, geoNameID uint, langs ...string) ip2region.Name {
	if names != nil {
		var en bool
		for _, name := range langs {
			if v, ok := names[name]; ok && v != "" {
				return ip2region.NewName(v, code, geoNameID)
			}
			if name == "en" {
				en = true
			}
		}

		if !en {
			if v, ok := names["en"]; ok && v != "" {
				return ip2region.NewName(v, code, geoNameID)
			}
		}
	}
	return ip2region.NewName("", code, geoNameID)
}
