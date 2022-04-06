package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

type BetaJob struct {
	Namespace    string
	Image        string
	JobReference string
	Jobscript    string
	User         string
	BackoffLimit int32
	MaxDuration  string
	// EJF Env variables have to be strings for Pod injection
	EJF_OUT          string
	EJF_ERR          string
	EJF_PATH         string
	EJF_JOBID        string
	EJF_JOBNAME      string
	EJF_SHORTJOBNAME string
	EJF_SJOBID       string
	EJF_SUBDATE      string
	EJF_SUBTIME      string
	EJF_SYSID        string
	EJF_DBSSID       string
}

// Add our Sidecars
// Overwrite Name, add Labels, Volumes, ENV Vars

func main() {
	// Pull Job Definition
	cmd := exec.Command("git", "clone", "https://github.com/hndrk-k/testjob.git", "tmpJob")

	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Git Output: " + string(stdout))

	// Parse JSON from File
	dat, err := os.ReadFile("./tmpJob/job.json")
	if err != nil {
		fmt.Printf("%#v", err)
	}
	fmt.Println("Read JSON: " + string(dat))

	// Convert JSON into K8s Structs
	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode([]byte(dat), nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}

	var cmMode int32 = 511
	var uniqueJobname = "testjob-1337-01"
	var newJob BetaJob = BetaJob{
		Namespace:    "",
		Image:        "git123",
		JobReference: "git456",
		Jobscript:    "echo \"HELLO!\"",
		User:         "heka",
		BackoffLimit: 3,
		MaxDuration:  "1m",
		// EJF Env variables have to be strings for Pod injection
		EJF_OUT:          "TEST",
		EJF_ERR:          "TEST",
		EJF_PATH:         "TEST",
		EJF_JOBID:        "TEST",
		EJF_JOBNAME:      "TEST",
		EJF_SHORTJOBNAME: "TEST",
		EJF_SJOBID:       "TEST",
		EJF_SUBDATE:      "TEST",
		EJF_SUBTIME:      "TEST",
		EJF_SYSID:        "TEST",
		EJF_DBSSID:       "TEST",
	}

	jobSpec := obj.(*batchv1.Job)
	jobSpec.Name = uniqueJobname
	jobSpec.ObjectMeta.Labels["job-name"] = uniqueJobname

	if jobSpec.Spec.Template.Spec.Containers[0].Name == "job" {

		jobSpec.Spec.Template.Spec.Containers[0].Command = []string{"/workdir/scripts/wrapper.sh"}
		jobSpec.Spec.Template.Spec.Containers[0].Env = append(jobSpec.Spec.Template.Spec.Containers[0].Env,
			v1.EnvVar{
				Name:  "EJF_OUT",
				Value: newJob.EJF_OUT,
			},
			v1.EnvVar{
				Name:  "EJF_ERR",
				Value: newJob.EJF_ERR,
			},
			v1.EnvVar{
				Name:  "EJF_PATH",
				Value: newJob.EJF_PATH,
			},
			v1.EnvVar{
				Name:  "EJF_JOBID",
				Value: newJob.EJF_JOBID,
			},
			v1.EnvVar{
				Name:  "EJF_JOBNAME",
				Value: newJob.EJF_JOBNAME,
			},
			v1.EnvVar{
				Name:  "EJF_SHORTJOBNAME",
				Value: newJob.EJF_SHORTJOBNAME,
			},
			v1.EnvVar{
				Name:  "EJF_SJOBID",
				Value: newJob.EJF_SJOBID,
			},
			v1.EnvVar{
				Name:  "EJF_SUBDATE",
				Value: newJob.EJF_SUBDATE,
			},
			v1.EnvVar{
				Name:  "EJF_SUBTIME",
				Value: newJob.EJF_SUBTIME,
			},
			v1.EnvVar{
				Name:  "EJF_SYSID",
				Value: newJob.EJF_SYSID,
			},
			v1.EnvVar{
				Name:  "EJF_DBSSID",
				Value: newJob.EJF_DBSSID,
			})

		jobSpec.Spec.Template.Spec.Containers[0].VolumeMounts = append(jobSpec.Spec.Template.Spec.Containers[0].VolumeMounts,
			v1.VolumeMount{
				Name:      "workdir",
				MountPath: "/workdir",
			},
			v1.VolumeMount{
				Name:      "scripts",
				MountPath: "/workdir/scripts",
			})

	}

	jobSpec.Spec.Template.Spec.Containers = append(jobSpec.Spec.Template.Spec.Containers, v1.Container{
		Name:    "sidecar",
		Image:   "curlimages/curl:latest",
		Command: []string{"/workdir/scripts/poll.sh"},
		Env: []v1.EnvVar{
			{
				Name:  "EJF_PATH",
				Value: newJob.EJF_PATH,
			},
			{
				Name:  "EJF_JOBID",
				Value: newJob.EJF_JOBID,
			},
			{
				Name:  "EJF_JOBNAME",
				Value: newJob.EJF_JOBNAME,
			},
			{
				Name:  "EJF_DBSSID",
				Value: newJob.EJF_DBSSID,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/workdir",
			},
			{
				Name:      "scripts",
				MountPath: "/workdir/scripts",
			},
		},
	})

	jobSpec.Spec.Template.Spec.Volumes = append(jobSpec.Spec.Template.Spec.Volumes,
		v1.Volume{
			Name: "workdir",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
		v1.Volume{
			Name: "scripts",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: uniqueJobname + "-cm",
					},
					DefaultMode: &cmMode,
				},
			},
		})

	jobSpec.Spec.Template.Spec.RestartPolicy = v1.RestartPolicy(v1.RestartPolicyNever)

	fmt.Printf("Created Job Spec:\n%v\n", jobSpec)

	jobSpecJSON, _ := json.MarshalIndent(jobSpec, "", "  ")
	fmt.Println("JSON FORMAT:\n" + string(jobSpecJSON))
}

