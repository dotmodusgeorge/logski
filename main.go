package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func CreateClient(kubeconfig *string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientset
}

type PodItem struct {
	Name string
	Time v1.Time
}

func GetPods(c *kubernetes.Clientset, namespace string, limit int, filter string) []PodItem {
	podList := make([]PodItem, 0)
	pods, err := c.CoreV1().Pods(namespace).List(v1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, pod := range pods.Items {
		if filter != "" {
			if matched, _ := regexp.MatchString(".*"+filter+".*", pod.Name); matched {
				podList = append(podList, PodItem{Name: pod.Name, Time: pod.ObjectMeta.CreationTimestamp})
			}
			continue
		}
		podList = append(podList)
	}

	if filter != "" {
		filteredList := make([]PodItem, 0)
		for i := 0; i < len(podList); i++ {
			podItem := podList[i]
			if matched, _ := regexp.MatchString(".*"+filter+".*", podItem.Name); matched {
				filteredList = append(filteredList, podItem)
			}
		}
		podList = filteredList
	}

	sort.Slice(podList, func(i, j int) bool {
		return podList[i].Time.Unix() > podList[j].Time.Unix()
	})
	if len(podList) < 20 {
		return podList
	}
	return podList[0:limit]
}

func getPodLogs(c *kubernetes.Clientset, namespace string, name string) string {
	podLogOpts := corev1.PodLogOptions{}
	req := c.CoreV1().Pods(namespace).GetLogs(name, &podLogOpts)
	podLogs, err := req.Stream()
	if err != nil {
		return "error in opening stream"
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "error in copy information from podLogs to buf"
	}
	str := buf.String()

	return str
}

func main() {
	fmt.Println("<< Logski >>\n")
	fmt.Println("Config->")
	fmt.Printf("\tPID::%d\n", os.Getpid())

	namespace := flag.String("n", "default", "the k8s namespace to access")
	filter := flag.String("f", "", "the filter for pods")
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	fmt.Printf("\tNAMESPACE::%s\n", *namespace)
	fmt.Printf("\tFILTER::%s\n\n", *filter)

	client := CreateClient(kubeconfig)
	for {
		pods := GetPods(client, *namespace, 20, *filter)
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
		logs := getPodLogs(client, *namespace, podItem.Name)
		fmt.Println(logs)
	}
}
