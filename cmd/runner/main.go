package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"logski/handler"
)


func main() {
	fmt.Println("<< Logski >>\n")
	fmt.Println("Config->")
	fmt.Printf("\tPID::%d\n", os.Getpid())

	namespace := flag.String("n", "default", "the k8s namespace to access")
	filter := flag.String("f", "", "the filter for pods")

	client := handling.CreateClient()
	for {
		pods := handling.GetPods(client, *namespace, 20, *filter)
		if len(pods) == 0 {
			fmt.Println("--> No pods found")
			fmt.Println("--> Trying again in 3 seconds")
			time.Sleep(1 * time.Second)
			fmt.Println("--> Trying again in 2 seconds")
			time.Sleep(1 * time.Second)
			fmt.Println("--> Trying again in 1 second")
			time.Sleep(1 * time.Second)
			continue
		}
		fmt.Printf("Latest Pods in %s\n", *namespace)
		for index, pod := range pods {
			fmt.Printf("%d. %s | %s \n", index+1, pod.Name, pod.Time.Format(time.RFC3339))
		}
		fmt.Println("Which Pod do you want to see logs for? EG: 1")
		reader := bufio.NewReader(os.Stdin)
		choiceRaw, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		choice := strings.ReplaceAll(choiceRaw, "\n", "")
		podID, err := strconv.Atoi(choice)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("Please ensure you enter a valid number! Input received: %s", choice)
			continue
		}

		if podID > 20 || podID < 1 {
			fmt.Printf("Please ensure you provide a number between 1 and 20 Input received: %s", choice)
			continue
		}

		podItem := pods[podID-1]
		logs := handling.GetPodLogs(client, *namespace, podItem.Name)
		fmt.Println(logs)
	}
}
