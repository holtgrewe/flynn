package cli

import (
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
usage: flynn-host demote ADDR

Demotes a Flynn node, removing it from the consensus cluster.
`)
}

// TODO(jpg): ADDR should default to the current node for ease of use.

func runPromote(args *docopt.Args, client *cluster.Client) error {
	addr := args.String["ADDR"]
	return discoverd.DefaultClient.RaftAddPeer(addr)
}

func runDemote(args *docopt.Args, client *cluster.Client) error {
	addr := args.String["ADDR"]
	return discoverd.DefaultClient.RaftRemovePeer(addr)
}
