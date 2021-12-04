package rollwriter

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lestrrat-go/strftime"
)

var _ io.WriteCloser = (*RollWriter)(nil)

var reopenFileTime int64 = 10 // 10s为单位

type Options struct {
	MaxSize    int64  // 日志文件最大大小
	MaxHistory int    // 保留的最大文件数
	MaxDay     int    // 日志最大保留时间
	IfCompress bool   // 日志文件是否压缩
	TimeFormat string // 按时间分割文件的时间格式
}

type Option func(*Options)

func WithMaxSize(size int64) Option {
	return func(opt *Options) {
		opt.MaxSize = size * 1024 * 1024
	}
}

func WithMaxDay(day int) Option {
	return func(opt *Options) {
		opt.MaxDay = day
	}
}

func WithMaxHistory(n int) Option {
	return func(opt *Options) {
		opt.MaxHistory = n
	}
}

func WithCompress(c bool) Option {
	return func(opt *Options) {
		opt.IfCompress = c
	}
}

func WithTimeFormat(s string) Option {
	return func(opt *Options) {
		opt.TimeFormat = s
	}
}

type RollWriter struct {
	filePath string   // 文件路径
	opts     *Options // 配置

	pattern  *strftime.Strftime // 文件模式
	currDir  string
	currPath string
	currSize int64
	currFile atomic.Value

	openTime int64

	mu       sync.Mutex
	once     sync.Once
	closeCh  chan *os.File // 待关闭的文件句柄
	notifyCh chan bool     // 触发日志清理
}

// 获取当前日志句柄
func (w *RollWriter) getCurrFile() *os.File {
	if file, ok := w.currFile.Load().(*os.File); ok {
		return file
	}
	return nil
}

// 设置当前日志句柄
func (w *RollWriter) setCurrFile(file *os.File) {
	w.currFile.Store(file)
}

func dirExist(dir string) bool {
	_, err := os.Stat(dir)
	return err == nil || os.IsExist(err)
}

func NewRollWriter(filePath string, opt ...Option) (*RollWriter, error) {
	opts := &Options{
		MaxSize:    0,
		MaxDay:     0,
		MaxHistory: 0,
		IfCompress: false,
	}

	for _, o := range opt {
		o(opts)
	}

	if filePath == "" {
		return nil, errors.New("no file path")
	}

	pattern, err := strftime.New(filePath + opts.TimeFormat)
	if err != nil {
		return nil, err
	}

	write := &RollWriter{
		filePath: filePath,
		opts:     opts,
		pattern:  pattern,
		currDir:  filepath.Dir(filePath),
	}

	// 如果文件已经存在，则直接返回成功，否则需要创建文件
	if dirExist(write.currDir) {
		return write, nil
	}
	err = os.Mkdir(write.currDir, 0755)
	if err != nil {
		return nil, err
	}

	return write, nil
}

func (w *RollWriter) doReopenFile(path string) error {
	atomic.StoreInt64(&w.openTime, time.Now().Unix())

	lastFile := w.getCurrFile()

	// 打开新的文件，如果不存在则创建
	curFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err == nil {
		w.setCurrFile(curFile)

		if lastFile != nil {
			// 需要延迟关闭
			w.closeCh <- lastFile
		}

		st, _ := os.Stat(path)
		if st != nil {
			atomic.StoreInt64(&w.currSize, st.Size())
		}
	}

	return err
}

// 定期重新打开文件
func (w *RollWriter) reopenFile() {
	if w.getCurrFile() == nil || time.Now().Unix()-atomic.LoadInt64(&w.openTime) > 10 {
		now := time.Now()
		atomic.StoreInt64(&w.openTime, now.Unix())
		currPath := w.pattern.FormatString(time.Now())
		if w.currPath != currPath { // 如果文件已经更新，
			w.currPath = currPath
			w.notifyClose()
		}

		_ = w.doReopenFile(w.currPath)
	}
}

func (w *RollWriter) Write(v []byte) (n int, err error) {
	// 每隔10s重新打开一次文件
	if w.getCurrFile() == nil || time.Now().Unix()-atomic.LoadInt64(&w.openTime) > 10 {
		w.mu.Lock()
		w.reopenFile()
		w.mu.Unlock()
	}

	// 获取当前文件句柄
	if w.getCurrFile() == nil {
		return 0, errors.New("curr file not exist")
	}

	// 写文件
	n, err = w.getCurrFile().Write(v)
	atomic.AddInt64(&w.currSize, int64(n))

	// 如果设置最大文件大小，则另开文件存储
	// 如果上面是err也会触发检查
	if w.opts.MaxSize > 0 && atomic.LoadInt64(&w.currSize) >= w.opts.MaxSize {
		w.mu.Lock()
		w.backupFile()
		w.mu.Unlock()
	}

	return n, err
}

