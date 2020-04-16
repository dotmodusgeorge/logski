package main

import (
	"flag"
	"os"
	"fmt"
	"time"
	"logski/handler"
)

func printFlagSetDefaults(command flag.FlagSet, name string) {
	fmt.Println("---" + name + "---")
	command.PrintDefaults()
}

func main() {
	namespace := flag.String("namespace", "default", "Add a namespace other than the default one.")

	podCommand := flag.NewFlagSet("pods", flag.ExitOnError)
	logCommand := flag.NewFlagSet("logs", flag.ExitOnError)
	client := handling.CreateClient()

	filter := podCommand.String("filter", "", "Adds wildcarded filter to to the pods list")
	limit := podCommand.Int("limit", 20, "Adds a limit to the amount of pods gotten")

	podName := logCommand.String("pod_name", "", "The name of the pod for to getthe log from (required)")

	switch os.Args[1] {
	case "pods":
		podCommand.Parse(os.Args[2:])
	case "logs":
		logCommand.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		printFlagSetDefaults(*podCommand, "pod")
		printFlagSetDefaults(*logCommand, "logs")
		os.Exit(1)
	}
	if (podCommand.Parsed()) {
		pods := handling.GetPods(client, *namespace, *limit, *filter)
		fmt.Printf("Latest Pods in %s\n", *namespace)
		for index, pod := range pods {
			fmt.Printf("%d. %s | %s \n", index+1, pod.Name, pod.Time.Format(time.RFC3339))
		}
	} else if (logCommand.Parsed()) {
		if (*podName == "") {
			logCommand.PrintDefaults()
			os.Exit(1)
		}
		logs := handling.GetPodLogs(client, *namespace, *podName)
		fmt.Println(logs)
	}
}
