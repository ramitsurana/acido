package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/sgotti/acido/util"

	"github.com/sgotti/acido/Godeps/_workspace/src/github.com/appc/spec/schema"
	"github.com/sgotti/acido/Godeps/_workspace/src/github.com/appc/spec/schema/types"
	"github.com/sgotti/acido/Godeps/_workspace/src/github.com/coreos/fleet/log"
	"github.com/sgotti/acido/Godeps/_workspace/src/github.com/coreos/rocket/cas"
	"github.com/sgotti/acido/Godeps/_workspace/src/github.com/coreos/rocket/pkg/aci"
)

var (
	cmdStartBuild = &Command{
		Name:        "startbuild",
		Summary:     "Prepare an image for future build. If base image is specified it's extracted (satisfying all its dependencies) and a base app-manifest with dependencies set to the baseimage is created",
		Usage:       "BASEIMAGEHASH...",
		Description: `BASEIMAGEHASH hash of base image (it must exists in the store or it sould be imported with the \"import\" command.`,
		Run:         runStartBuild,
	}
)

func init() {
	commands = append(commands, cmdStartBuild)
}

func startBuild(args []string) error {
	ds, err := cas.NewStore(globalFlags.Dir)
	if err != nil {
		return err
	}

	baseImageIDStr := args[0]

	tmpdir, err := ioutil.TempDir(globalFlags.WorkDir, "")
	if err != nil {
		return err
	}
	log.Debugf("tmpdir: %s", tmpdir)

	key, err := util.KeyFromArg(baseImageIDStr, ds)
	if err != nil {
		return err
	}
	log.Debugf("key: %s", key)

	baseImageID, err := types.NewHash(key)
	if err != nil {
		return err
	}

	baseim, err := ds.GetImageManifest(key)
	if err != nil {
		return err
	}
	log.Debugf("baseim: %s", baseim)

	err = aci.RenderACIWithImageID(*baseImageID, tmpdir, ds)
	if err != nil {
		return err
	}
	log.Infof("Image extracted to %s", tmpdir)

	version := schema.AppContainerVersion
	im := schema.ImageManifest{
		ACKind:    "ImageManifest",
		ACVersion: version,
		Name:      baseim.Name,
		Labels:    baseim.Labels,
		Dependencies: types.Dependencies{
			types.Dependency{
				App:     baseim.Name,
				ImageID: baseImageID,
				Labels:  baseim.Labels,
			},
		},
	}

	out, err := im.MarshalJSON()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(tmpdir, "manifest"), out, 0644)
	if err != nil {
		return err
	}

	return nil
}

func runStartBuild(args []string) (exit int) {
	err := startBuild(args)
	if err != nil {
		log.Errorf("error: %v", err)
		return 1
	}
	return 0
}
