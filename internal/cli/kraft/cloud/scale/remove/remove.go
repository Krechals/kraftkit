// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2024, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.

package remove

import (
	"context"
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	kraftcloud "sdk.kraft.cloud"
	kraftcloudautoscale "sdk.kraft.cloud/services/autoscale"

	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/config"
	"kraftkit.sh/internal/cli/kraft/cloud/utils"
	"kraftkit.sh/log"
)

type RemoveOptions struct {
	Auth   *config.AuthConfig                   `noattribute:"true"`
	Client kraftcloudautoscale.AutoscaleService `noattribute:"true"`
	Metro  string                               `noattribute:"true"`
}

func NewCmd() *cobra.Command {
	cmd, err := cmdfactory.New(&RemoveOptions{}, cobra.Command{
		Short:   "Delete an autoscale configuration policy",
		Use:     "remove [FLAGS] UUID NAME",
		Aliases: []string{"delete", "del", "rm"},
		Long:    "Delete an autoscale configuration policy.",
		Example: heredoc.Doc(`
			# Delete an autoscale configuration policy by UUID
			$ kraft cloud scale remove fd1684ea-7970-4994-92d6-61dcc7905f2b my-policy

			# Delete an autoscale configuration policy by name
			$ kraft cloud scale remove my-instance-431342 my-policy
		`),
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "kraftcloud-scale",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (opts *RemoveOptions) Pre(cmd *cobra.Command, args []string) error {
	if len(args) == 0 || len(args) == 1 {
		return fmt.Errorf("specify service group UUID and policy name")
	}

	opts.Metro = cmd.Flag("metro").Value.String()
	if opts.Metro == "" {
		opts.Metro = os.Getenv("KRAFTCLOUD_METRO")
	}
	if opts.Metro == "" {
		return fmt.Errorf("kraftcloud metro is unset")
	}

	log.G(cmd.Context()).WithField("metro", opts.Metro).Debug("using")

	return nil
}

func (opts *RemoveOptions) Run(ctx context.Context, args []string) error {
	var err error

	if !utils.IsUUID(args[0]) {
		return fmt.Errorf("specify a valid service group UUID")
	}

	if opts.Auth == nil {
		opts.Auth, err = config.GetKraftCloudAuthConfigFromContext(ctx)
		if err != nil {
			return fmt.Errorf("could not retrieve credentials: %w", err)
		}
	}

	if opts.Client == nil {
		opts.Client = kraftcloud.NewAutoscaleClient(
			kraftcloud.WithToken(config.GetKraftCloudTokenAuthConfig(*opts.Auth)),
		)
	}

	_, err = opts.Client.WithMetro(opts.Metro).DeletePolicyByName(ctx, args[0], args[1])
	if err != nil {
		return fmt.Errorf("could not delete policy: %w", err)
	}

	return nil
}