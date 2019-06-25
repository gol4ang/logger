package logger_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gol4ng/logger/formatter"
	"github.com/gol4ng/logger/handler"
	"github.com/gol4ng/logger/middleware"
	"log/syslog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gol4ng/logger"
	"github.com/stretchr/testify/assert"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    logger.Level
		expected string
	}{
		{level: logger.DebugLevel, expected: "debug"},
		{level: logger.InfoLevel, expected: "info"},
		{level: logger.NoticeLevel, expected: "notice"},
		{level: logger.WarningLevel, expected: "warning"},
		{level: logger.ErrorLevel, expected: "error"},
		{level: logger.CriticalLevel, expected: "critical"},
		{level: logger.AlertLevel, expected: "alert"},
		{level: logger.EmergencyLevel, expected: "emergency"},
		{level: 123, expected: "level(123)"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.level.String())
	}
}

func TestLogger_Log(t *testing.T) {
	tests := []struct {
		name  string
		level logger.Level
	}{
		{
			name:  "test Debug(...)",
			level: logger.DebugLevel,
		},
		{
			name:  "test Info(...)",
			level: logger.InfoLevel,
		},
		{
			name:  "test Notice(...)",
			level: logger.NoticeLevel,
		},
		{
			name:  "test Warning(...)",
			level: logger.WarningLevel,
		},
		{
			name:  "test Error(...)",
			level: logger.ErrorLevel,
		},
		{
			name:  "test Critical(...)",
			level: logger.CriticalLevel,
		},
		{
			name:  "test Alert(...)",
			level: logger.AlertLevel,
		},
		{
			name:  "test Emergency(...)",
			level: logger.EmergencyLevel,
		},
		{
			name:  "test Log(custom level)",
			level: logger.Level(127),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := logger.NewLogger(func(entry logger.Entry) error {
				assert.Equal(t, "l message", entry.Message)
				assert.Equal(t, tt.level, entry.Level)
				assert.Nil(t, entry.Context)

				return nil
			})

			var err error
			switch tt.level {
			case logger.DebugLevel:
				err = l.Debug("l message", nil)
				break
			case logger.InfoLevel:
				err = l.Info("l message", nil)
				break
			case logger.NoticeLevel:
				err = l.Notice("l message", nil)
				break
			case logger.WarningLevel:
				err = l.Warning("l message", nil)
				break
			case logger.ErrorLevel:
				err = l.Error("l message", nil)
				break
			case logger.CriticalLevel:
				err = l.Critical("l message", nil)
				break
			case logger.AlertLevel:
				err = l.Alert("l message", nil)
				break
			case logger.EmergencyLevel:
				err = l.Emergency("l message", nil)
				break
			default:
				err = l.Log("l message", tt.level, nil)
			}
			assert.Nil(t, err)
		})
	}
}

func TestNewLogger_WillPanic(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			assert.Equal(t, err, errors.New("handler must not be <nil>"))
		}
	}()
	logger.NewLogger(nil)
}

func TestNewLogger_LogWithError(t *testing.T) {
	err := errors.New("my error")
	l := logger.NewLogger(func(entry logger.Entry) error {
		return err
	})

	assert.Equal(t, err, l.Debug("l message", nil))
	assert.Equal(t, err, l.Info("l message", nil))
	assert.Equal(t, err, l.Notice("l message", nil))
	assert.Equal(t, err, l.Warning("l message", nil))
	assert.Equal(t, err, l.Error("l message", nil))
	assert.Equal(t, err, l.Critical("l message", nil))
	assert.Equal(t, err, l.Alert("l message", nil))
	assert.Equal(t, err, l.Emergency("l message", nil))
	assert.Equal(t, err, l.Log("l message", logger.Level(127), nil))
}

func TestNewNilLogger_Log(t *testing.T) {
	l := logger.NewNopLogger()

	assert.Nil(t, l.Debug("l message", nil))
	assert.Nil(t, l.Info("l message", nil))
	assert.Nil(t, l.Notice("l message", nil))
	assert.Nil(t, l.Warning("l message", nil))
	assert.Nil(t, l.Error("l message", nil))
	assert.Nil(t, l.Critical("l message", nil))
	assert.Nil(t, l.Alert("l message", nil))
	assert.Nil(t, l.Emergency("l message", nil))
	assert.Nil(t, l.Log("l message", logger.Level(127), nil))
}

