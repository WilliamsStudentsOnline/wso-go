package jobs

import (
	"math/rand"
	"time"
	"unsafe"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
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

func RunUpdateAllFactrakSurveyDeficits(cfg *config.Config) (*batchv1.Job, error) {
	return RunJob(cfg, "update-all-factrak-survey-deficits", "update-all-factrak-survey-deficits", []string{
		"--config=/etc/configs/config.yaml",
	})
}

func RunDormsUpdateJob(cfg *config.Config) (*batchv1.Job, error) {
	return RunJob(cfg, "dorms-update", "dorms-update", []string{
		"--config=/etc/configs/config.yaml",
		"--local=true",
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
			Name:      "wso-job-" + name + "-" + RandString(5),
			Namespace: cfg.KubeNamespace,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:         "wso-job-" + name + "-" + RandString(5),
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

// Generates a random string of n length efficiently.
// Copied from https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go.
func RandString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, randSrc.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = randSrc.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// Random source, so we don't have to create a new one every time we call RandString()
var randSrc = rand.NewSource(time.Now().UnixNano())

// Valid letters we can use for RandString()
const letterBytes = "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"

// Letter masking for RandString()
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)
