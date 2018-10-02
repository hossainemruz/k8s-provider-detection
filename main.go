package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/the-redback/go-oneliners"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeConfig string
)

func init() {
	//flag.StringVar(&kubeConfig, "kubeconfig", filepath.Join(homedir.HomeDir(),".kube/config"), "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&kubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

// kubedb-98dc2725.hcp.centralus.azmk8s.io

func main() {
	flag.Parse()
	log.Println("kubeconfig: ", kubeConfig, "masterURL: ", masterURL)
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		log.Fatalf("can't build config. Reason: %v", err.Error())
	}

	err = rest.LoadTLSFiles(cfg)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(string(cfg.CAData))

	//kubeClient := kubernetes.NewForConfigOrDie(cfg)
	//nodes,err:=kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	//if err!=nil{
	//	log.Println(err)
	//	return
	//}
	//oneliners.PrettyJson(nodes,"Nodes")

	// create ca cert pool
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(cfg.CAData)

	if !ok {
		log.Fatalf("Can't append caCert to caCertPool.")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: caCertPool},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get("https://kubernetes.default.svc")
	if err != nil {
		log.Println(err)
		return
	}
	if !resp.TLS.HandshakeComplete {
		log.Fatalln("failed to complete TLS handshake.")
	}

	var addr string
DONE:
	for _, chain := range resp.TLS.VerifiedChains {
		if len(chain) == 0 {
			continue
		}
		for _, domain := range chain[0].DNSNames {
			if strings.HasSuffix(domain, ".azmk8s.io") {
				var err error
				addr, err = tryAKS(domain)
				if err == nil {
					break DONE
				}
			}
		}
	}
	fmt.Println(addr)

	c2 := rest.CopyConfig(cfg)
	c2.Host = addr

	k2 := kubernetes.NewForConfigOrDie(c2)
	nodes, err := k2.CoreV1().Nodes().List(metav1.ListOptions{})
	fmt.Println(err)
	oneliners.PrettyJson(nodes.Items)

	time.Sleep(2 * time.Minute)
}

// ref: https://cloud.google.com/compute/docs/storing-retrieving-metadata
func tryAKS(domain string) (string, error) {
	data, err := ioutil.ReadFile("/sys/class/dmi/id/sys_vendor")
	if err != nil {
		return "", err
	}
	sysVendor := strings.TrimSpace(string(data))
	fmt.Println("sys_vendor = ", sysVendor)

	data, err = ioutil.ReadFile("/sys/class/dmi/id/product_name")
	if err != nil {
		return "", err
	}
	productName := strings.TrimSpace(string(data))
	fmt.Println("product_name = ", productName)

	if sysVendor != "Microsoft Corporation" && productName != "Virtual Machine" {
		return "", errors.New("not AKS")
	}
	return domain, nil
}
