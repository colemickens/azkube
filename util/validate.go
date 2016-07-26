package util

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	k8s "k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	validationDelay    = time.Second * 15
	validationAttempts = 20
)

func ValidateKubernetes(flavorArgs FlavorArguments) error {
	remainingRetries := validationAttempts
	for {
		remainingRetries--
		if remainingRetries <= 0 {
			break
		}

		log.Infof("Validating Kubernetes cluster.")

		c, err := getClient(flavorArgs)
		if err != nil {
			log.Warnf("Failed to get client for validation: %s", err)
			continue
		}

		err = validateStatus(flavorArgs, c)
		if err != nil {
			log.Warnf("Failed to validate components: %s", err)
			time.Sleep(validationDelay)
			continue
		}

		err = validateNodeCount(flavorArgs, c)
		if err != nil {
			log.Warnf("Failed to validate node count: %s", err)
			time.Sleep(validationDelay)
			continue
		}

		return nil
	}

	return fmt.Errorf("Failed to validate cluster after %d tries.", validationAttempts)
}

func getClient(flavorArgs FlavorArguments) (*k8s.Client, error) {
	config := &restclient.Config{
		Host: "https://" + flavorArgs.MasterFQDN + ":6443",
		TLSClientConfig: restclient.TLSClientConfig{
			CAData:   []byte(flavorArgs.CAKeyPair.CertificatePem),
			CertData: []byte(flavorArgs.ClientKeyPair.CertificatePem),
			KeyData:  []byte(flavorArgs.ClientKeyPair.PrivateKeyPem),
		},
	}

	return k8s.New(config)
}

func validateStatus(flavorArgs FlavorArguments, c *k8s.Client) error {
	log.Debugf("validate: status check")

	statuses := c.ComponentStatuses()
	statusList, err := statuses.List(k8sapi.ListOptions{})
	if err != nil {
		return err
	}

	log.Debugf("validate: got status list")
	for _, status := range statusList.Items {
		for _, condition := range status.Conditions {
			log.Debugf("validate: status (%q) type=%q status=%q message=%q error=%q",
				status.Name, condition.Type, condition.Status, condition.Message, condition.Error)

			if condition.Type == k8sapi.ComponentHealthy &&
				condition.Status != k8sapi.ConditionTrue {
				return fmt.Errorf("validate: component not healthy. component=%q status=%q message=%q error=%q", status.Name, condition.Status, condition.Message, condition.Error)
			}
		}
	}

	return nil
}

func validateNodeCount(flavorArgs FlavorArguments, c *k8s.Client) error {
	log.Debugf("validate: counting nodes")

	healthyNodeCount := 0
	expectedHealthyNodeCount := flavorArgs.NodeCount

	nodes := c.Nodes()
	nodeList, err := nodes.List(k8sapi.ListOptions{})
	if err != nil {
		return err
	}

	for _, node := range nodeList.Items {
		for _, condition := range node.Status.Conditions {
			log.Debugf("validate: node (%q) type=%q status=%q message=%q reason=%q", node.Name, condition.Type, condition.Status, condition.Message, condition.Reason)
			if condition.Type == k8sapi.NodeReady {
				if condition.Status == k8sapi.ConditionTrue {
					healthyNodeCount++
					continue
				} else {
					return fmt.Errorf("node not ready. node=%q status=%q message=%q reason=%q", node.Name, condition.Status, condition.Message, condition.Reason)
				}
			}
		}
	}

	if healthyNodeCount < expectedHealthyNodeCount {
		return fmt.Errorf("validate: incorrect healthy count. expected=%d actual=%d", flavorArgs.NodeCount, healthyNodeCount)
	} else if healthyNodeCount > expectedHealthyNodeCount {
		log.Warnf("validate: incorrect healthy count. expected=%d actual=%d", flavorArgs.NodeCount, healthyNodeCount)
	}

	return nil
}
