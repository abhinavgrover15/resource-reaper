package controller

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func init() {
	log.SetLogger(zap.New())
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    time.Duration
		shouldError bool
	}{
		{
			name:        "valid duration - hours",
			input:       "24h",
			expected:    24 * time.Hour,
			shouldError: false,
		},
		{
			name:        "valid duration - minutes",
			input:       "30m",
			expected:    30 * time.Minute,
			shouldError: false,
		},
		{
			name:        "invalid duration",
			input:       "invalid",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := parseDuration(tt.input)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, duration)
			}
		})
	}
}

func createTestPod(name string, ttl string, creationTime time.Time) *corev1.Pod {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: creationTime},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test",
					Image: "test:latest",
				},
			},
		},
	}

	if ttl != "" {
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		pod.Annotations[ResourceReaperAnnotation] = ttl
	}

	return pod
}

func createTestDeployment(name string, ttl string, creationTime time.Time) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: creationTime},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "nginx",
						},
					},
				},
			},
		},
	}

	if ttl != "" {
		deployment.Annotations = map[string]string{
			ResourceReaperAnnotation: ttl,
		}
	}

	return deployment
}

func createTestService(name string, ttl string, creationTime time.Time) *corev1.Service {
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: creationTime},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
			Selector: map[string]string{
				"app": name,
			},
		},
	}

	if ttl != "" {
		svc.Annotations = map[string]string{
			ResourceReaperAnnotation: ttl,
		}
	}

	return svc
}

func createTestConfigMap(name string, ttl string, creationTime time.Time) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: creationTime},
		},
		Data: map[string]string{
			"test": "data",
		},
	}

	if ttl != "" {
		cm.Annotations = map[string]string{
			ResourceReaperAnnotation: ttl,
		}
	}

	return cm
}

func createTestSecret(name string, ttl string, creationTime time.Time) *corev1.Secret {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: creationTime},
		},
		Data: map[string][]byte{
			"test": []byte("data"),
		},
	}

	if ttl != "" {
		secret.Annotations = map[string]string{
			ResourceReaperAnnotation: ttl,
		}
	}

	return secret
}

func createTestStatefulSet(name string, ttl string, creationTime time.Time) *appsv1.StatefulSet {
	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: creationTime},
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "nginx",
						},
					},
				},
			},
		},
	}

	if ttl != "" {
		sts.Annotations = map[string]string{
			ResourceReaperAnnotation: ttl,
		}
	}

	return sts
}

func createTestDaemonSet(name string, ttl string, creationTime time.Time) *appsv1.DaemonSet {
	ds := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "DaemonSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: creationTime},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "nginx",
						},
					},
				},
			},
		},
	}

	if ttl != "" {
		ds.Annotations = map[string]string{
			ResourceReaperAnnotation: ttl,
		}
	}

	return ds
}

func createTestJob(name string, ttl string, creationTime time.Time) *batchv1.Job {
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: creationTime},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "nginx",
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	if ttl != "" {
		job.Annotations = map[string]string{
			ResourceReaperAnnotation: ttl,
		}
	}

	return job
}

func createTestCronJob(name string, ttl string, creationTime time.Time) *batchv1.CronJob {
	cronJob := &batchv1.CronJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "CronJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: creationTime},
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "*/5 * * * *",
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "nginx",
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
		},
	}

	if ttl != "" {
		cronJob.Annotations = map[string]string{
			ResourceReaperAnnotation: ttl,
		}
	}

	return cronJob
}

func setupTestReconciler(t *testing.T, objects ...client.Object) *ResourceReaper {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, appsv1.AddToScheme(scheme))
	require.NoError(t, batchv1.AddToScheme(scheme))
	require.NoError(t, metav1.AddMetaToScheme(scheme))

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objects...).
		Build()

	logger := zap.New()

	return &ResourceReaper{
		Client: client,
		Log:    logger.WithName("test"),
		Scheme: scheme,
	}
}

func TestReconcile(t *testing.T) {
	testCases := []struct {
		name           string
		ttl            string
		creationTime   time.Time
		expectedResult ctrl.Result
		expectDelete   bool
		expectError    bool
	}{
		{
			name:           "Pod with zero TTL should be deleted immediately",
			ttl:            "0s",
			creationTime:   time.Now().Add(-1 * time.Hour),
			expectedResult: ctrl.Result{},
			expectDelete:   true,
			expectError:    false,
		},
		{
			name:           "Pod with expired TTL should be deleted",
			ttl:            "1h",
			creationTime:   time.Now().Add(-2 * time.Hour),
			expectedResult: ctrl.Result{},
			expectDelete:   true,
			expectError:    false,
		},
		{
			name:           "Pod with future TTL should be requeued",
			ttl:            "2h",
			creationTime:   time.Now(),
			expectedResult: ctrl.Result{RequeueAfter: 2 * time.Hour},
			expectDelete:   false,
			expectError:    false,
		},
		{
			name:           "Pod with invalid TTL format should return error",
			ttl:            "invalid",
			creationTime:   time.Now(),
			expectedResult: ctrl.Result{},
			expectDelete:   false,
			expectError:    true,
		},
		{
			name:           "Pod with negative TTL should return error",
			ttl:            "-1h",
			creationTime:   time.Now(),
			expectedResult: ctrl.Result{},
			expectDelete:   false,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pod := createTestPod("test-pod", tc.ttl, tc.creationTime)
			r := setupTestReconciler(t, pod)

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-pod",
					Namespace: "default",
				},
			}

			result, err := r.Reconcile(context.Background(), req)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			err = r.Get(context.Background(), req.NamespacedName, &metav1.PartialObjectMetadata{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
			})

			if tc.expectDelete {
				require.True(t, errors.IsNotFound(err), "Resource should be deleted")
			} else if !tc.expectError {
				require.NoError(t, err, "Resource should still exist")
				duration, err := parseDuration(tc.ttl)
				require.NoError(t, err)
				require.WithinDuration(t,
					tc.creationTime.Add(duration),
					time.Now().Add(result.RequeueAfter),
					2*time.Second)
			}
		})
	}
}

