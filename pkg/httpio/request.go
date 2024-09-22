package httpio

import (
	"fmt"
	"io"
	"io/fs"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/cnk3x/ip2region/pkg/fileio"
)

const (
	HeaderContentType  = "Content-Type"
	MimeOctetStream    = "application/octet-stream"
	MimeFormUrlEncoded = "application/x-www-form-urlencoded"
)

// RequestOptions request options
type RequestOptions struct {
	URL        string
	Method     string
	CreateBody func() (body io.Reader, contentType string, err error)
	Headers    []string
}

// RequestOption request option
type RequestOption func(*RequestOptions)

func (f RequestOption) apply(ro *RequestOptions) { f(ro) }

// BuildRequest build request
func BuildRequest(options ...RequestOption) (req *http.Request, err error) {
	ropts := &RequestOptions{Method: http.MethodGet}
	for _, opt := range options {
		opt.apply(ropts)
	}

	var body io.Reader
	var contentType string
	if ropts.CreateBody != nil {
		if body, contentType, err = ropts.CreateBody(); err != nil {
			return
		}
	}

	if req, err = http.NewRequest(ropts.Method, ropts.URL, body); err != nil {
		return
	}

	for _, hdr := range ropts.Headers {
		nv := strings.SplitN(hdr, ":", 2)
		if len(nv) < 2 {
			nv = append(nv, "")
		}

		k, v := strings.TrimSpace(nv[0]), strings.TrimSpace(nv[1])
		if k == "" {
			continue
		}

		if k[0] == '+' {
			req.Header.Add(k, v)
		}

		switch k[0] {
		case '+':
			req.Header.Add(k[1:], v)
		case '-':
			req.Header.Del(k[1:])
		default:
			req.Header.Set(k, v)
		}
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return
}

func Headers(hdrs ...string) RequestOption {
	return func(ro *RequestOptions) {
		ro.Headers = append(ro.Headers, hdrs...)
	}
}

func HeaderSet(k, v string) RequestOption {
	return Headers(fmt.Sprintf("%s:%s", k, v))
}

func HeaderAdd(k, v string) RequestOption {
	return Headers(fmt.Sprintf("+%s:%s", k, v))
}

func HeaderDel(k string) RequestOption {
	return Headers(fmt.Sprintf("-%s", k))
}

func Form(data url.Values) RequestOption {
	return func(ro *RequestOptions) {
		ro.CreateBody = func() (io.Reader, string, error) {
			return strings.NewReader(data.Encode()), MimeFormUrlEncoded, nil
		}
	}
}

func File(fn string) RequestOption {
	return func(ro *RequestOptions) {
		ro.CreateBody = func() (body io.Reader, ct string, err error) {
			if body, err = os.Open(fn); err != nil {
				return
			}
			if ct = mime.TypeByExtension(filepath.Ext(fn)); ct == "" {
				ct = MimeOctetStream
			}
			return
		}
	}
}

func FormData(data url.Values, files map[string]string) RequestOption {
	return func(ro *RequestOptions) {
		handleFileOpen := func(name string, flag int, mode fs.FileMode, do func(f *os.File) error) (err error) {
			var f *os.File
			if f, err = os.OpenFile(name, flag, mode); err != nil {
				return
			}
			err = do(f)
			if e := f.Close(); e != nil && err == nil {
				err = e
			}
			return
		}

		ro.CreateBody = func() (io.Reader, string, error) {
			pr, pw := io.Pipe()
			bw := multipart.NewWriter(pw)

			checkErr := func() (err error) {
				do := func() (err error) {
					err = func() (err error) {
						for name, vs := range data {
							for _, v := range vs {
								var fw io.Writer
								if fw, err = bw.CreateFormField(name); err != nil {
									return
								}
								if _, err = fw.Write([]byte(v)); err != nil {
									return
								}
							}
						}

						for name, path := range files {
							if err = handleFileOpen(path, 0, 0, func(f *os.File) error {
								fw, err := bw.CreateFormFile(name, filepath.Base(path))
								if err != nil {
									return err
								}
								_, err = io.Copy(fw, f)
								return err
							}); err != nil {
								return
							}
						}

						return
					}()

					if e := bw.Close(); e != nil && err == nil {
						err = e
					}

					if e := pw.Close(); e != nil && err == nil {
						err = e
					}
					return
				}

				errc := make(chan error, 1)

				go func() {
					errc <- do()
					close(errc)
				}()

				select {
				case err = <-errc:
				default:
				}
				return
			}

			return fileio.Reader(func(b []byte) (n int, err error) {
				if err = checkErr(); err != nil {
					return
				}
				return pr.Read(b)
			}), bw.FormDataContentType(), nil
		}
	}
}
