package etcdleader

import (
	"context"

	"distributed-learning-lab/util/log"

	"go.etcd.io/etcd/client/v3/concurrency"
)

func Campaign(s *concurrency.Session, nodeId string) error {
	log.Debugf("node: %s start election", nodeId)

	e1 := concurrency.NewElection(s, "/my-test-election")
	if err := e1.Campaign(context.Background(), nodeId); err != nil {
		log.Errorf("campaign failed, err: %v", err)
		return err
	}

	log.Debugf("%s wins the election", nodeId)
	return nil
}
