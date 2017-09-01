package v3action

import (
	"errors"
	"io"
	"os"

	"golang.org/x/crypto/ssh"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/cf/sshcmd"
	"code.cloudfoundry.org/cli/cf/sshcmd/options"
	"code.cloudfoundry.org/cli/cf/sshcmd/terminal"
)

func (a Actor) ExecuteSecureShellByAppNameAndSpace(
	appName,
	spaceGUID,
	processType string,
	processIndex int,
	sshInfo v2action.SSHInfo,
	stdin io.ReadCloser,
	stdout,
	stderr io.Writer,
) error {
	summary, _, err := a.GetApplicationSummaryByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return err
	}

	var processGUID string
	for _, process := range summary.Processes {
		if process.Type == processType {
			processGUID = process.GUID
		}
	}

	if processGUID == "" {
		return errors.New("process does not exist")
	}

	secureShell := sshcmd.NewSecureShell(
		sshcmd.DefaultSecureDialer(),
		terminal.DefaultHelper(),
		sshcmd.DefaultListenerFactory(),
		sshcmd.DefaultKeepAliveInterval,
		summary.State,
		processGUID,
		sshInfo.SSHHostKeyFingerprint,
		sshInfo.SSHEndpoint,
		sshInfo.SSHPasscode,
	)

	opts := options.SSHOptions{
		AppName:         appName,
		Index:           uint(processIndex),
		TerminalRequest: options.RequestTTYAuto,
	}

	err = secureShell.Connect(&opts)
	if err != nil {
		return errors.New("Error opening SSH connection: " + err.Error())
	}
	defer secureShell.Close()

	err = secureShell.InteractiveSession(stdin, stdout, stderr)
	if err != nil {
		if exitError, ok := err.(*ssh.ExitError); ok {
			exitStatus := exitError.ExitStatus()
			os.Exit(exitStatus)
		} else {
			return errors.New("Error: " + err.Error())
		}
	}

	return nil
}
