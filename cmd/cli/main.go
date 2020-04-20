package main

import (
	"io/ioutil"
	"flag"
	"fmt"
	"logski/k8s"
	"os"
	"time"
)

func main() {
	namespace := flag.String("namespaces", "default", "Add a namespace other than the default one.")
	
	nameSpaceCommand := flag.NewFlagSet("namespaces", flag.ExitOnError)
	podCommand := flag.NewFlagSet("pods", flag.ExitOnError)
	logCommand := flag.NewFlagSet("logs", flag.ExitOnError)
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Println("  namespaces string \n\tBase command for the 'namespaces' command.")
		fmt.Println("  pods string \n\tBase command for the 'pods' command.")
		fmt.Println("  logs string \n\tBase command for the 'logs' command.")
	}
	client := k8s.CreateClient()

	nameSpaceFilter := nameSpaceCommand.String("filter", "", "Adds wildcarded filter to the namespace list")

	filter := podCommand.String("filter", "", "Adds wildcarded filter to to the pods list")
	limit := podCommand.Int("limit", 20, "Adds a limit to the amount of pods gotten")

	podName := logCommand.String("pod_name", "", "The name of the pod for to getthe log from (required)")
	outPut := logCommand.String("output", "", "The file that the logs will be outputed into")

	if (len(os.Args) < 2) {
		flag.Usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "pods":
		podCommand.Parse(os.Args[2:])
	case "logs":
		logCommand.Parse(os.Args[2:])
	case "namespaces":
		nameSpaceCommand.Parse(os.Args[2:])
	default:
		flag.Usage()
		os.Exit(1)
	}
	if podCommand.Parsed() {
		pods := k8s.GetPods(client, *namespace, *limit, *filter)
		fmt.Printf("Latest Pods in %s\n", *namespace)
		for index, pod := range pods {
			fmt.Printf("%d. %s | %s \n", index+1, pod.Name, pod.Time.Format(time.RFC3339))
		}
	} else if logCommand.Parsed() {
		if *podName == "" {
			logCommand.PrintDefaults()
			os.Exit(1)
		}
		logs := k8s.GetPodLogs(client, *namespace, *podName)

		if (*outPut != "") {
			data := []byte(logs)
			err := ioutil.WriteFile(*outPut, data, 0644)
			if (err != nil) {
				panic(err)
			}
		} else {
			fmt.Println(logs)
		}
	} else if nameSpaceCommand.Parsed() {
		nameSpaces, err := k8s.GetNameSpaces(client, *nameSpaceFilter)
		if (err != nil) {
			panic(err)
		} 
		for i, namespace := range nameSpaces {
			fmt.Println(fmt.Sprintf("%d %s", i, namespace))
		}
	}
}
