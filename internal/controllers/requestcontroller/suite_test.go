/*
Copyright 2022 Matt Wise.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package requestcontroller

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/diranged/oz/internal/api/v1alpha1"
	crdsv1alpha1 "github.com/diranged/oz/internal/api/v1alpha1"
	"github.com/diranged/oz/internal/builders"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite / Request Controller")
}

var _ = BeforeSuite(func() {
	logger := zap.New(
		zap.WriteTo(GinkgoWriter),
		zap.UseDevMode(true),
		zap.Level(zapcore.DebugLevel),
	)
	logf.SetLogger(logger)

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = crdsv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

// mockBuilder is used to test our reconciler logic without using real resource
// builders - this lets us test failures and really focus only on the
// reconciler logic itself.
type mockBuilder struct {
	getTemplateResp v1alpha1.ITemplateResource
	getTemplateErr  error

	getDurationResp time.Duration
	getDurationErr  error

	setOwnerReferenceErr error

	createResourcesResp string
	createResourcesErr  error

	accessResourcesAreReadyResp bool
	accessResourcesAreReadyErr  error
}

// https://stackoverflow.com/questions/33089523/how-to-mark-golang-struct-as-implementing-interface
var (
	_ builders.IBuilder = &mockBuilder{}
	_ builders.IBuilder = (*mockBuilder)(nil)
)

func (b *mockBuilder) GetTemplate(
	_ context.Context,
	_ client.Client,
	_ v1alpha1.IRequestResource,
) (v1alpha1.ITemplateResource, error) {
	return b.getTemplateResp, b.getTemplateErr
}

func (b *mockBuilder) GetAccessDuration(
	_ v1alpha1.IRequestResource,
	_ v1alpha1.ITemplateResource,
) (time.Duration, string, error) {
	return b.getDurationResp, "test", b.getDurationErr
}

func (b *mockBuilder) SetRequestOwnerReference(
	_ context.Context,
	_ client.Client,
	_ v1alpha1.IRequestResource,
	_ v1alpha1.ITemplateResource,
) error {
	return b.setOwnerReferenceErr
}

func (b *mockBuilder) CreateAccessResources(
	_ context.Context,
	_ client.Client,
	_ v1alpha1.IRequestResource,
	_ v1alpha1.ITemplateResource,
) (string, error) {
	return b.createResourcesResp, b.createResourcesErr
}

func (b *mockBuilder) AccessResourcesAreReady(
	_ context.Context,
	_ client.Client,
	_ v1alpha1.IRequestResource,
	_ v1alpha1.ITemplateResource,
) (bool, error) {
	return b.accessResourcesAreReadyResp, b.accessResourcesAreReadyErr
}
