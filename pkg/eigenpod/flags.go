package eigenpod

import "github.com/urfave/cli/v2"

var (
	PodAddressFlag = cli.StringFlag{
		Name:     "pod-address",
		Usage:    "Specify the address of the EigenPod",
		Required: true,
		EnvVars:  []string{"POD_ADDRESS"},
	}

	BatchSizeFlag = cli.Uint64Flag{
		Name:    "batch-size",
		Usage:   "Specify the batch size for submitting proofs",
		Aliases: []string{"bsz"},
		EnvVars: []string{"BATCH_SIZE"},
	}

	ProofPathFlag = cli.StringFlag{
		Name:    "proof",
		Usage:   "the `path` to a previous proof generated from this step (via -o proof.json).",
		Aliases: []string{"p"},
		EnvVars: []string{"PROOF_PATH"},
	}

	ForceCheckpointFlag = cli.BoolFlag{
		Name:    "force",
		Usage:   "If true, starts a checkpoint even if the pod has no native ETH to award shares",
		Aliases: []string{"f"},
		EnvVars: []string{"FORCE_CHECKPOINT"},
	}

	CheckpointSubmitterAddressFlag = cli.StringFlag{
		Name:    "checkpoint-submitter",
		Usage:   "Address of checkpoint submitter. This is only needed for Web3 Signer",
		Aliases: []string{"cp"},
		EnvVars: []string{"CHECKPOINT_SUBMITTER"},
	}
)
