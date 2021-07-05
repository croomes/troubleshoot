package collect

import (
	"reflect"
	"testing"

	troubleshootv1beta2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_constructPod(t *testing.T) {
	var valTrue = true
	tests := []struct {
		name         string
		namespace    string
		runCollector *troubleshootv1beta2.Run
		want         *corev1.Pod
	}{
		{
			name:      "example",
			namespace: "default",
			runCollector: &troubleshootv1beta2.Run{
				CollectorMeta: troubleshootv1beta2.CollectorMeta{
					CollectorName: "run-ping",
				},
				Namespace:       "default",
				Image:           "busybox:1",
				Command:         []string{"ping"},
				Args:            []string{"-w", "5", "www.google.com"},
				ImagePullPolicy: "IfNotPresent",
			},
			want: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "run-ping",
					Namespace: "default",
					Labels: map[string]string{
						"troubleshoot-role": "run-collector",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Image:           "busybox:1",
							ImagePullPolicy: corev1.PullPolicy("IfNotPresent"),
							Name:            "collector",
							Command:         []string{"ping"},
							Args:            []string{"-w", "5", "www.google.com"},
						},
					},
				},
			},
		},
		{
			name:         "defaults",
			namespace:    "default",
			runCollector: &troubleshootv1beta2.Run{},
			want: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "",
					Namespace: "default",
					Labels: map[string]string{
						"troubleshoot-role": "run-collector",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:            "collector",
							ImagePullPolicy: corev1.PullIfNotPresent,
						},
					},
				},
			},
		},
		{
			name:      "privileged pod with volume",
			namespace: "myns",
			runCollector: &troubleshootv1beta2.Run{
				CollectorMeta: troubleshootv1beta2.CollectorMeta{
					CollectorName: "run-lsmod",
				},
				Namespace: "default",
				Pod: &corev1.Pod{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "overwritten-by-CollectorName",
						Namespace: "overwritten-by-namespace",
						Labels: map[string]string{
							"my-label": "my-label-val",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:    "collector-lsmod",
								Command: []string{"lsmod"},
								SecurityContext: &corev1.SecurityContext{
									AllowPrivilegeEscalation: &valTrue,
									Privileged:               &valTrue,
									Capabilities: &corev1.Capabilities{
										Add: []corev1.Capability{"SYS_ADMIN"},
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "modules",
										MountPath: "/lib/modules",
										ReadOnly:  true,
									},
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "modules",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										Path: "/lib/modules",
									},
								},
							},
						},
					},
				},
			},
			want: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "run-lsmod",
					Namespace: "myns",
					Labels: map[string]string{
						"my-label":          "my-label-val",
						"troubleshoot-role": "run-collector",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    "collector-lsmod",
							Command: []string{"lsmod"},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: &valTrue,
								Privileged:               &valTrue,
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{"SYS_ADMIN"},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "modules",
									MountPath: "/lib/modules",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "modules",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/lib/modules",
								},
							},
						},
					},
				},
			},
		},
		{
			name:      "minimal custom pod",
			namespace: "myns",
			runCollector: &troubleshootv1beta2.Run{
				CollectorMeta: troubleshootv1beta2.CollectorMeta{
					CollectorName: "run-minimal",
				},
				Pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Command: []string{"true"},
							},
						},
					},
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "run-minimal",
					Namespace: "myns",
					Labels: map[string]string{
						"troubleshoot-role": "run-collector",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Command: []string{"true"},
						},
					},
				},
			},
		},
		{
			name:      "multiple containers",
			namespace: "myns",
			runCollector: &troubleshootv1beta2.Run{
				CollectorMeta: troubleshootv1beta2.CollectorMeta{
					CollectorName: "run-multiple",
				},
				Pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Command: []string{"true"},
							},
							{
								Command: []string{"false"},
							},
						},
					},
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "run-multiple",
					Namespace: "myns",
					Labels: map[string]string{
						"troubleshoot-role": "run-collector",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Command: []string{"true"},
						},
						{
							Command: []string{"false"},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := constructPod(test.runCollector, test.namespace); !reflect.DeepEqual(got, test.want) {
				t.Errorf("constructPod() = \n%v\nwant \n%v", got, test.want)
			}
		})
	}
}
