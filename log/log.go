package log

import (
	"fmt"
	"io"
	"log/syslog"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	TIME_HIPACHE_MODE = "02/Jan/2006:15:04:05 -0700"
)

var (
	ErrorLogger = NewWriterLogger(syncCloser{os.Stderr})

	fullBufferMsg = &LogEntry{
		Err: &ErrEntry{
			RawMsgs: []interface{}{"Dropping log messages to due to full channel buffer."},
		},
	}
)

type Logger struct {
	logChan    chan *LogEntry
	wg         *sync.WaitGroup
	writer     io.WriteCloser
	done       chan struct{}
	nextNotify <-chan time.Time
}

type ErrEntry struct {
	Path    string
	Rid     string
	Err     string
	Backend string
	Host    string
	RawMsgs []interface{}
}

type LogEntry struct {
	Now             time.Time
	BackendDuration time.Duration
	TotalDuration   time.Duration
	BackendKey      string
	RemoteAddr      string
	Method          string
	Path            string
	Proto           string
	Referer         string
	UserAgent       string
	RequestIDHeader string
	RequestID       string
	ForwardedFor    string
	StatusCode      int
	ContentLength   int64
	Err             *ErrEntry
}

func NewFileLogger(path string) (*Logger, error) {
	if path == "syslog" {
		return NewSyslogLogger()
	} else if path == "stdout" {
		return NewStdoutLogger()
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		return nil, err
	}
	return NewWriterLogger(file), nil
}

func NewSyslogLogger() (*Logger, error) {
	writer, err := syslog.New(syslog.LOG_INFO|syslog.LOG_LOCAL0, "hipache")
	if err != nil {
		return nil, err
	}
	return NewWriterLogger(writer), nil
}

type syncCloser struct{ *os.File }

func (n syncCloser) Close() error { return n.Sync() }

func NewStdoutLogger() (*Logger, error) {
	return NewWriterLogger(syncCloser{os.Stdout}), nil
}

func NewWriterLogger(writer io.WriteCloser) *Logger {
	l := Logger{
		logChan:    make(chan *LogEntry, 10000),
		done:       make(chan struct{}),
		wg:         &sync.WaitGroup{},
		writer:     writer,
		nextNotify: time.After(0),
	}
	l.wg.Add(1)
	go l.logWriter()
	return &l
}

func (l *Logger) Print(msgs ...interface{}) {
	l.MessageRaw(&LogEntry{Err: &ErrEntry{RawMsgs: msgs}})
}

func (l *Logger) MessageRaw(entry *LogEntry) {
	select {
	case l.logChan <- entry:
	default:
		select {
		case <-l.nextNotify:
			l.wg.Add(1)
			go func() {
				defer l.wg.Done()
				l.logChan <- fullBufferMsg
			}()
			l.nextNotify = time.After(time.Minute)
		default:
		}
	}
}

func (l *Logger) Stop() {
	l.wg.Done()
	l.wg.Wait()
	l.logChan <- nil
	<-l.done
}

func (l *Logger) logWriter() {
	defer close(l.done)
	defer l.writer.Close()
	for el := range l.logChan {
		if el == nil {
			return
		}
		if el.Err != nil {
			if len(el.Err.RawMsgs) > 0 {
				fmt.Fprintln(l.writer, el.Err.RawMsgs...)
				continue
			}
			backend := el.Err.Backend
			if backend == "" {
				backend = "?"
			}
			fmt.Fprintf(
				l.writer,
				"ERROR in %s -> %s - %s - %s - %s\n",
				el.Err.Host,
				backend,
				el.Err.Path,
				el.Err.Rid,
				el.Err.Err,
			)
			continue
		}
		nowFormatted := el.Now.Format(TIME_HIPACHE_MODE)
		ip, _, _ := net.SplitHostPort(el.RemoteAddr)
		if ip == "" {
			ip = el.RemoteAddr
		}
		if !strings.HasPrefix(ip, "::") {
			ip = "::ffff:" + ip
		}
		fmt.Fprintf(l.writer,
			"%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" \"%s:%s\" \"%s\" \"%s\" %0.3f %0.3f\n",
			ip,
			nowFormatted,
			el.Method,
			el.Path,
			el.Proto,
			el.StatusCode,
			el.ContentLength,
			el.Referer,
			el.UserAgent,
			el.RequestIDHeader,
			el.RequestID,
			el.BackendKey,
			el.ForwardedFor,
			float64(el.TotalDuration)/float64(time.Second),
			float64(el.BackendDuration)/float64(time.Second),
		)
	}
}