func TestTTLReconciler_Reconcile(t *testing.T) {
	tests := []struct {
		name           string
		ttl            string
		creationTime   time.Time
		expectDeletion bool
		resourceType   string
	}{
		{
			name:           "pod_with_valid_TTL_annotation_should_be_deleted_after_expiry",
			ttl:            "1h",
			creationTime:   time.Now().Add(-2 * time.Hour),
			expectDeletion: true,
			resourceType:   "pod",
		},
		{
			name:           "pod_with_valid_TTL_annotation_but_not_expired_should_not_be_deleted",
			ttl:            "24h",
			creationTime:   time.Now(),
			expectDeletion: false,
			resourceType:   "pod",
		},
		{
			name:           "pod_without_TTL_annotation_should_be_ignored",
			ttl:            "",
			creationTime:   time.Now().Add(-24 * time.Hour),
			expectDeletion: false,
			resourceType:   "pod",
		},
		{
			name:           "pod_with_invalid_TTL_annotation_should_log_error",
			ttl:            "invalid",
			creationTime:   time.Now(),
			expectDeletion: false,
			resourceType:   "pod",
		},
		{
			name:           "pod_with_zero_TTL_should_be_deleted_immediately",
			ttl:            "0s",
			creationTime:   time.Now(),
			expectDeletion: true,
			resourceType:   "pod",
		},
		{
			name:           "pod_with_negative_TTL_should_be_treated_as_invalid",
			ttl:            "-1h",
			creationTime:   time.Now(),
			expectDeletion: false,
			resourceType:   "pod",
		},
		{
			name:           "deployment_with_valid_TTL_annotation_should_be_deleted_after_expiry",
			ttl:            "1h",
			creationTime:   time.Now().Add(-2 * time.Hour),
			expectDeletion: true,
			resourceType:   "deployment",
		},
		{
			name:           "deployment_with_valid_TTL_annotation_but_not_expired_should_not_be_deleted",
			ttl:            "24h",
			creationTime:   time.Now(),
			expectDeletion: false,
			resourceType:   "deployment",
		},
		{
			name:           "deployment_without_TTL_annotation_should_be_ignored",
			ttl:            "",
			creationTime:   time.Now().Add(-24 * time.Hour),
			expectDeletion: false,
			resourceType:   "deployment",
		},
		{
			name:           "deployment_with_invalid_TTL_annotation_should_log_error",
			ttl:            "invalid",
			creationTime:   time.Now(),
			expectDeletion: false,
			resourceType:   "deployment",
		},
		{
			name:           "deployment_with_zero_TTL_should_be_deleted_immediately",
			ttl:            "0s",
			creationTime:   time.Now(),
			expectDeletion: true,
			resourceType:   "deployment",
		},
		{
			name:           "deployment_with_negative_TTL_should_be_treated_as_invalid",
			ttl:            "-1h",
			creationTime:   time.Now(),
			expectDeletion: false,
			resourceType:   "deployment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var obj client.Object
			if tt.resourceType == "pod" {
				obj = createTestPod("test-pod", tt.ttl, tt.creationTime)
			} else {
				obj = createTestDeployment("test-deployment", tt.ttl, tt.creationTime)
			}

			r := setupTestReconciler(t, obj)

			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      obj.GetName(),
					Namespace: obj.GetNamespace(),
				},
			}

			result, err := r.Reconcile(context.Background(), req)
			if tt.ttl == "invalid" || tt.ttl == "-1h" {
				assert.Error(t, err, "Expected error for invalid TTL")
			} else {
				assert.NoError(t, err, "Reconciliation should not return error")
			}

			if tt.ttl != "" && tt.ttl != "invalid" && tt.ttl != "-1h" {
				duration, err := parseDuration(tt.ttl)
				assert.NoError(t, err)
				if !tt.expectDeletion {
					assert.InDelta(t, duration, result.RequeueAfter, float64(time.Minute))
				}
			}

			err = r.Get(context.Background(), req.NamespacedName, obj)
			if tt.expectDeletion {
				assert.True(t, errors.IsNotFound(err), "Resource should be deleted")
			} else {
				assert.NoError(t, err, "Resource should still exist")
			}
		})
	}
}

