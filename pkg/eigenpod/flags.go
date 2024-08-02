package eigenpod

import "github.com/urfave/cli/v2"

var (
	PodAddressFlag = cli.StringFlag{
		Name:     "pod-address",
		Aliases:  []string{"pa"},
		Usage:    "Specify the address of the EigenPod",
		Required: true,
		EnvVars:  []string{"POD_ADDRESS"},
	}

	BatchSizeFlag = cli.Uint64Flag{
		Name:    "batch-size",
		Aliases: []string{"bsz"},
		Usage:   "Submit proofs in groups of size `--batch-size <batchSize>`, to avoid gas limit.",
		EnvVars: []string{"BATCH_SIZE"},
	}

	ProofSubmitterAddress = cli.StringFlag{
		Name:     "proof-submitter-address",
		Aliases:  []string{"psa"},
		Usage:    "Specify the address of the proof submitter. Only needed if you are using Web3 Signer.",
		Required: false,
		EnvVars:  []string{"PROOF_SUBMITTER_ADDRESS"},
	}
)
