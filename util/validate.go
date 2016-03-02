package util

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	k8sapi "k8s.io/kubernetes/pkg/api"
	k8s "k8s.io/kubernetes/pkg/client/unversioned"
)

func ValidateDeployment(flavorArgs FlavorArguments) error {
	remainingRetries := 20
	for {
		remainingRetries--
		if remainingRetries <= 0 {
			break
		}

		c, err := getClient(flavorArgs)
		if err != nil {
			log.Warnf("validate: failed to get client: %s", err)
			time.Sleep(10 * time.Second)
			continue
		}

		err = validateStatus(flavorArgs, c)
		if err != nil {
			log.Warnf("validate: failed to validate components: %s", err)
			time.Sleep(10 * time.Second)
			continue
		}

		err = validateNodeCount(flavorArgs, c)
		if err != nil {
			log.Warnf("validate: failed to validate node count: %s", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// we're good if we got here

		return nil
	}

	log.Errorf("validate: cluster validation failed.")
	return fmt.Errorf("validate: cluster validation failed.")
}

func getClient(flavorArgs FlavorArguments) (*k8s.Client, error) {
	config := &k8s.Config{
		Host: "https://" + flavorArgs.MasterFQDN + ":6443",
		TLSClientConfig: k8s.TLSClientConfig{
			CAData:   []byte(flavorArgs.CAKeyPair.CertificatePem),
			CertData: []byte(flavorArgs.ClientKeyPair.CertificatePem),
			KeyData:  []byte(flavorArgs.ClientKeyPair.PrivateKeyPem),
		},
	}

	return k8s.New(config)
}

func validateStatus(flavorArgs FlavorArguments, c *k8s.Client) error {
	log.Infof("validate: status check")

	statuses := c.ComponentStatuses()
	statusList, err := statuses.List(nil, nil)
	if err != nil {
		return err
	}

	log.Infof("validate: got status list")
	for _, status := range statusList.Items {
		for _, condition := range status.Conditions {
			log.Infof("validate: status (%q) type=%q status=%q message=%q error=%q", status.Name, condition.Type, condition.Status, condition.Message, condition.Error)
			if condition.Type == k8sapi.ComponentHealthy && condition.Status != k8sapi.ConditionTrue {
				return fmt.Errorf("validate: component not healthy. component=%q status=%q message=%q error=%q", status.Name, condition.Status, condition.Message, condition.Error)
			}
		}
	}

	return nil
}

func validateNodeCount(flavorArgs FlavorArguments, c *k8s.Client) error {
	log.Infof("validate: counting nodes")

	healthyNodeCount := 0
	nodes := c.Nodes()
	nodeList, err := nodes.List(nil, nil)
	if err != nil {
		return err
	}

	for _, node := range nodeList.Items {
		for _, condition := range node.Status.Conditions {
			log.Infof("validate: node (%q) type=%q status=%q message=%q reason=%q", node.Name, condition.Type, condition.Status, condition.Message, condition.Reason)
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

	if healthyNodeCount != flavorArgs.NodeCount {
		return fmt.Errorf("validate: incorrect healthy count. expected=%d actual=%d", flavorArgs.NodeCount, healthyNodeCount)
	}

	return nil
}
