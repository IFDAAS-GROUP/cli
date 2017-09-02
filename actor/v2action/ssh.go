package v2action

import (
	"errors"
	"os"

	"code.cloudfoundry.org/cli/util/clissh"
	"code.cloudfoundry.org/cli/util/clissh/sshoptions"
	"code.cloudfoundry.org/cli/util/clissh/sshterminal"
	"golang.org/x/crypto/ssh"
)

func (actor Actor) GetSSHPasscode() (string, error) {
	return actor.UAAClient.GetSSHPasscode(actor.Config.AccessToken(), actor.Config.SSHOAuthClient())
}

func (actor Actor) RunSecureShell(sshOptions sshoptions.SSHOptions, ui UI) error {
	app, _, err := actor.GetApplicationByNameAndSpace(sshOptions.AppName, sshOptions.SpaceGUID)
	if err != nil {
		return err
	}

	passcode, err := actor.GetSSHPasscode()
	if err != nil {
		return err
	}

	secureShell := clissh.NewSecureShell(
		clissh.DefaultSecureDialer(),
		sshterminal.DefaultHelper(),
		clissh.DefaultListenerFactory(),
		clissh.DefaultKeepAliveInterval,
		app,
		actor.CloudControllerClient.AppSSHHostKeyFingerprint(),
		actor.CloudControllerClient.AppSSHEndpoint(),
		passcode,
	)

	sshOptions.TerminalRequest = sshoptions.RequestTTYAuto
	err = secureShell.Connect(&sshOptions)
	if err != nil {
		return errors.New("Error opening SSH connection: " + err.Error())
	}
	defer secureShell.Close()

	err = secureShell.InteractiveSession(ui.GetIn(), ui.GetOut(), ui.GetErr())
	if err != nil {
		if exitError, ok := err.(*ssh.ExitError); ok {
			os.Exit(exitError.ExitStatus())
		} else {
			return errors.New("Error: " + err.Error())
		}
	}

	return nil
}