// func createJobSpec(newJob BetaJob, uniqueJobname string) *batchv1.Job {
// 	var cmMode int32 = 511 // !! Für Security reviewen!

// 	return &batchv1.Job{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      uniqueJobname,
// 			Namespace: newJob.Namespace,
// 		},
// 		Spec: batchv1.JobSpec{
// 			Template: v1.PodTemplateSpec{
// 				Spec: v1.PodSpec{
// 					Containers: []v1.Container{
// 						{
// 							Name:  "job",
// 							Image: newJob.Image,
// 							// Command: []string{"/bin/sh", "-c"},
// 							// Args:    []string{"sleep 600"},
// 							Command: []string{"/workdir/scripts/wrapper.sh"},
// 							Env: []v1.EnvVar{
// 								{
// 									Name:  "EJF_OUT",
// 									Value: newJob.EJF_OUT,
// 								},
// 								{
// 									Name:  "EJF_ERR",
// 									Value: newJob.EJF_ERR,
// 								},
// 								{
// 									Name:  "EJF_PATH",
// 									Value: newJob.EJF_PATH,
// 								},
// 								{
// 									Name:  "EJF_JOBID",
// 									Value: newJob.EJF_JOBID,
// 								},
// 								{
// 									Name:  "EJF_JOBNAME",
// 									Value: newJob.EJF_JOBNAME,
// 								},
// 								{
// 									Name:  "EJF_SHORTJOBNAME",
// 									Value: newJob.EJF_SHORTJOBNAME,
// 								},
// 								{
// 									Name:  "EJF_SJOBID",
// 									Value: newJob.EJF_SJOBID,
// 								},
// 								{
// 									Name:  "EJF_SUBDATE",
// 									Value: newJob.EJF_SUBDATE,
// 								},
// 								{
// 									Name:  "EJF_SUBTIME",
// 									Value: newJob.EJF_SUBTIME,
// 								},
// 								{
// 									Name:  "EJF_SYSID",
// 									Value: newJob.EJF_SYSID,
// 								},
// 								{
// 									Name:  "EJF_DBSSID",
// 									Value: newJob.EJF_DBSSID,
// 								},
// 							},
// 							VolumeMounts: []v1.VolumeMount{
// 								{
// 									Name:      "workdir",
// 									MountPath: "/workdir",
// 								},
// 								{
// 									Name:      "scripts",
// 									MountPath: "/workdir/scripts",
// 								},
// 							},
// 						},
// 						{
// 							Name:  "sidecar",
// 							Image: "curlimages/curl:latest",
// 							// Command: []string{"/bin/sh", "-c"},
// 							// Args:    []string{"sleep 600"},
// 							Command: []string{"/workdir/scripts/poll.sh"},
// 							Env: []v1.EnvVar{
// 								{
// 									Name:  "EJF_PATH",
// 									Value: newJob.EJF_PATH,
// 								},
// 								{
// 									Name:  "EJF_JOBID",
// 									Value: newJob.EJF_JOBID,
// 								},
// 								{
// 									Name:  "EJF_JOBNAME",
// 									Value: newJob.EJF_JOBNAME,
// 								},
// 								{
// 									Name:  "EJF_DBSSID",
// 									Value: newJob.EJF_DBSSID,
// 								},
// 							},
// 							VolumeMounts: []v1.VolumeMount{
// 								{
// 									Name:      "workdir",
// 									MountPath: "/workdir",
// 								},
// 								{
// 									Name:      "scripts",
// 									MountPath: "/workdir/scripts",
// 								},
// 							},
// 						},
// 					},
// 					Volumes: []v1.Volume{
// 						{
// 							Name: "workdir",
// 							VolumeSource: v1.VolumeSource{
// 								EmptyDir: &v1.EmptyDirVolumeSource{},
// 							},
// 						},
// 						{
// 							Name: "scripts",
// 							VolumeSource: v1.VolumeSource{
// 								ConfigMap: &v1.ConfigMapVolumeSource{
// 									LocalObjectReference: v1.LocalObjectReference{
// 										Name: uniqueJobname + "-cm",
// 									},
// 									DefaultMode: &cmMode,
// 								},
// 							},
// 						},
// 					},
// 					RestartPolicy: v1.RestartPolicy(v1.RestartPolicyNever),
// 				},
// 			},
// 			BackoffLimit: &newJob.BackoffLimit,
// 		},
// 	}
// }
