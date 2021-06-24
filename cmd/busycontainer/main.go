package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var numNamespaces = 4

var namespaces = []string{"loki", "workload", "data-science", "payments"}
var podLists = make([]*v1.PodList, 4)
var podName = os.Getenv("POD_NAME")
var podNamespace = os.Getenv("POD_NAMESPACE")

var portsToConnectOn = []int{3998, 3999, 4000, 4001, 4002, 4003}

func startPodConnectionParty() {
	// Discover the configuration of the cluster running this process
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for {
		var skipIteration = false

		// Get all the pods in the target namespaces
		// Repeated so that we handle the case that some of the pods are still creating
		for i := 0; i < numNamespaces; i++ {
			podLists[i], err = clientset.CoreV1().Pods(namespaces[i]).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				skipIteration = true
				log.Println(fmt.Sprintf("ERROR: unable to get all pods in namespace %s: %s", namespaces[i], err.Error()))
			}
		}
		if !skipIteration {
			// Loop through all pods
			for i := 0; i < len(podLists); i++ {
				for j := 0; j < len(podLists[i].Items); j++ {

					// Check that we're not connecting to this pod, and also that the pod has the label 'type: target'
					if (podLists[i].Items[j].ObjectMeta.Name != podName || podLists[i].Items[j].ObjectMeta.Namespace != podNamespace) && podLists[i].Items[j].ObjectMeta.Labels["type"] == "target" {

						// Loop through all given ports
						for k := 0; k < len(portsToConnectOn); k++ {

							log.Println(fmt.Sprintf("Sending request to %s on port %d", podLists[i].Items[j].Status.PodIP, portsToConnectOn[k]))

							// Attempt to connect to the pod
							resp, err := http.Get(fmt.Sprintf("http://%s:%d/connect", string(podLists[i].Items[j].Status.PodIP), portsToConnectOn[k]))

							if err == nil { // is successful, print out the response
								log.Println(resp)
							} else { // print out the error message if unsuccessful
								log.Println(fmt.Sprintf("Unable to connect to pod %s in namespace %s on port %d: %s", podLists[i].Items[j].ObjectMeta.Name, podLists[i].Items[j].ObjectMeta.Namespace, portsToConnectOn[k], err.Error()))
							}
						}
					}

				}
			}
		}

		time.Sleep(10 * time.Second)
	}

}

func startServer(listeningPort int) {
	handleFunc := func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(fmt.Sprintf("Successfully connected to pod %s in namespace %s on port %s!\n", podName, podNamespace, req.URL.Port())))
	}

	log.Println(fmt.Sprintf("Listening for requests at http://localhost:%d/connect", listeningPort))

	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", listeningPort), http.HandlerFunc(handleFunc))
}

func main() {
	for i := 0; i < len(portsToConnectOn); i++ {
		go startServer(portsToConnectOn[i])
	}

	go startPodConnectionParty()

	// Block the main thread until the process ends
	<-make(chan int)
}
