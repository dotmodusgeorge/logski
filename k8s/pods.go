package k8s

import (
	"bytes"
	"io"
	"sort"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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

	sort.Slice(podList, func(i, j int) bool {
		return podList[i].Time.Unix() > podList[j].Time.Unix()
	})
	if len(podList) < 20 {
		return podList
	}

	if limit <= 0 {
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

func GetNameSpaces(c *kubernetes.Clientset, filter string) ([]string, error) {
	namespaces, err := c.CoreV1().Namespaces().List(v1.ListOptions{})
	if (err != nil) {
		panic(err)
	}
	var namespaceStrings []string
	for _, namespace := range namespaces.Items {
		if filter != "" {
			if matched, _ := regexp.MatchString(".*"+filter+".*", namespace.Name); matched {
				namespaceStrings = append(namespaceStrings, namespace.Name)
			}
			continue
		} 
		namespaceStrings = append(namespaceStrings, namespace.Name)
	}

	return namespaceStrings, nil
}
