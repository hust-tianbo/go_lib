package rollwriter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"
)

type AsyncOptions struct {
	LogQueueSize      int
	WriteLogSize      int  // 刷盘的大小，单位字节
	WriterLogInterval int  // 刷盘的间隔时间，单位ms
	CanDropLog        bool // 是否丢弃日志
}

type AsyncOption func(*AsyncOptions)

func WithLogQueueSize(n int) AsyncOption {
	return func(opts *AsyncOptions) {
		opts.LogQueueSize = n
	}
}

func WithWriteLogSize(size int) AsyncOption {
	return func(opts *AsyncOptions) {
		opts.WriteLogSize = size
	}
}

func WithWriteLogInterval(interval int) AsyncOption {
	return func(opts *AsyncOptions) {
		opts.WriterLogInterval = interval
	}
}

func WithCanDropLog(drop bool) AsyncOption {
	return func(opts *AsyncOptions) {
		opts.CanDropLog = drop
	}
}

type AsyncRollWriter struct {
	logger io.Writer
	opts   *AsyncOptions

	logChan  chan []byte
	syncChan chan struct{}
}

// 封装一个异步写入的writer
func NewAsyncRollWriter(logger io.Writer, opt ...AsyncOption) *AsyncRollWriter {
	// 默认配置
	opts := &AsyncOptions{
		LogQueueSize:      1000,
		WriteLogSize:      2 * 1024,
		WriterLogInterval: 100,
	}

	for _, o := range opt {
		o(opts)
	}

	w := &AsyncRollWriter{}
	w.logger = logger
	w.opts = opts
	w.logChan = make(chan []byte, opts.LogQueueSize)
	w.syncChan = make(chan struct{})

	go w.batchWriteLog()

	return w
}

// 实现写文件的方法
func (w *AsyncRollWriter) Write(data []byte) (int, error) {
	fmt.Printf("[AsyncRollWriter]Write:%+v\n", string(data))
	log := make([]byte, len(data))

	copy(log, data)
	if w.opts.CanDropLog {
		select {
		case w.logChan <- log:
		default:
			return 0, errors.New("log is full, drop")
		}
	} else {
		w.logChan <- log
	}

	return len(data), nil
}

func (w *AsyncRollWriter) Sync() error {
	w.syncChan <- struct{}{}
	return nil
}

func (w *AsyncRollWriter) Close() error {
	return w.Sync()
}

func (w *AsyncRollWriter) batchWriteLog() {
	buffer := bytes.NewBuffer(make([]byte, 0, w.opts.WriteLogSize*2)) // 用来管理缓冲区，用于管理待写入的数据

	ticker := time.NewTicker(time.Millisecond * time.Duration(w.opts.WriterLogInterval)) // 用于定期刷新到日志的定时器

	for {
		select {
		case <-ticker.C: // 到了定期刷新的时间，则刷新一下
			if buffer.Len() > 0 {
				//fmt.Printf("[write log]buffer fresh\n")
				_, _ = w.logger.Write(buffer.Bytes())
				buffer.Reset()
			}
		case data := <-w.logChan: // 将日志从ch转移到buffer中
			buffer.Write(data)
			//fmt.Printf("[write log]buffer len\n")
			if buffer.Len() >= w.opts.WriteLogSize { // 如果超过了最大大小，则直接刷新
				_, _ = w.logger.Write(buffer.Bytes())
				buffer.Reset()
			}
		case <-w.syncChan:
			if buffer.Len() > 0 {
				_, _ = w.logger.Write(buffer.Bytes())
				buffer.Reset()
			}

			// todo
			// 是否需要强制刷新ch中的日志
		}
	}
}
