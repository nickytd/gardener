// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package garden

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	operatorclient "github.com/gardener/gardener/pkg/operator/client"
	"github.com/gardener/gardener/pkg/utils"
	gardenerutils "github.com/gardener/gardener/pkg/utils/gardener"
	"github.com/gardener/gardener/test/e2e/operator/garden/internal/rotation"
	rotationutils "github.com/gardener/gardener/test/utils/rotation"
)

var _ = Describe("Garden Tests", Label("Garden", "default"), func() {
	var (
		backupSecret = defaultBackupSecret()
		garden       = defaultGarden(backupSecret)
	)

	It("Create Garden, Rotate Credentials and Delete Garden", Label("credentials-rotation"), func() {
		By("Create Garden")
		ctx, cancel := context.WithTimeout(parentCtx, 20*time.Minute)
		defer cancel()

		Expect(runtimeClient.Create(ctx, backupSecret)).To(Succeed())
		Expect(runtimeClient.Create(ctx, garden)).To(Succeed())
		waitForGardenToBeReconciled(ctx, garden)

		DeferCleanup(func() {
			By("Delete Garden")
			ctx, cancel = context.WithTimeout(parentCtx, 5*time.Minute)
			defer cancel()

			Expect(gardenerutils.ConfirmDeletion(ctx, runtimeClient, garden)).To(Succeed())
			Expect(runtimeClient.Delete(ctx, garden)).To(Succeed())
			Expect(runtimeClient.Delete(ctx, backupSecret)).To(Succeed())
			waitForGardenToBeDeleted(ctx, garden)
			cleanupVolumes(ctx)
			Expect(runtimeClient.DeleteAllOf(ctx, &corev1.Secret{}, client.InNamespace(namespace), client.MatchingLabels{"role": "kube-apiserver-etcd-encryption-configuration"})).To(Succeed())
			Expect(runtimeClient.DeleteAllOf(ctx, &corev1.Secret{}, client.InNamespace(namespace), client.MatchingLabels{"role": "gardener-apiserver-etcd-encryption-configuration"})).To(Succeed())
		})

		v := rotationutils.Verifiers{
			// basic verifiers checking secrets
			&rotation.CAVerifier{RuntimeClient: runtimeClient, Garden: garden},
			&rotationutils.ObservabilityVerifier{
				GetObservabilitySecretFunc: func(ctx context.Context) (*corev1.Secret, error) {
					secretList := &corev1.SecretList{}
					if err := runtimeClient.List(ctx, secretList, client.InNamespace(v1beta1constants.GardenNamespace), client.MatchingLabels{
						"managed-by":       "secrets-manager",
						"manager-identity": "gardener-operator",
						"name":             "observability-ingress",
					}); err != nil {
						return nil, err
					}

					if length := len(secretList.Items); length != 1 {
						return nil, fmt.Errorf("expect exactly one secret, found %d", length)
					}

					return &secretList.Items[0], nil
				},
				GetObservabilityEndpoint: func(_ *corev1.Secret) string {
					return "https://plutono-garden." + garden.Spec.RuntimeCluster.Ingress.Domains[0]
				},
				GetObservabilityRotation: func() *gardencorev1beta1.ObservabilityRotation {
					return garden.Status.Credentials.Rotation.Observability
				},
			},
			&rotationutils.ETCDEncryptionKeyVerifier{
				RuntimeClient:               runtimeClient,
				Namespace:                   namespace,
				SecretsManagerLabelSelector: rotation.ManagedByGardenerOperatorSecretsManager,
				GetETCDEncryptionKeyRotation: func() *gardencorev1beta1.ETCDEncryptionKeyRotation {
					return garden.Status.Credentials.Rotation.ETCDEncryptionKey
				},
				EncryptionKey:  v1beta1constants.SecretNameETCDEncryptionKey,
				RoleLabelValue: v1beta1constants.SecretNamePrefixETCDEncryptionConfiguration,
			},
			&rotationutils.ETCDEncryptionKeyVerifier{
				RuntimeClient:               runtimeClient,
				Namespace:                   namespace,
				SecretsManagerLabelSelector: rotation.ManagedByGardenerOperatorSecretsManager,
				GetETCDEncryptionKeyRotation: func() *gardencorev1beta1.ETCDEncryptionKeyRotation {
					return garden.Status.Credentials.Rotation.ETCDEncryptionKey
				},
				EncryptionKey:  v1beta1constants.SecretNameGardenerETCDEncryptionKey,
				RoleLabelValue: v1beta1constants.SecretNamePrefixGardenerETCDEncryptionConfiguration,
			},
			&rotationutils.ServiceAccountKeyVerifier{
				RuntimeClient:               runtimeClient,
				Namespace:                   namespace,
				SecretsManagerLabelSelector: rotation.ManagedByGardenerOperatorSecretsManager,
				GetServiceAccountKeyRotation: func() *gardencorev1beta1.ServiceAccountKeyRotation {
					return garden.Status.Credentials.Rotation.ServiceAccountKey
				},
			},

			// advanced verifiers testing things from the user's perspective
			&rotationutils.EncryptedDataVerifier{
				NewTargetClientFunc: func() (kubernetes.Interface, error) {
					return kubernetes.NewClientFromSecret(ctx, runtimeClient, namespace, "gardener",
						kubernetes.WithDisabledCachedClient(),
						kubernetes.WithClientOptions(client.Options{Scheme: operatorclient.VirtualScheme}),
					)
				},
				Resources: []rotationutils.EncryptedResource{
					{
						NewObject: func() client.Object {
							return &corev1.Secret{
								ObjectMeta: metav1.ObjectMeta{GenerateName: "test-foo-", Namespace: "default"},
								StringData: map[string]string{"content": "foo"},
							}
						},
						NewEmptyList: func() client.ObjectList { return &corev1.SecretList{} },
					},
					{
						NewObject: func() client.Object {
							return &gardencorev1beta1.InternalSecret{
								ObjectMeta: metav1.ObjectMeta{GenerateName: "test-foo-", Namespace: "default"},
								StringData: map[string]string{"content": "foo"},
							}
						},
						NewEmptyList: func() client.ObjectList { return &gardencorev1beta1.InternalSecretList{} },
					},
					{
						NewObject: func() client.Object {
							return &gardencorev1beta1.ShootState{
								ObjectMeta: metav1.ObjectMeta{GenerateName: "test-foo-", Namespace: "default"},
								Spec:       gardencorev1beta1.ShootStateSpec{Gardener: []gardencorev1beta1.GardenerResourceData{{Name: "foo"}}},
							}
						},
						NewEmptyList: func() client.ObjectList { return &gardencorev1beta1.ShootStateList{} },
					},
					{
						NewObject: func() client.Object {
							return &gardencorev1beta1.ControllerDeployment{
								ObjectMeta: metav1.ObjectMeta{GenerateName: "test-foo-", Namespace: "default"},
								Type:       "helm",
							}
						},
						NewEmptyList: func() client.ObjectList { return &gardencorev1beta1.ControllerDeploymentList{} },
					},
					{
						NewObject: func() client.Object {
							suffix, err := utils.GenerateRandomString(5)
							Expect(err).NotTo(HaveOccurred())
							return &gardencorev1beta1.ControllerRegistration{
								ObjectMeta: metav1.ObjectMeta{GenerateName: "test-foo-", Namespace: "default"},
								Spec:       gardencorev1beta1.ControllerRegistrationSpec{Resources: []gardencorev1beta1.ControllerResource{{Kind: "Infrastructure", Type: "test-foo-" + suffix}}},
							}
						},
						NewEmptyList: func() client.ObjectList { return &gardencorev1beta1.ControllerRegistrationList{} },
					},
				},
			},
			&rotation.VirtualGardenAccessVerifier{RuntimeClient: runtimeClient, Namespace: namespace},
		}

		DeferCleanup(func() {
			ctx, cancel := context.WithTimeout(parentCtx, 2*time.Minute)
			defer cancel()

			v.Cleanup(ctx)
		})

		v.Before(ctx)

		By("Start credentials rotation")
		ctx, cancel = context.WithTimeout(parentCtx, 20*time.Minute)
		defer cancel()

		patch := client.MergeFrom(garden.DeepCopy())
		metav1.SetMetaDataAnnotation(&garden.ObjectMeta, v1beta1constants.GardenerOperation, v1beta1constants.OperationRotateCredentialsStart)
		Eventually(func() error {
			return runtimeClient.Patch(ctx, garden, patch)
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(runtimeClient.Get(ctx, client.ObjectKeyFromObject(garden), garden)).To(Succeed())
			g.Expect(garden.Annotations).NotTo(HaveKey(v1beta1constants.GardenerOperation))
			v.ExpectPreparingStatus(g)
		}).Should(Succeed())

		waitForGardenToBeReconciled(ctx, garden)

		Eventually(func() error {
			return runtimeClient.Get(ctx, client.ObjectKeyFromObject(garden), garden)
		}).Should(Succeed())

		v.AfterPrepared(ctx)

		By("Complete credentials rotation")
		ctx, cancel = context.WithTimeout(parentCtx, 20*time.Minute)
		defer cancel()

		patch = client.MergeFrom(garden.DeepCopy())
		metav1.SetMetaDataAnnotation(&garden.ObjectMeta, v1beta1constants.GardenerOperation, v1beta1constants.OperationRotateCredentialsComplete)
		Eventually(func() error {
			return runtimeClient.Patch(ctx, garden, patch)
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(runtimeClient.Get(ctx, client.ObjectKeyFromObject(garden), garden)).To(Succeed())
			g.Expect(garden.Annotations).NotTo(HaveKey(v1beta1constants.GardenerOperation))
			v.ExpectCompletingStatus(g)
		}).Should(Succeed())

		waitForGardenToBeReconciled(ctx, garden)

		Eventually(func(g Gomega) {
			g.Expect(runtimeClient.Get(ctx, client.ObjectKeyFromObject(garden), garden)).To(Succeed())
		}).Should(Succeed())

		v.AfterCompleted(ctx)
	})
})
