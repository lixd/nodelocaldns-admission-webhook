package pkg

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NodeLocalDNSInjection = "node-local-dns-injection"
)

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
}

func StringPtr(s string) *string {
	return &s
}

func loadCustomDnsConfig(namespace string, config Config) (corev1.DNSPolicy, *corev1.PodDNSConfig) {
	nsSvc := fmt.Sprintf("%s.svc.cluster.local", namespace)
	return "None", &corev1.PodDNSConfig{
		Nameservers: []string{config.LocalDNS, config.KubeDNS},
		Searches:    []string{nsSvc, "svc.cluster.local", "cluster.local"},
		Options: []corev1.PodDNSConfigOption{
			{
				Name:  "ndots",
				Value: StringPtr("3"),
			},
			{
				Name:  "attempts",
				Value: StringPtr("2"),
			},
			{
				Name:  "timeout",
				Value: StringPtr("1"),
			},
		},
	}
}

// NeedMutation Check whether the target resoured need to be mutated
func (a *PodAnnotator) NeedMutation(pod *corev1.Pod) bool {
	if pod.Namespace == "" {
		pod.Namespace = "default"
	}
	/*
	   Pod will automatically inject DNS cache when all of the following conditions are met:
	   1. The newly created Pod is not in the kube-system and kube-public namespaces.
	   2. The Labels of the namespace where the new Pod is located contain node-local-dns-injection=enabled.
	   3. The newly created Pod is not labeled with the disabled DNS injection node-local-dns-injection=disabled label.
	   4. The network of the newly created Pod is hostNetwork and DNSPolicy is ClusterFirstWithHostNet, or the Pod is non-hostNetwork and DNSPolicy is ClusterFirst.
	*/
	//1. The newly created Pod is not in the kube-system and kube-public namespaces.
	for _, namespace := range ignoredNamespaces {
		if pod.Namespace == namespace {
			klog.V(1).Infof("Skip mutation for %v for it's in special namespace: %v", pod.Name, pod.Namespace)
			return false
		}
	}

	// Fetch the namespace where the Pod is located.
	var ns corev1.Namespace
	err := a.Client.Get(context.Background(), client.ObjectKey{Name: pod.GetNamespace()}, &ns)
	if err != nil {
		klog.V(1).ErrorS(err, "Failed to fetch namespace: %v", pod.Namespace)
		return false
	}

	//2. The Labels of the namespace where the new Pod is located contain node-local-dns-injection=enabled.
	if v, ok := ns.Labels[NodeLocalDNSInjection]; !ok || v != "enabled" {
		return false
	}

	//3. The newly created Pod is not labeled with the disabled DNS injection node-local-dns-injection=disabled label.
	if v, ok := pod.Labels[NodeLocalDNSInjection]; ok && v == "disabled" {
		return false
	}

	//4. The network of the newly created Pod is hostNetwork and DNSPolicy is ClusterFirstWithHostNet, or the Pod is non-hostNetwork and DNSPolicy is ClusterFirst.
	// The network of the Pod is hostNetwork, so DNSPolicy should be ClusterFirstWithHostNet.
	if pod.Spec.HostNetwork && pod.Spec.DNSPolicy != corev1.DNSClusterFirstWithHostNet {
		return false
	}

	// The network of the Pod is not hostNetwork, so DNSPolicy should be ClusterFirst.
	if !pod.Spec.HostNetwork && pod.Spec.DNSPolicy != corev1.DNSClusterFirst {
		return false
	}

	// If all conditions are met, return true.
	return true
}

func mutation(pod *corev1.Pod, conf Config) {
	ns := pod.Namespace
	if ns == "" {
		ns = "default"
	}
	pod.Spec.DNSPolicy, pod.Spec.DNSConfig = loadCustomDnsConfig(ns, conf)
}
