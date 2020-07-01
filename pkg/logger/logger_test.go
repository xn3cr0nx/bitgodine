package logger_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/xn3cr0nx/bitgodine/pkg/logger"
)

var _ = Describe("Testing with Ginkgo", func() {
	It("logger", func() {

		logger.Setup()
		hook := test.NewLocal(logger.Log)

		logger.Info("test", "testing Info function", logger.Params{"test": "Params"})
		logger.Warn("test", "testing Warn function", logger.Params{"test": "Params"})
		logger.Debug("test", "testing Debug function", logger.Params{"test": "Params"})
		logger.Error("test", errors.New("testing Error function"), logger.Params{"error": "error"})

		assert.Equal(GinkgoT(), 3, len(hook.Entries))
		assert.Equal(GinkgoT(), string(logrus.InfoLevel), string(hook.Entries[0].Level))
		assert.Equal(GinkgoT(), string(logrus.WarnLevel), string(hook.Entries[1].Level))
		assert.Equal(GinkgoT(), string(logrus.ErrorLevel), string(hook.LastEntry().Level))
		assert.Equal(GinkgoT(), "testing Error function", hook.LastEntry().Message)

		hook.Reset()
		assert.Nil(GinkgoT(), hook.LastEntry())
	})
})
