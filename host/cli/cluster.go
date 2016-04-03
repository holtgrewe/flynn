package cli

import (
	"log"

	"github.com/flynn/flynn/Godeps/_workspace/src/github.com/flynn/go-docopt"
	"github.com/flynn/flynn/discoverd/client"
	"github.com/flynn/flynn/pkg/cluster"
)

func init() {
	Register("promote", runPromote, `
usage: flynn-host promote ADDR

Promotes a Flynn node to a member of the consensus cluster.
`)
	Register("demote", runDemote, `
usage: flynn-host demote [-f|--force] ADDR

Demotes a Flynn node, removing it from the consensus cluster.
`)
}

// TODO(jpg): ADDR should default to the current node for ease of use.

func runPromote(args *docopt.Args, client *cluster.Client) error {
	addr := args.String["ADDR"]
	dd := discoverd.NewClientWithURL(addr)
	if err := dd.Promote(); err != nil {
		return err
	}
	log.Println("Promoted peer", addr)
	return nil
}

func runDemote(args *docopt.Args, client *cluster.Client) error {
	addr := args.String["ADDR"]
	force := false
	// first try to connect to the peer and gracefully demote it
	dd := discoverd.NewClientWithURL(addr)
	err := dd.Demote()
	// if that fails and --force is given forcefully remove it
	// by instructing the raft leader to remove it from the raft peers directly
	if err != nil && force {
		leader, err := discoverd.DefaultClient.RaftLeader()
		if err != nil {
			return err
		}
		dd = discoverd.NewClientWithURL(leader.Host)
		if err := dd.RaftRemovePeer(addr); err != nil {
			return err
		}
		log.Println("Forcefully removed peer", addr)
		return nil
	} else if err != nil {
		return err
	}
	log.Println("Demoted peer", addr)
	return nil
}