func TestNewLogger_Wrap(t *testing.T) {
	mockHandlerCalled := false
	mockHandler := func(entry logger.Entry) error {
		mockHandlerCalled = true
		assert.Equal(t, "l message", entry.Message)
		assert.Equal(t, logger.DebugLevel, entry.Level)
		assert.Nil(t, entry.Context)

		return nil
	}

	l := logger.NewLogger(mockHandler)

	mockHandlerWrapperCalled := false
	mockHandlerWrapper := func(entry logger.Entry) error {
		mockHandlerWrapperCalled = true
		return mockHandler(entry)
	}

	l.Wrap(func(h logger.HandlerInterface) logger.HandlerInterface {
		return mockHandlerWrapper
	})

	assert.Nil(t, l.Debug("l message", nil))

	assert.True(t, mockHandlerCalled)
	assert.True(t, mockHandlerWrapperCalled)
}

func TestNewLogger_WrapNew(t *testing.T) {
	mockHandlerCalled := false
	mockHandler := func(entry logger.Entry) error {
		mockHandlerCalled = true
		assert.Equal(t, "l message", entry.Message)
		assert.Equal(t, logger.DebugLevel, entry.Level)
		assert.Nil(t, entry.Context)

		return nil
	}

	l := logger.NewLogger(mockHandler)

	mockHandlerWrapperCalled := false
	mockHandlerWrapper := func(entry logger.Entry) error {
		mockHandlerWrapperCalled = true
		return mockHandler(entry)
	}

	l2 := l.WrapNew(func(h logger.HandlerInterface) logger.HandlerInterface {
		return mockHandlerWrapper
	})

	assert.Nil(t, l.Debug("l message", nil))
	assert.True(t, mockHandlerCalled)
	assert.False(t, mockHandlerWrapperCalled)

	mockHandlerCalled = false
	mockHandlerWrapperCalled = false

	assert.Nil(t, l2.Debug("l message", nil))
	assert.True(t, mockHandlerCalled)
	assert.True(t, mockHandlerWrapperCalled)
}

/////////////////////
// Examples
/////////////////////

func ExampleLogger_callerHandler() {
	output := &Output{}
	myLogger := logger.NewLogger(
		middleware.Caller(3)(
			handler.Stream(output, formatter.NewLine("lvl: %[2]s | msg: %[1]s | ctx: %[3]v")),
		),
	)

	myLogger.Debug("Log example", nil)
	myLogger.Info("Log example", nil)
	myLogger.Notice("Log example", nil)
	myLogger.Warning("Log example", nil)
	myLogger.Error("Log example", nil)
	myLogger.Critical("Log example", nil)
	myLogger.Alert("Log example", nil)
	myLogger.Emergency("Log example", nil)

	output.Constains([]string{
		"lvl: debug | msg: Log example | ctx:", "<file:/", "<line:",
		"lvl: info | msg: Log example | ctx:", "<file:/", "<line:",
		"lvl: notice | msg: Log example | ctx:", "<file:/", "<line:",
		"lvl: warning | msg: Log example | ctx:", "<file:/", "<line:",
		"lvl: error | msg: Log example | ctx:", "<file:/", "<line:",
		"lvl: critical | msg: Log example | ctx:", "<file:/", "<line:",
		"lvl: alert | msg: Log example | ctx:", "<file:/", "<line:",
		"lvl: emergency | msg: Log example | ctx:", "<file:/", "<line:",
	})

	//Output:
}

