package logger

import (
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {

	Setup()
	hook := test.NewLocal(Log)

	Info("test", "testing Info function", Params{"test": "Params"})
	Warn("test", "testing Warn function", Params{"test": "Params"})
	Debug("test", "testing Debug function", Params{"test": "Params"})
	Error("test", errors.New("testing Error function"), Params{"error": "error"})

	assert.Equal(t, 3, len(hook.Entries))
	assert.Equal(t, string(logrus.InfoLevel), string(hook.Entries[0].Level))
	assert.Equal(t, string(logrus.WarnLevel), string(hook.Entries[1].Level))
	assert.Equal(t, string(logrus.ErrorLevel), string(hook.LastEntry().Level))
	assert.Equal(t, "testing Error function", hook.LastEntry().Message)

	hook.Reset()
	assert.Nil(t, hook.LastEntry())

}
