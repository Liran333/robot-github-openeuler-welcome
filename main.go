package main

import (
	"flag"
	"os"

	"github.com/opensourceways/community-robot-lib/giteeclient"
	"github.com/opensourceways/community-robot-lib/githubclient"
	"github.com/opensourceways/community-robot-lib/logrusutil"
	liboptions "github.com/opensourceways/community-robot-lib/options"
	framework "github.com/opensourceways/community-robot-lib/robot-github-framework"
	"github.com/opensourceways/community-robot-lib/secret"
	"github.com/sirupsen/logrus"
)

type options struct {
	service liboptions.ServiceOptions
	github  liboptions.GithubOptions
	gitee   liboptions.GiteeOptions
}

func (o *options) Validate() error {
	if err := o.service.Validate(); err != nil {
		return err
	}

	if err := o.gitee.Validate(); err != nil {
		return err
	}

	return o.github.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.github.AddFlags(fs)
	o.service.AddFlags(fs)
	o.gitee.AddFlags(fs)

	_ = fs.Parse(args)
	return o
}

func main() {
	logrusutil.ComponentInit(botName)

	o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatal("Invalid options")
	}

	secretAgent := new(secret.Agent)
	if err := secretAgent.Start([]string{o.github.TokenPath, o.gitee.TokenPath}); err != nil {
		logrus.WithError(err).Fatal("Error starting secret agent.")
	}

	defer secretAgent.Stop()

	c := githubclient.NewClient(secretAgent.GetTokenGenerator(o.github.TokenPath))
	g := giteeclient.NewClient(secretAgent.GetTokenGenerator(o.gitee.TokenPath))

	r := newRobot(c, g)

	framework.Run(r, o.service)
}
