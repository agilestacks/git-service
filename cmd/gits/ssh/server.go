package ssh

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/agilestacks/git-service/cmd/gits/config"
	"github.com/agilestacks/git-service/cmd/gits/extapi"
	"github.com/agilestacks/git-service/cmd/gits/repo"
	"github.com/agilestacks/git-service/cmd/gits/util"
)

/* https://github.com/go-gitea/gitea/blob/HEAD/modules/ssh/ssh.go */

const (
	usersExtensionKey = "users"
)

func Listen(host string, port int) {
	keyBytes, err := ioutil.ReadFile(config.HostKeyFile)
	if err != nil {
		log.Fatalf("Failed to load SSH host private key from `%s`: %v", config.HostKeyFile, err)
	}
	key, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		log.Fatalf("Failed to parse SSH host private key from `%s`: %v", config.HostKeyFile, err)
	}
	server := &ssh.ServerConfig{MaxAuthTries: 20, PublicKeyCallback: checkKey}
	server.AddHostKey(key)
	go listen(server, host, port)
}

func listen(server *ssh.ServerConfig, host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting SSH connection: %v", err)
			continue
		}
		go accept(conn, server)
	}
}

func accept(conn net.Conn, server *ssh.ServerConfig) {
	if config.Verbose {
		log.Printf("SSH accepted connection from %v", conn.RemoteAddr())
	}
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, server)
	if err != nil {
		log.Printf("SSH handshake terminated: %v", err)
		return
	}
	go ssh.DiscardRequests(reqs)
	users := sshConn.Permissions.Extensions[usersExtensionKey]
	go handle(users, chans)
}

func checkKey(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	key64 := keyBase64(key)
	keyPrint := keyFingerprint(key)
	users, err := extapi.UsersBySshKey(key64, keyPrint)
	if err != nil {
		log.Printf("Unable to search for users by SSH key with fingerprint `%s`: %v", keyPrint, err)
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("No users matching SSH key `%s`", keyPrint)
	}
	return &ssh.Permissions{Extensions: map[string]string{usersExtensionKey: strings.Join(users, ",")}}, nil
}

func keyBase64(key ssh.PublicKey) string {
	return base64.StdEncoding.EncodeToString(key.Marshal())
}

func keyFingerprint(key ssh.PublicKey) string {
	return ssh.FingerprintSHA256(key)
}

func handle(_users string, newChannels <-chan ssh.NewChannel) {
	users := strings.Split(_users, ",")

	maintenance, maintMessage := util.Maintenance()

	for newChannel := range newChannels {
		if maintenance {
			newChannel.Reject(ssh.Prohibited, maintMessage)
			continue
		}
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "Only `session` channel is supported")
			continue
		}
		sshChannel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Error accepting SSH channel creation request: %v", err)
			continue
		}
		go handleRequests(users, sshChannel, requests)
	}
}

func gitCommand(cmd string) string {
	i := strings.Index(cmd, "git-")
	if i > 0 {
		return cmd[i:]
	}
	return cmd
}

func handleRequests(users []string, sshChannel ssh.Channel, requests <-chan *ssh.Request) {
	defer sshChannel.Close()

	for request := range requests {
		payload := string(request.Payload)
		if config.Debug {
			log.Printf("SSH request type `%s`; reply? %v; payload %q", request.Type, request.WantReply, payload)
		}

		switch request.Type {

		default:
			if config.Verbose {
				log.Printf("SSH request type `%s` not supported", request.Type)
			}
			break

		case "env":
			parts := strings.Split(strings.Replace(payload, "\x00", "", -1), "\v")
			if len(parts) != 2 {
				log.Printf("Bad SSH env request: %q", payload)
				continue
			}
			envVar := strings.TrimLeft(parts[0], "\b")
			value := parts[1]
			if config.Debug {
				log.Printf("SSH client requested env setup %s=%q", envVar, value)
			}
			break

		case "exec":
			cmd, err := repo.GitServer(gitCommand(payload), sshChannel, sshChannel, sshChannel.Stderr(), users)
			if err != nil {
				log.Printf("Failed to start Git server: %v", err)
				request.Reply(false, nil)
			} else {
				request.Reply(true, nil)
				status := byte(0)
				err = cmd.Wait()
				if err != nil {
					log.Printf("Git server failed: %v", err)
					status = byte(1)
				} else {
					if config.Debug {
						log.Print("Git server exited successfuly")
					}
				}
				sshChannel.SendRequest("exit-status", false, []byte{status, 0, 0, 0})
			}
			return
		}
	}
}
