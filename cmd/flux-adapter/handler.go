package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "net/http/pprof"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"

	"github.com/fluxcd/flux/pkg/event"
	"github.com/fluxcd/flux/pkg/resource"
)

type eventRecorder struct {
	// Map from Kind (lowercased) to server preferred GroupVersion, e.g. deployment -> apps/v1
	groupVersions map[string]string
	record.EventRecorder
}

func (recorder *eventRecorder) emitK8sEvent(body []byte) error {
	var e event.Event
	if err := json.NewDecoder(bytes.NewBuffer(body)).Decode(&e); err != nil {
		return err
	}

	switch e.Type {
	case "sync":
		message := ""
		em := e.Metadata.(*event.SyncEventMetadata)
		if len(em.Commits) > 0 {
			message += fmt.Sprintf("Commit %.12s: %s", em.Commits[0].Revision, em.Commits[0].Message)
		}
		if len(em.Commits) > 1 {
			message += " ..."
		}
		metadataJSON, _ := json.Marshal(e.Metadata)
		annotations := map[string]string{
			"syncMetadata": string(metadataJSON),
		}
		fmt.Printf("sending %d events\n", len(e.ServiceIDs))
		// Generate one Kubernetes event for each resource affected
		for _, id := range e.ServiceIDs {
			objRef := recorder.idToObjectRef(id)
			recorder.AnnotatedEventf(objRef, annotations, "Normal", "Sync", "%s", message)
		}
	default:
		eventJSON, _ := json.Marshal(&e)
		fmt.Printf("not handling: %s\n", string(eventJSON))
	}
	return nil
}

func (recorder *eventRecorder) idToObjectRef(id resource.ID) *corev1.ObjectReference {
	namespace, kind, name := id.Components()
	apiVersion := recorder.groupVersions[strings.ToLower(kind)]
	if namespace == "" {
		namespace = "default"
	}
	return &corev1.ObjectReference{
		APIVersion: apiVersion,
		Namespace:  namespace,
		Kind:       kind,
		Name:       name,
	}
}

func createK8sEventRecorder() (*eventRecorder, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	// connect to Kubernetes server
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to init clientSet: %w", err)
	}

	// Get all object Kinds supported by the server so we know the preferred GroupVersion
	_, resourceList, err := client.DiscoveryClient.ServerGroupsAndResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get server resource list: %w", err)
	}

	scheme := runtime.NewScheme()
	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: client.CoreV1().Events("")})
	handler := &eventRecorder{
		groupVersions: makeGVMap(resourceList),
		EventRecorder: broadcaster.NewRecorder(scheme, corev1.EventSource{Component: "flux"}),
	}
	return handler, nil
}

func makeGVMap(resourceList []*metav1.APIResourceList) map[string]string {
	groupVersions := make(map[string]string, len(resourceList))
	for _, apiResourceList := range resourceList {
		for _, apiResource := range apiResourceList.APIResources {
			groupVersions[strings.ToLower(apiResource.Kind)] = apiResourceList.GroupVersion
		}
	}
	return groupVersions
}
