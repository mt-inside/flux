package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/weaveworks/flux"
	"github.com/weaveworks/flux/update"
)

type controllerReleaseOpts struct {
	*rootOpts
	namespace      string
	controllers    []string
	allControllers bool
	image          string
	allImages      bool
	exclude        []string
	dryRun         bool
	outputOpts
	cause update.Cause

	// Deprecated
	services []string
}

func newControllerRelease(parent *rootOpts) *controllerReleaseOpts {
	return &controllerReleaseOpts{rootOpts: parent}
}

func (opts *controllerReleaseOpts) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Release a new version of a controller.",
		Example: makeExample(
			"fluxctl release -n default --controller=deployment/foo --update-image=library/hello:v2",
			"fluxctl release --all --update-image=library/hello:v2",
			"fluxctl release --controller=default:deployment/foo --update-all-images",
		),
		RunE: opts.RunE,
	}

	AddOutputFlags(cmd, &opts.outputOpts)
	AddCauseFlags(cmd, &opts.cause)
	cmd.Flags().StringVarP(&opts.namespace, "namespace", "n", "default", "Controller namespace")
	cmd.Flags().StringSliceVarP(&opts.controllers, "controller", "c", []string{}, "List of controllers to release <namespace>:<kind>/<name>")
	cmd.Flags().BoolVar(&opts.allControllers, "all", false, "Release all controllers")
	cmd.Flags().StringVarP(&opts.image, "update-image", "i", "", "Update a specific image")
	cmd.Flags().BoolVar(&opts.allImages, "update-all-images", false, "Update all images to latest versions")
	cmd.Flags().StringSliceVar(&opts.exclude, "exclude", []string{}, "List of controllers to exclude")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "Do not release anything; just report back what would have been done")

	// Deprecated
	cmd.Flags().StringSliceVarP(&opts.services, "service", "s", []string{}, "Service to release")
	cmd.Flags().MarkHidden("service")

	return cmd
}

func (opts *controllerReleaseOpts) RunE(cmd *cobra.Command, args []string) error {
	if len(opts.services) > 0 {
		return errorServiceFlagDeprecated
	}

	if len(args) != 0 {
		return errorWantedNoArgs
	}

	if err := checkExactlyOne("--update-image=<image> or --update-all-images", opts.image != "", opts.allImages); err != nil {
		return err
	}

	if len(opts.controllers) <= 0 && !opts.allControllers {
		return newUsageError("please supply either --all, or at least one --controller=<controller>")
	}

	var controllers []update.ResourceSpec
	if opts.allControllers {
		controllers = []update.ResourceSpec{update.ResourceSpecAll}
	} else {
		for _, controller := range opts.controllers {
			id, err := flux.ParseResourceIDOptionalNamespace(opts.namespace, controller)
			if err != nil {
				return err
			}
			controllers = append(controllers, update.MakeResourceSpec(id))
		}
	}

	var (
		image update.ImageSpec
		err   error
	)
	switch {
	case opts.image != "":
		image, err = update.ParseImageSpec(opts.image)
		if err != nil {
			return err
		}
	case opts.allImages:
		image = update.ImageSpecLatest
	}

	var kind update.ReleaseKind = update.ReleaseKindExecute
	if opts.dryRun {
		kind = update.ReleaseKindPlan
	}

	var excludes []flux.ResourceID
	for _, exclude := range opts.exclude {
		s, err := flux.ParseResourceIDOptionalNamespace(opts.namespace, exclude)
		if err != nil {
			return err
		}
		excludes = append(excludes, s)
	}

	if opts.dryRun {
		fmt.Fprintf(cmd.OutOrStderr(), "Submitting dry-run release...\n")
	} else {
		fmt.Fprintf(cmd.OutOrStderr(), "Submitting release ...\n")
	}

	ctx := context.Background()
	spec := update.ReleaseSpec{
		ServiceSpecs: controllers,
		ImageSpec:    image,
		Kind:         kind,
		Excludes:     excludes,
	}
	jobID, err := opts.API.UpdateManifests(ctx, update.Spec{
		Type:  update.SpecImages,
		Cause: opts.cause,
		Spec:  spec,
	})
	if err != nil {
		return err
	}

	return await(ctx, cmd.OutOrStdout(), cmd.OutOrStderr(), opts.API, jobID, !opts.dryRun, opts.verbosity)
}
