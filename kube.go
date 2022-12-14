package main

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"os"
	"path/filepath"
)

func getCubeConfig(devMode bool) (*rest.Config, error) {
	if devMode {
		var kubeconfigFile = os.Getenv("kubeconfigPath")
		kubeConfigPath := filepath.Join(kubeconfigFile)
		klog.Infof("Using kubeconfig: %s\n", kubeConfigPath)

		kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			klog.Error("error getting Kubernetes config: %v\n", err)
			os.Exit(1)
		}

		return kubeConfig, nil
	} else {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}

		return config, nil
	}
}

func createKubeConfig() (*kubernetes.Clientset, client.Client) {

	config, err := getCubeConfig(devMode)
	if err != nil {
		klog.Info("Config init error...", err)
		os.Exit(1)
	}
	forConfig, err := kubernetes.NewForConfig(config)
	c, _ := client.New(config, client.Options{})
	if err != nil {
		klog.Info("Config init error...", err)
		os.Exit(1)
	}
	return forConfig, c
}

func loadReposFromKube() map[string]string {

	configmap, err := clientSet.CoreV1().ConfigMaps(NameSpace).Get(context.Background(), ConfigmapName, metav1.GetOptions{})
	if err != nil {
		klog.Error("error get configmap: %v\n", err)
	}

	return configmap.Data
}

func saveReposToKube(repositories map[string]string) {
	ctx := context.TODO()

	configMap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConfigmapName,
			Namespace: NameSpace,
		},
		Data: repositories,
	}

	_, err := clientSet.CoreV1().ConfigMaps(NameSpace).Update(ctx, &configMap, metav1.UpdateOptions{})
	if err != nil {
		klog.Error("error saving configmap: %v\n", err)
	}
}
