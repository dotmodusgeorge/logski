package handling

import (
	"bytes"
	"io"
	"regexp"
	"sort"
	"path/filepath"
	"flag"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func CreateClient() *kubernetes.Clientset {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
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
		podList = append(podList, PodItem{Name: pod.Name, Time: pod.ObjectMeta.CreationTimestamp})
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

	if (limit <= 0) {
		return podList
	}
	return podList[0:limit]
}

func GetPodLogs(c *kubernetes.Clientset, namespace string, name string) string {
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
