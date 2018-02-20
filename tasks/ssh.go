package tasks

import (
	"context"
	"fmt"
	"github.com/ixday/antfarm"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io"
	"net"
	"os"
	"os/user"
)

type SSHConfig struct {
	User string
	Host string
	Port int
	Auth []ssh.AuthMethod
	PTY  SSHConfigPTY
}

type SSHConfigPTY struct {
	Modes         ssh.TerminalModes
	Term          string
	Height, Width int
}

type SSHConfigRun struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type SSH struct {
	client *ssh.Client
	config *SSHConfig
}

func (ssh *SSH) Close() antfarm.Task {
	return antfarm.TaskFunc(func(_ context.Context) error {
		ssh.client.Close()
		return nil
	})
}

func NewSSH(options ...func(*SSHConfig)) *SSH {
	sshConfig := &SSHConfig{Host: "localhost", Port: 22, PTY: SSHConfigPTY{
		ssh.TerminalModes{
			ssh.ECHO:          0,     // disable echoing
			ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
			ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		},
		"xterm", 40, 80,
	}}
	authMethods := []ssh.AuthMethod{}

	login, err := user.Current()
	if err != nil {
		login = &user.User{Username: "root"}
	}
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err == nil {
		authMethods = append(authMethods, ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
	}

	sshConfig.User = login.Username
	sshConfig.Auth = authMethods

	for _, option := range options {
		option(sshConfig)
	}

	return &SSH{nil, sshConfig}
}

func (_ssh *SSH) Connect() antfarm.Task {
	return antfarm.TaskFunc(func(ctx context.Context) (err error) {
		_ssh.client, err = ssh.Dial(
			"tcp",
			fmt.Sprintf("%s:%d", _ssh.config.Host, _ssh.config.Port),
			&ssh.ClientConfig{
				User:            _ssh.config.User,
				Auth:            _ssh.config.Auth,
				HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
			})
		return
	})
}

func (ssh *SSH) Run(cmd string, options ...func(*SSHConfigRun)) antfarm.Task {
	sshConfigRun := &SSHConfigRun{os.Stdin, os.Stdout, os.Stderr}
	for _, option := range options {
		option(sshConfigRun)
	}

	return antfarm.TaskFunc(func(ctx context.Context) error {
		if ssh.client == nil {
			if err := ssh.Connect().Start(ctx); err != nil {
				return err
			}
		}
		session, err := ssh.client.NewSession()
		if err != nil {
			return fmt.Errorf("Failed to create session: %s", err)
		}

		defer session.Close()
		pty := ssh.config.PTY

		if err := session.RequestPty(pty.Term, pty.Height, pty.Width, pty.Modes); err != nil {
			return fmt.Errorf("request for pseudo terminal failed: %s", err)
		}

		stdin, err := session.StdinPipe()
		if err != nil {
			return fmt.Errorf("Unable to setup stdin for session: %v", err)
		}
		go io.Copy(stdin, sshConfigRun.Stdin)

		stdout, err := session.StdoutPipe()
		if err != nil {
			return fmt.Errorf("Unable to setup stdout for session: %v", err)
		}
		go io.Copy(sshConfigRun.Stdout, stdout)

		stderr, err := session.StderrPipe()
		if err != nil {
			return fmt.Errorf("Unable to setup stderr for session: %v", err)
		}
		go io.Copy(sshConfigRun.Stderr, stderr)

		return session.Run(cmd)

	})
}