func TestTTLReconciler(t *testing.T) {
	tests := []struct {
		name            string
		ttl             string
		shouldDelete    bool
		expectedError   bool
		expectedRequeue bool
	}{
		{
			name:            "Valid TTL not expired",
			ttl:             "24h",
			shouldDelete:    false,
			expectedError:   false,
			expectedRequeue: true,
		},
		{
			name:            "Zero TTL",
			ttl:             "0s",
			shouldDelete:    true,
			expectedError:   false,
			expectedRequeue: false,
		},
		{
			name:            "Invalid TTL",
			ttl:             "invalid",
			shouldDelete:    false,
			expectedError:   true,
			expectedRequeue: false,
		},
		{
			name:            "Negative TTL",
			ttl:             "-1h",
			shouldDelete:    false,
			expectedError:   true,
			expectedRequeue: false,
		},
	}

	resources := []struct {
		name string
		obj  client.Object
	}{
		{
			name: "Pod",
			obj: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "test:latest",
						},
					},
				},
			},
		},
		{
			name: "Deployment",
			obj: &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "test",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test:latest",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Service",
			obj: &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "test",
					},
					Ports: []corev1.ServicePort{
						{
							Port:       80,
							TargetPort: intstr.FromInt(8080),
						},
					},
				},
			},
		},
		{
			name: "ConfigMap",
			obj: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-configmap",
					Namespace: "default",
				},
				Data: map[string]string{
					"key": "value",
				},
			},
		},
		{
			name: "Secret",
			obj: &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-secret",
					Namespace: "default",
				},
				StringData: map[string]string{
					"username": "admin",
					"password": "secret",
				},
			},
		},
		{
			name: "StatefulSet",
			obj: &appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-statefulset",
					Namespace: "default",
				},
				Spec: appsv1.StatefulSetSpec{
					ServiceName: "test",
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "test",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test:latest",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "DaemonSet",
			obj: &appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "apps/v1",
					Kind:       "DaemonSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-daemonset",
					Namespace: "default",
				},
				Spec: appsv1.DaemonSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "test",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test:latest",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Job",
			obj: &batchv1.Job{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "batch/v1",
					Kind:       "Job",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-job",
					Namespace: "default",
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test:latest",
								},
							},
							RestartPolicy: corev1.RestartPolicyNever,
						},
					},
				},
			},
		},
		{
			name: "CronJob",
			obj: &batchv1.CronJob{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "batch/v1",
					Kind:       "CronJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cronjob",
					Namespace: "default",
				},
				Spec: batchv1.CronJobSpec{
					Schedule: "*/5 * * * *",
					JobTemplate: batchv1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "test",
											Image: "test:latest",
										},
									},
									RestartPolicy: corev1.RestartPolicyNever,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, res := range resources {
		t.Run(res.name, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					s := runtime.NewScheme()
					require.NoError(t, scheme.AddToScheme(s))
					require.NoError(t, appsv1.AddToScheme(s))
					require.NoError(t, batchv1.AddToScheme(s))

					r := &ResourceReaper{
						Client: fake.NewClientBuilder().WithScheme(s).Build(),
						Log:    zap.New(),
						Scheme: s,
					}

					// Use a fake rest config for testing
					cfg := &rest.Config{
						Host: "localhost",
					}

					mgr, err := ctrl.NewManager(cfg, ctrl.Options{
						Scheme: s,
						//Metrics: ctrl.MetricsConfig{
						//	BindAddress: "0", // Equivalent to previous MetricsBindAddress
						//},
						NewClient: func(cfg *rest.Config, opts client.Options) (client.Client, error) {
							return fake.NewClientBuilder().WithScheme(s).Build(), nil // Use the fake client for testing
						},
					})
					require.NoError(t, err)

					err = r.SetupWithManager(mgr)
					assert.NoError(t, err)
				})
			}
		})
	}
}

func TestTTLReconciler_SetupWithManager(t *testing.T) {
	s := runtime.NewScheme()
	err := corev1.AddToScheme(s)
	require.NoError(t, err)
	err = appsv1.AddToScheme(s)
	require.NoError(t, err)

	// Use a fake rest config for testing
	cfg := &rest.Config{
		Host: "localhost",
	}

	r := &ResourceReaper{
		Client: fake.NewClientBuilder().WithScheme(s).Build(),
		Log:    log.Log.WithName("test"),
		Scheme: s,
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: s,
		//Metrics: ctrl.MetricsConfig{
		//	BindAddress: "0", // Equivalent to previous MetricsBindAddress
		//},
		NewClient: func(cfg *rest.Config, opts client.Options) (client.Client, error) {
			return fake.NewClientBuilder().WithScheme(s).Build(), nil // Use the fake client for testing
		},
	})
	require.NoError(t, err)

	err = r.SetupWithManager(mgr)
	assert.NoError(t, err)
}
