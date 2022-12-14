package webhook

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestWebhookControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhook Suite")
}

var _ = BeforeSuite(func() {
	logger := zap.New(
		zap.WriteTo(GinkgoWriter),
		zap.UseDevMode(true),
		zap.Level(zapcore.DebugLevel),
	)
	logf.SetLogger(logger)
})
