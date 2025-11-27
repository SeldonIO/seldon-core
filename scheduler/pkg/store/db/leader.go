package db

import (
	"context"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

func leader() {
	etcClient, err := clientv3.New(clientv3.Config{})
	if err != nil {
		panic(err)
	}

	defer etcClient.Close()

	// create a new session for leader election
	electionSession, err := concurrency.NewSession(etcClient, concurrency.WithTTL(30))
	if err != nil {
		panic(err)
	}

	defer electionSession.Close()

	election := concurrency.NewElection(electionSession, "election-prefix")
	ctx := context.Background()

	fmt.Println("attempting to become a leader")

	//start leader election
	if err := election.Campaign(ctx, "election-prefix"); err != nil {
		panic(err)
	}

	fmt.Println("become a leader")

}
