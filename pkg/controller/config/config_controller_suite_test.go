/*

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

package config

import (
	"context"
	stdlog "log"
	"os"
	"path/filepath"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/open-policy-agent/gatekeeper/apis"
)

var cfg *rest.Config

func TestMain(m *testing.M) {
	t := &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}
	if err := apis.AddToScheme(scheme.Scheme); err != nil {
		stdlog.Fatal(err)
	}

	var err error
	if cfg, err = t.Start(); err != nil {
		stdlog.Fatal(err)
	}

	if err := createGatekeeperNamespace(cfg); err != nil {
		stdlog.Printf("creating namespace: %v", err)
	}

	code := m.Run()
	if err = t.Stop(); err != nil {
		stdlog.Printf("error while trying to stop server: %v", err)
	}
	os.Exit(code)
}

// SetupTestReconcile returns a reconcile.Reconcile implementation that delegates to inner and
// writes the request to requests after Reconcile is finished.
func SetupTestReconcile(inner reconcile.Reconciler) (reconcile.Reconciler, chan reconcile.Request) {
	requests := make(chan reconcile.Request)
	fn := reconcile.Func(func(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
		result, err := inner.Reconcile(ctx, req)
		requests <- req
		return result, err
	})
	return fn, requests
}

// Bootstrap the gatekeeper-system namespace for use in tests.
func createGatekeeperNamespace(cfg *rest.Config) error {
	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return err
	}

	// Create gatekeeper namespace
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gatekeeper-system",
		},
	}

	ctx := context.Background()
	_, err = controllerutil.CreateOrUpdate(ctx, c, ns, func() error { return nil })
	return err
}
