package v2

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

type SSHActor interface {
	RunSecureShell(sshOptions v2action.SSHOptions, ui v2action.UI) error
}

type SSHCommand struct {
	RequiredArgs        flag.AppName `positional-args:"yes"`
	AppInstanceIndex    int          `long:"app-instance-index" short:"i" description:"Application instance index (Default: 0)"`
	Command             string       `long:"command" short:"c" description:"Command to run. This flag can be defined more than once."`
	DisablePseudoTTY    bool         `long:"disable-pseudo-tty" short:"T" description:"Disable pseudo-tty allocation"`
	ForcePseudoTTY      bool         `long:"force-pseudo-tty" description:"Force pseudo-tty allocation"`
	LocalPort           string       `short:"L" description:"Local port forward specification. This flag can be defined more than once."`
	RemotePseudoTTY     bool         `long:"request-pseudo-tty" short:"t" description:"Request pseudo-tty allocation"`
	SkipHostValidation  bool         `long:"skip-host-validation" short:"k" description:"Skip host key validation"`
	SkipRemoteExecution bool         `long:"skip-remote-execution" short:"N" description:"Do not execute a remote command"`
	usage               interface{}  `usage:"CF_NAME ssh APP_NAME [-i INDEX] [-c COMMAND]... [-L [BIND_ADDRESS:]PORT:HOST:HOST_PORT] [--skip-host-validation] [--skip-remote-execution] [--disable-pseudo-tty | --force-pseudo-tty | --request-pseudo-tty]"`
	relatedCommands     interface{}  `related_commands:"allow-space-ssh, enable-ssh, space-ssh-allowed, ssh-code, ssh-enabled"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SSHActor
}

func (cmd *SSHCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd SSHCommand) Execute([]string) error {
	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayTextWithFlavor("Sshing into app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	sshOptions := v2action.SSHOptions{
		AppName:   cmd.RequiredArgs.AppName,
		SpaceGUID: cmd.Config.TargetedOrganization().GUID,
		Index:     cmd.AppInstanceIndex,
	}
	return cmd.Actor.RunSecureShell(sshOptions, cmd.UI)
}
