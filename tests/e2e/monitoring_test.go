/*
Copyright © contributors to CloudNativePG, established as
CloudNativePG a Series of LF Projects, LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0
*/

package e2e

import (
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/types"
	k8client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cloudnative-pg/cloudnative-pg/pkg/utils"
	"github.com/cloudnative-pg/cloudnative-pg/tests"
	"github.com/cloudnative-pg/cloudnative-pg/tests/utils/objects"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Set of tests that set up a cluster with monitoring support enabled
var _ = Describe("PodMonitor support", Serial, Label(tests.LabelObservability), func() {
	getPodMonitorFunc := func(
		ctx context.Context,
		crudClient k8client.Client,
		namespace, name string,
	) (*monitoringv1.PodMonitor, error) {
		podMonitor := &monitoringv1.PodMonitor{}
		namespacedName := types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}

		err := objects.Get(ctx, crudClient, namespacedName, podMonitor)
		if err != nil {
			return nil, err
		}
		return podMonitor, nil
	}

	const (
		namespacePrefix              = "cluster-monitoring-e2e"
		level                        = tests.Medium
		clusterDefaultName           = "cluster-default-monitoring"
		clusterDefaultMonitoringFile = fixturesDir + "/monitoring/cluster-default-monitoring.yaml"
	)
	var err error
	var namespace string

	BeforeEach(func() {
		if testLevelEnv.Depth < int(level) {
			Skip("Test depth is lower than the amount requested for this test")
		}

		if !IsLocal() {
			Skip("PodMonitor test only runs on Local deployment")
		}
	})

	It("requires existence of the PodMonitor CRD", func() {
		// Check if CRD exists, otherwise test is invalid
		exist, err := utils.PodMonitorExist(env.APIExtensionClient.Discovery())
		Expect(err).ToNot(HaveOccurred())
		Expect(exist).To(BeTrue())
	})

	It("sets up a cluster enabling PodMonitor feature", func() {
		namespace, err = env.CreateUniqueTestNamespace(env.Ctx, env.Client, namespacePrefix)
		Expect(err).ToNot(HaveOccurred())

		AssertCreateCluster(namespace, clusterDefaultName, clusterDefaultMonitoringFile, env)

		By("verifying PodMonitor existence", func() {
			podMonitor, err := getPodMonitorFunc(env.Ctx, env.Client, namespace, clusterDefaultName)
			Expect(err).ToNot(HaveOccurred())

			endpoints := podMonitor.Spec.PodMetricsEndpoints
			Expect(endpoints).Should(HaveLen(1), "endpoints should be of size 1")
			Expect(endpoints[0].Interval).Should(BeEmpty(), "should not be set as spec")
			Expect(endpoints[0].ScrapeTimeout).Should(BeEmpty(), "should not be set as spec")
		})
	})
})
