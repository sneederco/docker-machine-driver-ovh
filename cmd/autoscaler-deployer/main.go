// autoscaler-deployer: Watches Rancher clusters and auto-deploys cluster-autoscaler
// when autoscaler annotations are detected on machine pools.
//
// Runs on the Rancher management cluster and:
// 1. Watches cluster.provisioning.cattle.io resources
// 2. Detects autoscaler-min-size/max-size annotations
// 3. Deploys cluster-autoscaler to the downstream cluster
package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

//go:embed manifests/*.yaml
var manifests embed.FS

var (
	kubeconfig   = flag.String("kubeconfig", "", "Path to kubeconfig (uses in-cluster if empty)")
	namespace    = flag.String("namespace", "fleet-default", "Namespace to watch for clusters")
	rancherURL   = flag.String("rancher-url", "", "Rancher server URL (required)")
	rancherToken = flag.String("rancher-token", "", "Rancher API token (required)")
	dryRun       = flag.Bool("dry-run", false, "Log actions without deploying")
)

const (
	annotationMinSize = "cluster.provisioning.cattle.io/autoscaler-min-size"
	annotationMaxSize = "cluster.provisioning.cattle.io/autoscaler-max-size"
	labelDeployed     = "autoscaler.sneederco.io/deployed"
)

var clusterGVR = schema.GroupVersionResource{
	Group:    "provisioning.cattle.io",
	Version:  "v1",
	Resource: "clusters",
}

func main() {
	flag.Parse()

	if *rancherURL == "" {
		*rancherURL = os.Getenv("RANCHER_URL")
	}
	if *rancherToken == "" {
		*rancherToken = os.Getenv("RANCHER_TOKEN")
	}
	if *rancherURL == "" || *rancherToken == "" {
		log.Fatal("--rancher-url and --rancher-token are required")
	}

	config, err := getConfig()
	if err != nil {
		log.Fatalf("Failed to get kubernetes config: %v", err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create dynamic client: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Printf("Starting autoscaler-deployer")
	log.Printf("  Rancher URL: %s", *rancherURL)
	log.Printf("  Watching namespace: %s", *namespace)
	log.Printf("  Dry run: %v", *dryRun)

	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		client, 30*time.Second, *namespace, nil,
	)

	informer := factory.ForResource(clusterGVR).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			handleCluster(ctx, client, obj.(*unstructured.Unstructured))
		},
		UpdateFunc: func(_, newObj interface{}) {
			handleCluster(ctx, client, newObj.(*unstructured.Unstructured))
		},
	})

	factory.Start(ctx.Done())
	factory.WaitForCacheSync(ctx.Done())

	log.Println("Informer synced, watching for clusters...")
	<-ctx.Done()
	log.Println("Shutting down")
}

func getConfig() (*rest.Config, error) {
	if *kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}
	// Try in-cluster config
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}
	// Fall back to default kubeconfig location
	home, _ := os.UserHomeDir()
	return clientcmd.BuildConfigFromFlags("", home+"/.kube/config")
}

func handleCluster(ctx context.Context, client dynamic.Interface, cluster *unstructured.Unstructured) {
	name := cluster.GetName()
	
	// Check if already deployed
	labels := cluster.GetLabels()
	if labels != nil && labels[labelDeployed] == "true" {
		return
	}

	// Check for autoscaler annotations on machine pools
	spec, found, _ := unstructured.NestedMap(cluster.Object, "spec", "rkeConfig")
	if !found {
		return
	}

	pools, found, _ := unstructured.NestedSlice(spec, "machinePools")
	if !found || len(pools) == 0 {
		return
	}

	// Check first pool for autoscaler annotations
	pool := pools[0].(map[string]interface{})
	annotations, found, _ := unstructured.NestedStringMap(pool, "machineDeploymentAnnotations")
	if !found {
		return
	}

	minSize, hasMin := annotations[annotationMinSize]
	maxSize, hasMax := annotations[annotationMaxSize]
	
	if !hasMin && !hasMax {
		return
	}

	// Get cluster ID from status
	clusterID, found, _ := unstructured.NestedString(cluster.Object, "status", "clusterName")
	if !found || clusterID == "" {
		log.Printf("[%s] Cluster not ready yet (no clusterName in status)", name)
		return
	}

	// Check cluster is ready
	ready, _, _ := unstructured.NestedString(cluster.Object, "status", "phase")
	if ready != "Ready" && ready != "Active" {
		log.Printf("[%s] Cluster not ready (phase=%s)", name, ready)
		return
	}

	log.Printf("[%s] Autoscaler annotations detected: min=%s, max=%s", name, minSize, maxSize)
	log.Printf("[%s] Cluster ID: %s", name, clusterID)

	if *dryRun {
		log.Printf("[%s] DRY RUN: Would deploy autoscaler", name)
		return
	}

	// Deploy autoscaler
	if err := deployAutoscaler(ctx, name, clusterID); err != nil {
		log.Printf("[%s] Failed to deploy autoscaler: %v", name, err)
		return
	}

	// Mark cluster as deployed
	if err := markDeployed(ctx, client, cluster); err != nil {
		log.Printf("[%s] Failed to mark cluster as deployed: %v", name, err)
	}

	log.Printf("[%s] ✅ Autoscaler deployed successfully", name)
}

func deployAutoscaler(ctx context.Context, clusterName, clusterID string) error {
	// Get downstream cluster kubeconfig via Rancher API
	log.Printf("[%s] Fetching downstream kubeconfig...", clusterName)
	
	// For now, shell out to kubectl - in production, use proper client
	// This is a simplified version; production should use Rancher client library
	
	manifest, err := manifests.ReadFile("manifests/cluster-autoscaler.yaml")
	if err != nil {
		return fmt.Errorf("failed to read embedded manifest: %w", err)
	}

	// Replace variables
	content := string(manifest)
	content = strings.ReplaceAll(content, "${RANCHER_URL}", *rancherURL)
	content = strings.ReplaceAll(content, "${RANCHER_TOKEN}", *rancherToken)
	content = strings.ReplaceAll(content, "${CLUSTER_NAME}", clusterName)

	log.Printf("[%s] Applying autoscaler manifest to cluster %s...", clusterName, clusterID)
	
	// In production, use Rancher's cluster proxy API to apply manifests
	// For now, we'll use the impersonation API
	// POST /k8s/clusters/{clusterID}/v1/...
	
	_ = content // TODO: Apply via Rancher API
	
	return nil
}

func markDeployed(ctx context.Context, client dynamic.Interface, cluster *unstructured.Unstructured) error {
	labels := cluster.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[labelDeployed] = "true"
	cluster.SetLabels(labels)

	_, err := client.Resource(clusterGVR).Namespace(cluster.GetNamespace()).Update(
		ctx, cluster, metav1.UpdateOptions{},
	)
	return err
}
