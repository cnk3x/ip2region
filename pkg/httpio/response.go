package httpio

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/cnk3x/ip2region/pkg/fileio"
)

type (
	// ResponseProcess 响应处理
	ResponseProcess func(resp *http.Response) (err error)
	// ResponseMiddleware 响应中间件
	ResponseMiddleware func(next ResponseProcess) ResponseProcess
)

// ProcessResponse 处理响应
func ProcessResponse(resp *http.Response, process ResponseProcess, middlewares ...ResponseMiddleware) (err error) {
	body := resp.Body
	defer body.Close()
	resp.Body = io.NopCloser(body)
	for i := len(middlewares) - 1; i >= 0; i-- {
		process = middlewares[i](process)
	}
	return process(resp)
}

// WriteTo write response body to writer
func WriteTo(w io.Writer) ResponseProcess {
	return func(resp *http.Response) (err error) {
		_, err = io.Copy(w, resp.Body)
		return
	}
}

// Download write response body to file
func Download(filePath string, saveOptions ...fileio.SaveOption) ResponseProcess {
	return func(resp *http.Response) (err error) {
		return fileio.Save(resp.Body, filePath, saveOptions...)
	}
}

// JSON parse response body to json
func JSON(out any) ResponseProcess {
	return func(resp *http.Response) (err error) {
		return json.NewDecoder(resp.Body).Decode(out)
	}
}

// StatusOK check response status code
func StatusOK(next ResponseProcess) ResponseProcess {
	return func(resp *http.Response) (err error) {
		if resp.StatusCode >= http.StatusBadRequest {
			err = fmt.Errorf("http status code is not 200, %d", resp.StatusCode)
		} else {
			err = next(resp)
		}
		return
	}
}

type ProgressState struct {
	Total   int64
	Current int64
	StartAt time.Time
	EndAt   time.Time
	Speed   float64
}

func (p ProgressState) String() string {
	return fmt.Sprintf("total: %s, current: %s, speed: %s/s",
		fileio.HumanBytes(p.Total), fileio.HumanBytes(p.Current), fileio.HumanBytes(p.AvSpeed()))
}

// Percent 返回百分比, 0-100, 精度2位小数，如果 total 为 0 则返回 -1
func (p ProgressState) Percent() float64 {
	return fileio.MathDiv(p.Current*100, p.Total, 2)
}

func (p ProgressState) Completed() bool {
	return p.Current == p.Total
}

func (p ProgressState) AvSpeed() float64 {
	return fileio.MathDiv(p.Current, p.Elapsed().Seconds(), 0)
}

func (p ProgressState) Elapsed() time.Duration {
	if p.EndAt.IsZero() {
		return time.Since(p.StartAt)
	}
	return p.EndAt.Sub(p.StartAt)
}

// Progress report progress
func Progress(report func(p ProgressState)) ResponseMiddleware {
	return func(next ResponseProcess) ResponseProcess {
		if report == nil {
			return next
		}

		return func(resp *http.Response) (err error) {
			var (
				r   = resp.Body
				p   = ProgressState{Total: resp.ContentLength, StartAt: time.Now()}
				t   time.Time
				s   int32
				old int64
			)

			doReport := func(downloaded int, completed bool) {
				cur := atomic.AddInt64(&p.Current, int64(downloaded))

				now := time.Now()

				if completed {
					p.Speed = fileio.MathDiv(cur-old, now.Sub(t).Seconds())
					p.EndAt = now
					t = now
					old = cur
					report(p)
					return
				}

				if atomic.CompareAndSwapInt32(&s, 0, 1) {
					p.Speed = fileio.MathDiv(downloaded, now.Sub(p.StartAt).Seconds())
					t = now
					old = cur
					report(p)
					return
				}

				if elapsed := now.Sub(t); elapsed >= time.Second {
					p.Speed = fileio.MathDiv(cur-old, elapsed.Seconds())
					t = now
					old = cur
					report(p)
					return
				}
			}

			resp.Body = fileio.ReadCloser(func(b []byte) (n int, err error) {
				n, err = r.Read(b)
				doReport(n, err == io.EOF)
				return
			})
			return next(resp)
		}
	}
}
