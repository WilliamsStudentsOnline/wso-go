package jobs

import (
	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func RunCatalogUpdateJob(cfg *config.Config) (*batchv1.Job, error) {
	return RunJob(cfg, "catalog-update", "catalog-update", nil)
}

func RunJob(cfg *config.Config, name string, command string, args []string) (*batchv1.Job, error) {
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	jobsClient := clientset.BatchV1().Jobs(cfg.KubeNamespace)
	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wso-job-" + name,
			Namespace: cfg.KubeNamespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "wso-job-" + name,
							Image:   "wso-backend-jobs:" + cfg.KubeJobImageVersion,
							Command: []string{"/wso-jobs/" + command},
							Args:    args,
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
			BackoffLimit: lib.Int32ToPtr(4),
		},
	}

	job, err := jobsClient.Create(jobSpec)
	if err != nil {
		return nil, err
	}

	return job, nil
}