func (w *RollWriter) Close() error {
	if w.getCurrFile() == nil {
		return nil
	}

	err := w.getCurrFile().Close()
	w.setCurrFile(nil)
	return err
}

// 超过大小重命名文件
func (w *RollWriter) backupFile() {
	if w.opts.MaxSize > 0 && atomic.LoadInt64(&w.currSize) >= w.opts.MaxSize {
		atomic.StoreInt64(&w.currSize, 0) // todo  为啥设置当前size为0不放在重开文件后

		// 修改老文件名字
		newFileName := w.currPath + "." + time.Now().Format("bk-20060102-150405.00000")
		if _, e := os.Stat(w.currPath); !os.IsNotExist(e) { // todo 为啥需要先查文件stat
			_ = os.Rename(w.currPath, newFileName)
		}

		// 重新开新文件
		_ = w.doReopenFile(w.currPath)
		w.notifyClose()
	}

}

// 删除已经关闭的文件句柄
// 删除已经过期的文件
func (w *RollWriter) notifyClose() {
	w.once.Do(func() {
		w.notifyCh = make(chan bool, 1)
		w.closeCh = make(chan *os.File, 100)

		go w.cleanClosedFile()
		go w.cleanExpireFile()
	})

	select {
	case w.notifyCh <- true:
	default:
	}
}

func (w *RollWriter) cleanClosedFile() {
	for f := range w.closeCh {
		time.Sleep(30 * time.Millisecond)
		f.Close()
	}
}

func (w *RollWriter) cleanExpireFile() {
	for _ = range w.notifyCh {
		if w.opts.MaxHistory == 0 && w.opts.MaxDay == 0 { //这种情况下不清理历史日志
			continue
		}

		w.expireFile()
	}
}

func (w *RollWriter) expireFile() {
	oldFiles, getErr := w.getDirHistory()

	if getErr != nil || len(oldFiles) == 0 {
		return
	}

	var remove []logWithT

	oldFiles = expireWithMaxHistory(oldFiles, &remove, w.opts.MaxHistory)

	oldFiles = expireWithDay(oldFiles, &remove, w.opts.MaxDay)

	w.removeFile(remove)

}

func (w *RollWriter) removeFile(remove []logWithT) {
	for _, f := range remove {
		os.Remove(filepath.Join(w.currDir, f.Name()))
	}
}

// 超过最大日志个数后，即删除
func expireWithMaxHistory(files []logWithT, remove *[]logWithT, maxHistory int) []logWithT {
	if maxHistory == 0 || len(files) < maxHistory { // 如果全部文件数低于最大文件数，则直接返回
		return files
	}

	var remain []logWithT
	preserved := make(map[string]bool)
	for _, f := range files {
		preserved[f.Name()] = true
		if len(preserved) > maxHistory {
			*remove = append(*remove, f)
		} else {
			remain = append(remain, f)
		}
	}

	return remain
}

// 超过最大日期后，即删除
func expireWithDay(files []logWithT, remove *[]logWithT, maxDay int) []logWithT {
	if maxDay == 0 {
		return files
	}

	var remain []logWithT
	detTs := time.Now().Add(-1 * time.Duration(int64(24*time.Hour)*int64(maxDay)))
	for _, f := range files {
		if f.modTime.Before(detTs) {
			*remove = append(*remove, f)
		} else {
			remain = append(remain, f)
		}
	}

	return remain
}

// 查找目录下与当前文件匹配的文件
func (w *RollWriter) getDirHistory() ([]logWithT, error) {
	files, err := ioutil.ReadDir(w.currDir)

	if err != nil {
		return nil, fmt.Errorf("can not read dir files:%+v", err)
	}

	logWithTs := make([]logWithT, 0)
	fileName := filepath.Base(w.filePath) // 文件名

	for _, f := range files { // 遍历目录下的文件，找到和当前文件匹配的文件
		if f.IsDir() {
			continue
		}

		if filepath.Base(w.currPath) == f.Name() { // 和当前路径一样，则直接过滤
			continue
		}

		if !strings.HasPrefix(f.Name(), fileName) { // 如果不是同样的前缀，则说明不是相同log生成的文件
			continue
		}

		st, _ := os.Stat(filepath.Join(w.currDir, fileName))
		if st == nil { // 取不到文件属性
			continue
		}

		logWithTs = append(logWithTs, logWithT{
			modTime:  st.ModTime(),
			FileInfo: f,
		})
	}

	sort.Sort(byModTimeLogInfo(logWithTs)) // 对匹配当前文件的历史文件按时间排序
	return logWithTs, nil
}

type logWithT struct {
	modTime time.Time
	os.FileInfo
}

type byModTimeLogInfo []logWithT

func (b byModTimeLogInfo) Less(i, j int) bool {
	return b[i].modTime.After(b[j].modTime)
}

func (b byModTimeLogInfo) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byModTimeLogInfo) Len() int {
	return len(b)
}
