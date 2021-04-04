package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	e2enode "k8s.io/kubernetes/test/e2e/framework/node"
	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"
	e2eservice "k8s.io/kubernetes/test/e2e/framework/service"
)

var _ = ginkgo.Describe("Topology Aware Hints", func() {

	f := framework.NewDefaultFramework("topology")

	var cs clientset.Interface

	ginkgo.BeforeEach(func() {
		cs = f.ClientSet
	})

	ginkgo.It("Services with topology annotation should forward traffic to nodes on the same zone", func() {
		namespace := f.Namespace.Name
		serviceName := "svc-topology"
		port := 80

		jig := e2eservice.NewTestJig(cs, namespace, serviceName)
		nodes, err := e2enode.GetBoundedReadySchedulableNodes(cs, e2eservice.MaxNodesForEndpointsTests)
		framework.ExpectNoError(err)

		ginkgo.By("creating an annotated service with no endpoints and topology aware annotation")
		svc, err := jig.CreateTCPServiceWithPort(func(svc *v1.Service) {
			svc.Annotations = map[string]string{"service.kubernetes.io/topology-aware-hints": "auto"}
			svc.Spec.Ports = []v1.ServicePort{
				{Port: int32(port), Name: "http", Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(9376)},
			}
		}, int32(port))
		framework.ExpectNoError(err)
		svcIP := svc.Spec.ClusterIP

		ginkgo.By("creating backend pods for the service on each node" + serviceName)
		// hints are only generated if there is a reasonable distribution of endpoints
		err = jig.CreateServicePods(3 * len(nodes.Items))
		framework.ExpectNoError(err)

		// check hints are created
		opts := metav1.ListOptions{
			LabelSelector: "kubernetes.io/service-name=" + serviceName,
		}

		// map that contains the zone associated to each pod
		podsZones := map[string]string{}
		err = wait.PollImmediate(3*time.Second, 30*time.Second, func() (bool, error) {
			es, err := cs.DiscoveryV1().EndpointSlices(namespace).List(context.TODO(), opts)
			if err != nil {
				framework.Logf("Failed go list EndpointSlice objects: %v", err)
				// Retry the error
				return false, nil
			}
			for _, endpointSlice := range es.Items {
				for _, ep := range endpointSlice.Endpoints {
					if ep.Hints == nil {
						return false, nil
					}
					podsZones[ep.TargetRef.Name] = *ep.Zone
					framework.Logf("pod %s on zone %s with hints %v", ep.TargetRef.Name, *ep.Zone, ep.Hints)
				}
			}
			return true, nil

		})
		framework.ExpectNoError(err)

		// select one node
		nodeName := nodes.Items[0].Name
		nodeZone := nodes.Items[0].Labels["topology.kubernetes.io/zone"]
		execPod := e2epod.CreateExecPodOrFail(cs, namespace, "execpod-affinity", func(pod *v1.Pod) {
			pod.Spec.NodeName = nodeName
		})
		defer func() {
			framework.Logf("Cleaning up the exec pod")
			err := cs.CoreV1().Pods(namespace).Delete(context.TODO(), execPod.Name, metav1.DeleteOptions{})
			framework.ExpectNoError(err, "failed to delete pod: %s in namespace: %s", execPod.Name, namespace)
		}()
		err = jig.CheckServiceReachability(svc, execPod)
		framework.ExpectNoError(err)

		zones := map[string]int{}
		cmd := fmt.Sprintf(`echo hostName | nc -v -w 5 %s %d`, svcIP, port)
		for i := 0; i < 100; i++ {
			hostname, err := framework.RunHostCmd(execPod.Namespace, execPod.Name, cmd)
			if err == nil && hostname != "" {
				z, ok := podsZones[hostname]
				if !ok {
					framework.Failf("hostname %s not in any zone", hostname)
				}
				zones[z]++
			}
		}
		framework.Logf("Connections from %v distributed with Topology %v", nodeZone, zones)

	})

	ginkgo.It("Services without topology annotation should forward traffic to nodes on all zones", func() {
		namespace := f.Namespace.Name
		serviceName := "svc-no-topology"
		port := 80

		jig := e2eservice.NewTestJig(cs, namespace, serviceName)
		nodes, err := e2enode.GetBoundedReadySchedulableNodes(cs, e2eservice.MaxNodesForEndpointsTests)
		framework.ExpectNoError(err)

		ginkgo.By("creating an annotated service with no endpoints and topology aware annotation")
		svc, err := jig.CreateTCPServiceWithPort(func(svc *v1.Service) {
			svc.Spec.Ports = []v1.ServicePort{
				{Port: int32(port), Name: "http", Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(9376)},
			}
		}, int32(port))
		framework.ExpectNoError(err)
		svcIP := svc.Spec.ClusterIP

		ginkgo.By("creating backend pods for the service on each node" + serviceName)
		// hints are only generated if there is a reasonable distribution of endpoints
		err = jig.CreateServicePods(3 * len(nodes.Items))
		framework.ExpectNoError(err)
		// map that contains the zone associated to each pod
		podsZones := map[string]string{}
		// check hints are created
		opts := metav1.ListOptions{
			LabelSelector: "kubernetes.io/service-name=" + serviceName,
		}
		err = wait.PollImmediate(3*time.Second, 30*time.Second, func() (bool, error) {
			es, err := cs.DiscoveryV1().EndpointSlices(namespace).List(context.TODO(), opts)
			if err != nil {
				framework.Logf("Failed go list EndpointSlice objects: %v", err)
				// Retry the error
				return false, nil
			}
			for _, endpointSlice := range es.Items {
				for _, ep := range endpointSlice.Endpoints {
					if ep.Hints != nil {
						return false, nil
					}
					podsZones[ep.TargetRef.Name] = *ep.Zone
					framework.Logf("pod %s on zone %s", ep.TargetRef.Name, *ep.Zone)
				}
			}
			return true, nil
		})
		// select one node
		nodeName := nodes.Items[0].Name
		nodeZone := nodes.Items[0].Labels["topology.kubernetes.io/zone"]
		execPod := e2epod.CreateExecPodOrFail(cs, namespace, "execpod-affinity", func(pod *v1.Pod) {
			pod.Spec.NodeName = nodeName
		})
		defer func() {
			framework.Logf("Cleaning up the exec pod")
			err := cs.CoreV1().Pods(namespace).Delete(context.TODO(), execPod.Name, metav1.DeleteOptions{})
			framework.ExpectNoError(err, "failed to delete pod: %s in namespace: %s", execPod.Name, namespace)
		}()
		err = jig.CheckServiceReachability(svc, execPod)
		framework.ExpectNoError(err)
		zones := map[string]int{}
		cmd := fmt.Sprintf(`echo hostName | nc -v -w 5 %s %d`, svcIP, port)
		for i := 0; i < 100; i++ {
			hostname, err := framework.RunHostCmd(execPod.Namespace, execPod.Name, cmd)
			if err == nil && hostname != "" {
				framework.Logf("Connected to hostname %s", hostname)
				z := podsZones[hostname]
				zones[z]++
			}
		}
		framework.Logf("Connections from %s distributed without topology %v", nodeZone, zones)

	})

})
