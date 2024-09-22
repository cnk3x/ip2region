package main

import (
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/spf13/cobra"
)

func createWebCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "web",
		Short: "启动web服务",
		Args:  cobra.NoArgs,
		Run: func(c *cobra.Command, args []string) {
			dbt, _ := c.Flags().GetString("type")
			s, err := createSearcher(c.Context(), dbt)
			if err != nil {
				slog.Error("创建搜索器失败", "type", dbt, "err", err)
				return
			}
			defer s.Close()

			mux := chi.NewMux()
			mux.Use(middleware.Recoverer, middleware.Logger, cors.AllowAll().Handler, middleware.RealIP)

			mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
				ip := r.FormValue("ip")
				if ip == "" {
					if ip, _, _ = net.SplitHostPort(r.RemoteAddr); ip == "" {
						ip = r.RemoteAddr
					}
				}

				result, err := s.Search(r.Context(), ip)
				if err != nil {
					webErr(w, r, err, 500)
					return
				}
				result.IP = ip
				webRespond(w, r, result, result.String(), 200)
			})

			mux.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
				if err = s.Update(c.Context()); err != nil {
					webErr(w, r, err, 500)
				} else {
					webRespond(w, r, render.M{"msg": "OK"}, "OK", 200)
				}
			})

			listen, _ := c.Flags().GetString("listen")
			slog.Info("listen", "addr", listen)
			http.ListenAndServe(listen, mux)
		},
	}

	c.Flags().StringP("listen", "l", ":3824", "监听地址")
	c.Flags().StringP("type", "t", "xdb", "数据库类型, xdb, mmdb")

	return c
}

func webErr(w http.ResponseWriter, r *http.Request, err error, status int) {
	render.Status(r, status)
	render.Respond(w, r, render.M{"err": err.Error()})
}

func webRespond(w http.ResponseWriter, r *http.Request, data any, dataStr string, status int) {
	if userAgent := r.Header.Get("user-agent"); userAgent == "" || strings.Contains(userAgent, "curl") {
		render.Status(r, status)
		render.PlainText(w, r, dataStr)
	} else {
		render.Status(r, status)
		render.Respond(w, r, data)
	}
}
