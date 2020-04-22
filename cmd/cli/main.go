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
	
	podCommand := flag.NewFlagSet("pods", flag.ExitOnError)
	logCommand := flag.NewFlagSet("logs", flag.ExitOnError)
	flag.Usage = func() {
		flag.PrintDefaults()
		fmt.Println("  pods string \n\tBase command for the 'pods' command.")
		fmt.Println("  logs string \n\tBase command for the 'logs' command.")
	}
	client := k8s.CreateClient()

	filter := podCommand.String("filter", "", "Adds wildcarded filter to to the pods list")
	limit := podCommand.Int("limit", 20, "Adds a limit to the amount of pods gotten")

	podName := logCommand.String("pod_name", "", "The name of the pod for to getthe log from (required)")
	outPut := logCommand.String("output", "", "The file that the logs will be outputed into")

	if (len(os.Args) < 2) {
		flag.Usage()
		os.Exit(1)
	}

	var namespace *string
	switch os.Args[1] {
	case "pods":
		namespace = podCommand.String("namespace", "default", "Add a namespace other than the default one.")
		podCommand.Parse(os.Args[2:])
	case "logs":
		namespace = logCommand.String("namespace", "default", "Add a namespace other than the default one.")
		logCommand.Parse(os.Args[2:])
	default:
		flag.Usage()
		os.Exit(1)
	}

	nameSpaces, err := k8s.GetNameSpaces(client, *namespace)
	if (err != nil) {
		panic(err)
	}
	namespaceString := nameSpaces[0]

	fmt.Printf("Using namespace: %s", namespaceString)

	if podCommand.Parsed() {
		pods := k8s.GetPods(client, namespaceString, *limit, *filter)
		fmt.Printf("Latest Pods in %s\n", *namespace)
		if (len(pods) > 0) {
			for index, pod := range pods {
				fmt.Printf("%d. %s | %s \n", index+1, pod.Name, pod.Time.Format(time.RFC3339))
			}
		} else {
			fmt.Println("\nThere are no pods\n")
		}
	} else if logCommand.Parsed() {
		if *podName == "" {
			logCommand.PrintDefaults()
			os.Exit(1)
		}
		logs := k8s.GetPodLogs(client, namespaceString, *podName)

		if (*outPut != "") {
			data := []byte(logs)
			err := ioutil.WriteFile(*outPut, data, 0644)
			if (err != nil) {
				panic(err)
			}
		} else {
			fmt.Println(logs)
		}
	}
}

// ddsnetwork, adobe, dcmcampaignreach, dcmsitereach, ddsnetwork, prisma
