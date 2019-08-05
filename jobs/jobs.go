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

func RunUpdateAllUsersFromLDAPJob(cfg *config.Config) (*batchv1.Job, error) {
	return RunJob(cfg, "update-all-users-from-ldap", "update-all-users-from-ldap", []string{
		"--config=/etc/configs/config.yaml",
		"--disable-migration-check",
	})
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

	deploymentClient := clientset.AppsV1().Deployments(cfg.KubeNamespace)
	dply, err := deploymentClient.Get("backend", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// By default, job will be deleted after 5 minutes
	var jobLifetime int32 = 5 * 60

	// If it is production, keep job for 24h
	if cfg.IsProduction() {
		jobLifetime = 24 * 60 * 60
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
							Name:         "wso-job-" + name,
							Image:        "wso-backend-jobs:" + cfg.KubeJobImageVersion,
							Command:      []string{"/wso-jobs/" + command},
							Args:         args,
							Env:          dply.Spec.Template.Spec.Containers[0].Env,
							VolumeMounts: dply.Spec.Template.Spec.Containers[0].VolumeMounts,
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
					Volumes:       dply.Spec.Template.Spec.Volumes,
				},
			},
			BackoffLimit: lib.Int32ToPtr(4),
			// Cleanup job after a certain number of seconds
			TTLSecondsAfterFinished: &jobLifetime,
		},
	}

	job, err := jobsClient.Create(jobSpec)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func GetJob(cfg *config.Config, jobName string) (*batchv1.Job, error) {
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

	return jobsClient.Get(jobName, metav1.GetOptions{})
}
