package kube

import (
	"fmt"
	"log"

	"github.com/codegp/cloud-persister/models"
	jobs "github.com/codegp/job-client/jobclient"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/watch"
)

var _ jobs.JobClient = (*KubeClient)(nil)

// KubeClient creats jobs that are kube pods
type KubeClient struct {
	client *unversioned.Client
}

var _ jobs.Job = (*KubeJob)(nil)

// KubeJob is a watchable job
type KubeJob struct {
	pod    *api.Pod
	client *unversioned.Client
}

// NewClient returns a client that can communicate to the apiserver to start jobs.
// IMPORTANT: Creating a client with this function only work if running from inside kubernetes cluster!
// @returns valid client if successful, error otherwise
func NewClient() (*KubeClient, error) {
	var c *unversioned.Client
	var err error

	c, err = unversioned.NewInCluster()
	if err != nil {
		return nil, fmt.Errorf("Can't connect to Kubernetes API: %v", err)
	}

	return &KubeClient{
		client: c,
	}, nil
}

// StartGame starts a game, requires a game and the projects associated with the game
// @returns error if client fails to start the game runner pod
func (kc *KubeClient) StartGame(game *models.Game, projects []*models.Project) (jobs.Job, error) {
	pod := createGamePod(game)
	log.Printf("Starting Game! Spec:\n%v", pod)
	pod, err := kc.client.Pods(api.NamespaceDefault).Create(pod)
	if err != nil {
		return nil, err
	}
	return kc.podToKubeJob(pod), nil
}

// BuildGameType kicks off a gametype build
// @returns error if client fails to start the gametype builder pod
func (kc *KubeClient) BuildGameType(gameType *models.GameType) (jobs.Job, error) {
	pod := createGameTypeBuilderPod(gameType)
	log.Printf("Building Gametype Builder! Spec:\n%v", pod)
	pod, err := kc.client.Pods(api.NamespaceDefault).Create(pod)
	if err != nil {
		return nil, err
	}
	return kc.podToKubeJob(pod), nil
}

func (kc *KubeClient) podToKubeJob(pod *api.Pod) *KubeJob {
	return &KubeJob{
		pod:    pod,
		client: kc.client,
	}
}

type DesiredPodState func(*api.Pod) bool

func (kj *KubeJob) watchPod(isDesired DesiredPodState) error {
	watcher, err := kj.client.Pods(api.NamespaceDefault).Watch(nameSelector(kj.pod.Name))
	if err != nil {
		return err
	}
	for {
		e := <-watcher.ResultChan()
		pod := e.Object.(*api.Pod)
		log.Printf("Update received for pod %s, EventType %v\n", pod.Name, e.Type)
		if isDesired(pod) {
			log.Printf("Reached desired state for pod %s", pod.Name)
			return nil
		}
		if pod.Status.Phase == api.PodFailed {
			return fmt.Errorf("Pod failed.\nPod:\n%v", pod)
		}
		if e.Type == watch.Deleted {
			return fmt.Errorf("Pod was deleted before reaching desired state.\nPod:\n%v", pod)
		}
		if e.Type == watch.Error {
			return fmt.Errorf("Pod errored before reaching desired state.\nPod:\n%v", pod)
		}
	}
}

func nameSelector(name string) api.ListOptions {
	selector := fields.Set{"metadata.name": name}.AsSelector()
	return api.ListOptions{FieldSelector: selector}
}

func (kj *KubeJob) WatchToCompletion() error {
	log.Printf("Watching %s to completion\n", kj.pod.Name)
	return kj.watchPod(func(pod *api.Pod) bool {
		if pod.Status.Phase == api.PodSucceeded {
			return true
		}
		return false
	})
}

func (kj *KubeJob) WatchToStartup() error {
	log.Printf("Watching %s to startup\n", kj.pod.Name)
	return kj.watchPod(func(pod *api.Pod) bool {
		if pod.Status.Phase == api.PodRunning {
			return true
		}
		return false
	})
}
