package e2e

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/onsi/ginkgo"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

		ginkgo.By("creating a backend pod for the service on each node" + serviceName)
		err = jig.CreateServicePods(len(nodes.Items))
		framework.ExpectNoError(err)

		execPod := e2epod.CreateExecPodOrFail(cs, namespace, "execpod-affinity", nil)
		defer func() {
			framework.Logf("Cleaning up the exec pod")
			err := cs.CoreV1().Pods(namespace).Delete(context.TODO(), execPod.Name, metav1.DeleteOptions{})
			framework.ExpectNoError(err, "failed to delete pod: %s in namespace: %s", execPod.Name, namespace)
		}()
		err = jig.CheckServiceReachability(svc, execPod)
		framework.ExpectNoError(err)

		hosts := map[string]int{}
		cmd := fmt.Sprintf(`curl -q -s --connect-timeout 2 http://%s/`, net.JoinHostPort(svcIP, strconv.Itoa(port)))
		for i := 0; i < 100; i++ {
			hostname, err := framework.RunHostCmd(execPod.Namespace, execPod.Name, cmd)
			if err == nil {
				framework.Logf("Connected to hostname %s", hostname)
				hosts[hostname]++
			}
		}
		framework.Logf("Connections from %v distributed %v", execPod.Name, hosts)

	})

	ginkgo.It("Services without topology annotation should forward traffic to nodes on all zones", func() {
		namespace := f.Namespace.Name
		serviceName := "svc-topology"
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

		ginkgo.By("creating a backend pod for the service on each node" + serviceName)
		err = jig.CreateServicePods(len(nodes.Items))
		framework.ExpectNoError(err)

		execPod := e2epod.CreateExecPodOrFail(cs, namespace, "execpod-affinity", nil)
		defer func() {
			framework.Logf("Cleaning up the exec pod")
			err := cs.CoreV1().Pods(namespace).Delete(context.TODO(), execPod.Name, metav1.DeleteOptions{})
			framework.ExpectNoError(err, "failed to delete pod: %s in namespace: %s", execPod.Name, namespace)
		}()
		err = jig.CheckServiceReachability(svc, execPod)
		framework.ExpectNoError(err)

		hosts := map[string]int{}
		cmd := fmt.Sprintf(`curl -q -s --connect-timeout 2 http://%s/`, net.JoinHostPort(svcIP, strconv.Itoa(port)))
		for i := 0; i < 100; i++ {
			hostname, err := framework.RunHostCmd(execPod.Namespace, execPod.Name, cmd)
			if err == nil {
				framework.Logf("Connected to hostname %s", hostname)
				hosts[hostname]++
			}
		}
		framework.Logf("Connections from %s distributed %v", execPod.Name, hosts)

	})

})
