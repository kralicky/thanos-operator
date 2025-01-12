// Copyright 2020 Banzai Cloud
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

package bucketweb

import (
	"emperror.dev/errors"
	"github.com/banzaicloud/operator-tools/pkg/merge"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/thanos-operator/pkg/resources"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (b *BucketWeb) service() (runtime.Object, reconciler.DesiredState, error) {
	meta := b.getMeta(b.getName())
	if b.ObjectStore.Spec.BucketWeb != nil {
		bucketWeb := b.ObjectStore.Spec.BucketWeb
		service := &corev1.Service{
			ObjectMeta: bucketWeb.MetaOverrides.Merge(meta),
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Protocol: corev1.ProtocolTCP,
						Name:     "http",
						Port:     resources.GetPort(bucketWeb.HTTPAddress),
						TargetPort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: resources.GetPort(bucketWeb.HTTPAddress),
						},
					},
				},
				Selector: b.getLabels(),
			},
		}

		if bucketWeb.ServiceOverrides != nil {
			if err := merge.Merge(service, bucketWeb.ServiceOverrides); err != nil {
				return service, reconciler.StatePresent, errors.WrapIf(err, "unable to merge overrides to service base")
			}
		}

		return service, reconciler.DesiredStateHook(func(current runtime.Object) error {
			if s, ok := current.(*corev1.Service); ok {
				service.Spec.ClusterIP = s.Spec.ClusterIP
			} else {
				return errors.Errorf("failed to cast service object %+v", current)
			}
			return nil
		}), nil
	}

	return &corev1.Service{
		ObjectMeta: meta,
	}, reconciler.StateAbsent, nil
}
