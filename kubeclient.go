package kubeclient

import (
	"fmt"
	"log"

	"github.com/codegp/cloud-persister/models"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/watch"
)

type KubeClient struct {
	client *unversioned.Client
}

// NewClient returns a client that can communicate to the apiserver to start jobs.
// IMPORTANT: Creating a client with this function only work if running from inside kubernetes cluster!
// @returns valid client if successful, error otherwise
func NewClient() (*KubeClient, error) {
	var c *unversioned.Client
	var err error

	// if codegpenv.IsLocal() {
	// 	config := &restclient.Config{
	// 		Host: "http://localhost:8080",
	// 	}
	// 	c, err = unversioned.New(config)
	// } else {
	// 	c, err = unversioned.NewInCluster()
	// }
	c, err = unversioned.NewInCluster()
	if err != nil {
		return nil, fmt.Errorf("Can't connect to Kubernetes API: %v", err)
	}

	return &KubeClient{
		client: c,
	}, nil
}

// LoadFromFile parses an Info object from a file path.
// If the file does not exist, then os.IsNotExist(err) == true
// func LoadFromFile(path string) (*Info, error) {
// 	var info Info
// 	if _, err := os.Stat(path); os.IsNotExist(err) {
// 		return nil, err
// 	}
// 	data, err := ioutil.ReadFile(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = json.Unmarshal(data, &info)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &info, err
// }

type DesiredPodState func(*api.Pod) bool

func (kc *KubeClient) watchPod(pod *api.Pod, listOptions api.ListOptions, isDesired DesiredPodState) (*api.Pod, error) {
	watcher, err := kc.client.Pods(api.NamespaceDefault).Watch(listOptions)
	if err != nil {
		return nil, err
	}
	for {
		e := <-watcher.ResultChan()
		pod := e.Object.(*api.Pod)
		log.Printf("Update received for pod %s, EventType %v\n", pod.Name, e.Type)
		if isDesired(pod) {
			log.Printf("Reached desired state for pod %s", pod.Name)
			return pod, nil
		}
		if pod.Status.Phase == api.PodFailed {
			return nil, fmt.Errorf("Pod failed.\nPod:\n%v", pod)
		}
		if e.Type == watch.Deleted {
			return nil, fmt.Errorf("Pod was deleted before reaching desired state.\nPod:\n%v", pod)
		}
		if e.Type == watch.Error {
			return nil, fmt.Errorf("Pod errored before reaching desired state.\nPod:\n%v", pod)
		}
	}
}

func nameSelector(name string) api.ListOptions {
	selector := fields.Set{"metadata.name": name}.AsSelector()
	return api.ListOptions{FieldSelector: selector}
}

func (kc *KubeClient) WatchToCompletion(pod *api.Pod) (*api.Pod, error) {
	log.Printf("Watching %s to completion\n", pod.Name)
	return kc.watchPod(pod, nameSelector(pod.Name), func(pod *api.Pod) bool {
		if pod.Status.Phase == api.PodSucceeded {
			return true
		}
		return false
	})
}

func (kc *KubeClient) WatchToStartup(pod *api.Pod) (*api.Pod, error) {
	log.Printf("Watching %s to startup\n", pod.Name)
	return kc.watchPod(pod, nameSelector(pod.Name), func(pod *api.Pod) bool {
		if pod.Status.Phase == api.PodRunning {
			return true
		}
		return false
	})
}

// StartGame starts a game, requires a game and the projects associated with the game
// @returns error if client fails to start the game runner pod
func (kc *KubeClient) StartGame(game *models.Game) (*api.Pod, error) {
	pod := createGamePod(game)
	log.Printf("Starting Game! Spec:\n%v", pod)
	return kc.client.Pods(api.NamespaceDefault).Create(pod)
}

// StarBot starts a bot, which runs a users code in the container built for their language
// @returns error if client fails to start the bot pod
func (kc *KubeClient) StartBot(ip string, botID int32, proj *models.Project, game *models.Game) (*api.Pod, error) {
	pod := createBotPod(ip, botID, proj, game)
	log.Printf("Starting bot! Spec:\n%v", pod)
	return kc.client.Pods(api.NamespaceDefault).Create(pod)
}

// BuildGameType kicks off a gametype build
// @returns error if client fails to start the gametype builder pod
func (kc *KubeClient) BuildGameType(gameType *models.GameType) (*api.Pod, error) {
	pod := createGameTypeBuilderPod(gameType)
	log.Printf("Building Gametype Builder! Spec:\n%v", pod)
	return kc.client.Pods(api.NamespaceDefault).Create(pod)
}
