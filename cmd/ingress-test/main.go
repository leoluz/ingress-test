package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/nsf/jsondiff"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	configPath := filepath.Join(dirname, ".kube", "config")
	kubeconfig := flag.String("kubeconfig", configPath, "kubeconfig file")
	flag.Parse()
	cfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	kubeClient := kubernetes.NewForConfigOrDie(cfg)
	namespace := metav1.NamespaceAll
	resyncDuration := 1 * time.Second
	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(
		kubeClient,
		resyncDuration,
		kubeinformers.WithNamespace(namespace))
	ingressInformer := kubeInformerFactory.Networking().V1().Ingresses()
	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			fmt.Println("added")
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			if oldObj == newObj {
				return
			}
			oldJson, _ := json.Marshal(oldObj)
			newJson, _ := json.Marshal(newObj)
			opts := jsondiff.DefaultConsoleOptions()
			_, text := jsondiff.Compare(oldJson, newJson, &opts)

			fmt.Println("updated")
			fmt.Printf("%s\n", text)
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("deleted")
		},
	})
	kubeInformerFactory.Start(wait.NeverStop)
	kubeInformerFactory.WaitForCacheSync(wait.NeverStop)
	c := make(chan os.Signal, 1)
	<-c
}
