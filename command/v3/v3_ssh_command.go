package v3

import (
	"io"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	sharedV2 "code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

type V3SSHActor interface {
	ExecuteSecureShellByAppNameAndSpace(
		appName,
		spaceGUID,
		processType string,
		processIndex int,
		sshInfo v2action.SSHInfo,
		stdin io.ReadCloser,
		stdout,
		stderr io.Writer,
	) error
}

type V2SSHActor interface {
	GetSSHInfo() (v2action.SSHInfo, error)
}

type V3SSHCommand struct {
	RequiredArgs        flag.AppName `positional-args:"yes"`
	ProcessType         string       `short:"p" description:"The process name" required:"true"`
	ProcessIndex        int          `short:"i" description:"The process index" required:"true"`
	Command             string       `short:"c" description:"command" required:"false"`
	DisablePseudoTTY    bool         `short:"T" description:"disable pseudo-tty" required:"false"`
	ForcePseudoTTY      bool         `short:"F" description:"force pseudo-tty" required:"false"`
	RequestPseudoTTY    bool         `short:"t" description:"request pseudo-tty" required:"false"`
	Forward             []string     `short:"L" description:"forward" required:"false"`
	SkipHostValidation  bool         `short:"k" description:"skip host validation" required:"false"`
	SkipRemoteExecution bool         `short:"N" description:"skip remote execution" required:"false"`

	usage interface{} `usage:"CF_NAME v3-ssh APP_NAME"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V3SSHActor
	V2SSHActor  V2SSHActor
}

func (cmd *V3SSHCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(ccClient, config)

	ccClientV2, uaaClientV2, err := sharedV2.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.V2SSHActor = v2action.NewActor(ccClientV2, uaaClientV2, config)

	return nil
}

func (cmd V3SSHCommand) Execute(args []string) error {
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

	sshInfo, err := cmd.V2SSHActor.GetSSHInfo()
	if err != nil {
		return shared.HandleError(err)
	}

	return cmd.Actor.ExecuteSecureShellByAppNameAndSpace(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
		cmd.ProcessType,
		cmd.ProcessIndex,
		sshInfo,
		cmd.UI.GetIn(),
		cmd.UI.GetOut(),
		cmd.UI.GetErr(),
	)
}