func ExampleLogger_lineFormatter() {
	output := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(output, formatter.NewLine("lvl: %[2]s | msg: %[1]s | ctx: %[3]v")),
	)

	myLogger.Debug("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Info("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Notice("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Warning("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Error("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Critical("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Alert("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Emergency("Log example", logger.Ctx("my_key", "my_value"))

	output.Constains([]string{
		"lvl: debug | msg: Log example | ctx: <my_key:my_value>",
		"lvl: info | msg: Log example | ctx: <my_key:my_value>",
		"lvl: notice | msg: Log example | ctx: <my_key:my_value>",
		"lvl: warning | msg: Log example | ctx: <my_key:my_value>",
		"lvl: error | msg: Log example | ctx: <my_key:my_value>",
		"lvl: critical | msg: Log example | ctx: <my_key:my_value>",
		"lvl: alert | msg: Log example | ctx: <my_key:my_value>",
		"lvl: emergency | msg: Log example | ctx: <my_key:my_value>",
	})

	//Output:
}

func ExampleLogger_jsonFormatter() {
	output := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(output, formatter.NewJSONEncoder()),
	)

	myLogger.Debug("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Info("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Notice("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Warning("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Error("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Critical("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Alert("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Emergency("Log example", logger.Ctx("my_key", "my_value"))

	output.Constains([]string{
		`{"Message":"Log example","Level":7,"Context":{"my_key":"my_value"}}`,
		`{"Message":"Log example","Level":6,"Context":{"my_key":"my_value"}}`,
		`{"Message":"Log example","Level":5,"Context":{"my_key":"my_value"}}`,
		`{"Message":"Log example","Level":4,"Context":{"my_key":"my_value"}}`,
		`{"Message":"Log example","Level":3,"Context":{"my_key":"my_value"}}`,
		`{"Message":"Log example","Level":2,"Context":{"my_key":"my_value"}}`,
		`{"Message":"Log example","Level":1,"Context":{"my_key":"my_value"}}`,
		`{"Message":"Log example","Level":0,"Context":{"my_key":"my_value"}}`,
	})

	//Output:
}

func ExampleLogger_minLevelFilterHandler() {
	output := &Output{}
	myLogger := logger.NewLogger(
		middleware.MinLevelFilter(logger.WarningLevel)(
			handler.Stream(output, formatter.NewDefaultFormatter()),
		),
	)

	myLogger.Debug("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Info("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Notice("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Warning("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Error("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Critical("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Alert("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Emergency("Log example", logger.Ctx("my_key", "my_value"))

	output.Constains([]string{
		`<warning> Log example {"my_key":"my_value"}`,
		`<error> Log example {"my_key":"my_value"}`,
		`<critical> Log example {"my_key":"my_value"}`,
		`<alert> Log example {"my_key":"my_value"}`,
		`<emergency> Log example {"my_key":"my_value"}`,
	})

	//Output:
}

func ExampleLogger_groupHandler() {
	output := &Output{}
	output2 := &Output{}
	myLogger := logger.NewLogger(
		handler.Group(
			handler.Stream(output, formatter.NewJSONEncoder()),
			handler.Stream(output2, formatter.NewDefaultFormatter()),
		),
	)

	myLogger.Debug("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Info("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Notice("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Warning("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Error("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Critical("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Alert("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Emergency("Log example", logger.Ctx("my_key", "my_value"))

	output.Constains([]string{
		`{"Message":"Log example","Level":7,"Context":{"my_key":"my_value"}}`,
		`{"Message":"Log example","Level":6,"Context":{"my_key":"my_value"}}`,
		`{"Message":"Log example","Level":5,"Context":{"my_key":"my_value"}}`,
		`{"Message":"Log example","Level":4,"Context":{"my_key":"my_value"}}`,
		`{"Message":"Log example","Level":3,"Context":{"my_key":"my_value"}}`,
		`{"Message":"Log example","Level":2,"Context":{"my_key":"my_value"}}`,
		`{"Message":"Log example","Level":1,"Context":{"my_key":"my_value"}}`,
		`{"Message":"Log example","Level":0,"Context":{"my_key":"my_value"}}`,
	})

	output2.Constains([]string{
		`<debug> Log example {"my_key":"my_value"}`,
		`<info> Log example {"my_key":"my_value"}`,
		`<notice> Log example {"my_key":"my_value"}`,
		`<warning> Log example {"my_key":"my_value"}`,
		`<error> Log example {"my_key":"my_value"}`,
		`<critical> Log example {"my_key":"my_value"}`,
		`<alert> Log example {"my_key":"my_value"}`,
		`<emergency> Log example {"my_key":"my_value"}`,
	})

	//Output:
}

func ExampleLogger_wrapHandler() {
	output := &Output{}
	myLogger := logger.NewLogger(
		handler.Stream(output, formatter.NewDefaultFormatter()),
	)
	myLogger.Wrap(middleware.MinLevelFilter(logger.WarningLevel))

	myLogger.Debug("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Info("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Notice("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Warning("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Error("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Critical("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Alert("Log example", logger.Ctx("my_key", "my_value"))
	myLogger.Emergency("Log example", logger.Ctx("my_key", "my_value"))
	output.Constains([]string{
		`<warning> Log example {"my_key":"my_value"}`,
		`<error> Log example {"my_key":"my_value"}`,
		`<critical> Log example {"my_key":"my_value"}`,
		`<alert> Log example {"my_key":"my_value"}`,
		`<emergency> Log example {"my_key":"my_value"}`,
	})

	//Output:
}

func ExampleLogger_timeRotateHandler() {
	lineFormatter := formatter.NewLine("lvl: %[2]s | msg: %[1]s | ctx: %[3]v")

	rotateLogHandler, _ := handler.TimeRotateFileStream(os.TempDir()+"%s.log", time.Stamp, lineFormatter, 1*time.Second)
	myLogger := logger.NewLogger(rotateLogHandler)

	myLogger.Debug("Log example", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Info("Log example", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Warning("Log example", logger.Ctx("ctx_key", "ctx_value"))
	time.Sleep(1 * time.Second)
	myLogger.Error("Log example", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Alert("Log example", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Critical("Log example", logger.Ctx("ctx_key", "ctx_value"))

	//Output:
}

func ExampleLogger_logRotateHandler() {
	lineFormatter := formatter.NewLine("lvl: %[2]s | msg: %[1]s | ctx: %[3]v")

	rotateLogHandler, _ := handler.LogRotateFileStream("test", os.TempDir()+"%s.log", time.Stamp, lineFormatter, 1*time.Second)
	myLogger := logger.NewLogger(rotateLogHandler)

	myLogger.Debug("Log example", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Info("Log example", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Warning("Log example", logger.Ctx("ctx_key", "ctx_value"))
	time.Sleep(1 * time.Second)
	myLogger.Error("Log example", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Alert("Log example", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Critical("Log example", logger.Ctx("ctx_key", "ctx_value"))

	//Output:
}

// You can run the command below to show syslog messages
// syslog -F '$Time $Host $(Sender)[$(PID)] <$((Level)(str))>: $Message'
//Apr 26 12:22:06 hades my_go_logger[69302] <Notice>: <notice> Log example2 {"ctx_key":"ctx_value"}
//Apr 26 12:22:06 hades my_go_logger[69302] <Warning>: <warning> Log example3 {"ctx_key":"ctx_value"}
//Apr 26 12:22:06 hades my_go_logger[69302] <Error>: <error> Log example4 {"ctx_key":"ctx_value"}
//Apr 26 12:22:06 hades my_go_logger[69302] <Critical>: <critical> Log example5 {"ctx_key":"ctx_value"}
//Apr 26 12:22:06 hades my_go_logger[69302] <Alert>: <alert> Log example6 {"ctx_key":"ctx_value"}
//Apr 26 12:22:06 hades my_go_logger[69302] <Emergency>: <emergency> Log example7 {"ctx_key":"ctx_value"}
func ExampleLogger_syslogHandler() {
	syslogHandler, _ := handler.Syslog(
		formatter.NewDefaultFormatter(),
		"",
		"",
		syslog.LOG_DEBUG,
		"my_go_logger")
	myLogger := logger.NewLogger(syslogHandler)

	myLogger.Debug("Log example", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Info("Log example1", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Notice("Log example2", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Warning("Log example3", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Error("Log example4", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Critical("Log example5", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Alert("Log example6", logger.Ctx("ctx_key", "ctx_value"))
	myLogger.Emergency("Log example7", logger.Ctx("ctx_key", "ctx_value"))

	//Output:
}

type Output struct {
	bytes.Buffer
}

func (o *Output) Constains(str []string) {
	b := o.String()
	for _, s := range str {
		if strings.Contains(b, s) != true {
			fmt.Printf("buffer %s must contain %s\n", b, s)
		}
	}
}
