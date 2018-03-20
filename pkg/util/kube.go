package util

import (
	"fmt"
	"time"

	"github.com/ashcrow/smooth-operator/pkg/retryutil"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// atomicUpdatorVolumeMounts returns the mount structure for an AUContainer.
func atomicUpdatorVolumeMounts() []v1.VolumeMount {
	return []v1.VolumeMount{
		{Name: "/usr/share/rpm", MountPath: "/usr/share/rpm"},
		{Name: "/etc", MountPath: "/etc"},
		{Name: "/ostree", MountPath: "/ostree"},
		{Name: "/boot", MountPath: "/boot"},
	}
}

// AUContainer returns the full spec for an Atomic Updator container
func AUContainer(command []string, repo, tag string) v1.Container {
	fullImage := fmt.Sprintf("%s:%v", repo, tag)
	container := v1.Container{
		Command:      command,
		Name:         "au",
		Image:        fullImage,
		VolumeMounts: atomicUpdatorVolumeMounts(),
	}

	return container
}

// AUPod creates an Atomic Updator pod spec
func AUPod(container v1.Container, labels map[string]string) *v1.Pod {
	runAsNonRoot := true
	id := int64(9000)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "au",
			Labels:      labels,
			Annotations: map[string]string{},
		},
		Spec: v1.PodSpec{
			//			AutomountServiceAccountToken: "",
			Containers:    []v1.Container{container},
			RestartPolicy: v1.RestartPolicyNever,
			//			Volumes:                      volumes,
			//			Hostname:                     "hostname",
			//			Subdomain:                    clusterName,
			SecurityContext: &v1.PodSecurityContext{
				RunAsUser:    &id,
				RunAsNonRoot: &runAsNonRoot,
				FSGroup:      &id,
			},
		},
	}

	return pod
}

// CreateAndWaitPod creates a pod and waits until it is running
// From: https://github.com/coreos/etcd-operator/blob/master/pkg/util/k8sutil/k8sutil.go#L190
func CreateAndWaitPod(kubecli kubernetes.Interface, ns string, pod *v1.Pod, timeout time.Duration) (*v1.Pod, error) {
	_, err := kubecli.CoreV1().Pods(ns).Create(pod)
	if err != nil {
		return nil, err
	}

	interval := 5 * time.Second
	var retPod *v1.Pod
	err = retryutil.Retry(interval, int(timeout/(interval)), func() (bool, error) {
		retPod, err = kubecli.CoreV1().Pods(ns).Get(pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch retPod.Status.Phase {
		case v1.PodRunning:
			return true, nil
		case v1.PodPending:
			return false, nil
		default:
			return false, fmt.Errorf("unexpected pod status.phase: %v", retPod.Status.Phase)
		}
	})

	if err != nil {
		if retryutil.IsRetryFailure(err) {
			return nil, fmt.Errorf("failed to wait pod running, it is still pending: %v", err)
		}
		return nil, fmt.Errorf("failed to wait pod running: %v", err)
	}

	return retPod, nil
}
