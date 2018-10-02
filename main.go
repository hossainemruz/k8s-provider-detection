package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"k8s.io/client-go/rest"
	"log"
	"net/http"
	"time"

	"github.com/the-redback/go-oneliners"
	_ "k8s.io/client-go/rest"
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

func main() {
	flag.Parse()
	log.Println("kubeconfig: ",kubeConfig,"masterURL: ",masterURL)
	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		log.Fatalf("can't build config. Reason: %v", err.Error())
	}

	err=rest.LoadTLSFiles(cfg)
	if err!=nil{
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
	ok:=caCertPool.AppendCertsFromPEM(cfg.CAData)

	if !ok{
		log.Fatalf("Can't append caCert to caCertPool.")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs:caCertPool},
	}
	client := &http.Client{Transport: tr}

	resp,err:=client.Get("https://kubernetes.default.svc")

	if err!=nil{
		log.Println(err)
		return
	}

	for _,cert:=range resp.TLS.PeerCertificates{
		fmt.Println(cert.DNSNames)
	}

	oneliners.PrettyJson(resp,"Response")

	time.Sleep(2*time.Minute)
}
