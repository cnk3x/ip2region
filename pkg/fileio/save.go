package fileio

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
)

type SaveOptions struct {
	Overwrite   bool         `json:"overwrite,omitempty"`
	Mode        fs.FileMode  `json:"mode,omitempty"`
	TotalBytes  int64        `json:"total_bytes,omitempty"`
	UseTempFile bool         `json:"use_temp_file,omitempty"`
	Progress    ProgressHook `json:"-"`
	BeforeSave  []SaveHook   `json:"-"`
	AfterSave   []SaveHook   `json:"-"`
}

type ProgressHook func(cur, total int64)
type SaveHook func() error

func (options *SaveOptions) With(opts ...SaveOption) *SaveOptions {
	for _, opt := range opts {
		opt(options)
	}
	return options
}

type SaveOption func(*SaveOptions)

func Overwrite(opts *SaveOptions)      { opts.Overwrite = true }
func UseTempFile(opts *SaveOptions)    { opts.UseTempFile = true }
func Mode(mode fs.FileMode) SaveOption { return func(opts *SaveOptions) { opts.Mode = mode } }
func Progress(report ProgressHook) SaveOption {
	return func(opts *SaveOptions) { opts.Progress = report }
}
func TotalBytes(total int64) SaveOption { return func(opts *SaveOptions) { opts.TotalBytes = total } }

func BeforeSave(before SaveHook) SaveOption {
	return func(opts *SaveOptions) { opts.BeforeSave = append(opts.BeforeSave, before) }
}

func AfterSave(after SaveHook) SaveOption {
	return func(opts *SaveOptions) { opts.AfterSave = append(opts.AfterSave, after) }
}

func Save(src io.Reader, filePath string, opts ...SaveOption) (err error) {
	options := (&SaveOptions{Mode: 0666}).With(opts...)

	if err = os.MkdirAll(filepath.Dir(filePath), 0777); err != nil {
		return
	}

	handleCheck := func(filePath string, remove bool) (err error) {
		if stat, e := os.Lstat(filePath); e != nil {
			if !os.IsNotExist(e) {
				err = fmt.Errorf("target: %w", e)
				return
			}
		} else {
			if !options.Overwrite {
				err = fmt.Errorf("target %s already exists", filePath)
				return
			}

			if !stat.Mode().IsRegular() && stat.Mode()&fs.ModeSymlink == 0 {
				err = fmt.Errorf("file %s is not a regular file, can not overwrite", filePath)
				return
			}

			if remove {
				if err = os.Remove(filePath); err != nil {
					err = fmt.Errorf("can not overwrite file: %w", err)
				}
			}
		}

		return
	}

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

	handleRename := func(oldpath, newpath string) (err error) {
		if err = handleCheck(newpath, true); err != nil {
			return
		}
		err = os.Rename(oldpath, newpath)
		return
	}

	createFlag := func(overwrite bool) int {
		flag := os.O_RDWR | os.O_CREATE
		if overwrite {
			flag |= os.O_TRUNC
		} else {
			flag |= os.O_EXCL
		}
		return flag
	}

	copyIt := func(src io.Reader) func(f *os.File) error {
		return func(f *os.File) (err error) {
			var w io.Writer
			if options.Progress != nil {
				var cur int64
				w = Writer(func(b []byte) (n int, err error) {
					if n, err = f.Write(b); err != nil {
						return
					}
					options.Progress(atomic.AddInt64(&cur, int64(n)), options.TotalBytes)
					return
				})
			} else {
				w = f
			}
			_, err = io.Copy(w, src)
			return
		}
	}

	beforeSave := func() (err error) {
		for _, f := range options.BeforeSave {
			if err = f(); err != nil {
				return
			}
		}
		return
	}

	afterSave := func() (err error) {
		for _, f := range options.AfterSave {
			if err = f(); err != nil {
				return
			}
		}
		return
	}

	doSave := func(saveFn func() error) (err error) {
		if err = beforeSave(); err != nil {
			return
		}

		if err = saveFn(); err != nil {
			return
		}

		if err = afterSave(); err != nil {
			return
		}

		return
	}

	if options.UseTempFile {
		tempFilePath := filePath + ".savetmp"
		defer os.Remove(tempFilePath)

		if err = handleFileOpen(tempFilePath, createFlag(true), options.Mode, copyIt(src)); err != nil {
			return
		}

		return doSave(func() error { return handleRename(tempFilePath, filePath) })
	}

	if err = handleCheck(filePath, false); err != nil {
		return
	}

	return doSave(func() error {
		return handleFileOpen(filePath, createFlag(options.Overwrite), options.Mode, copyIt(src))
	})
}

func ConsoleProgress() SaveOption {
	var p int64
	return Progress(func(cur, total int64) {
		if total == 0 {
			slog.Info(fmt.Sprintf("地址库正在下载: %6s", HumanBytes(cur)))
			return
		}

		if c := cur * 10 / total; p != c {
			p = c
			slog.Info(fmt.Sprintf("地址库正在下载: %3d%% %7s", p*10, HumanBytes(cur)))
		}

		if cur == total {
			slog.Info("地址库下载完成")
		}
	})
}
