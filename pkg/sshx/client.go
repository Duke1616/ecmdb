package sshx

import (
	"fmt"
	"golang.org/x/crypto/ssh"
)

func ConnectToNextJumpHost(currentClient *ssh.Client, user string, host string, port int, method ssh.AuthMethod) (*ssh.Client, error) {
	// 创建下一层跳板机的 SSH 客户端配置
	nextConfig := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{method},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	address := fmt.Sprintf("%s:%d", host, port)
	// 通过当前跳板机连接到下一层跳板机
	nextConn, err := currentClient.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to next jump host: %v", err)
	}

	// 创建下一层跳板机的 SSH 会话
	nextClient, nextChan, nextReqs, err := ssh.NewClientConn(nextConn, address, nextConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client connection to next jump host: %v", err)
	}

	// 创建下一层跳板机的 SSH 客户端
	return ssh.NewClient(nextClient, nextChan, nextReqs), nil
}

func (mgm *MultiGatewayManager) Connect() (*ssh.Client, error) {
	var client *ssh.Client
	var err error

	for i, gateway := range mgm.Gateways {
		// 获取认证方式
		var authMethod ssh.AuthMethod
		authMethod, err = Auth(gateway)
		if err != nil {
			return nil, err
		}

		// 配置 SSH 客户端
		config := &ssh.ClientConfig{
			User:            gateway.Username,
			Auth:            []ssh.AuthMethod{authMethod},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		if i == 0 {
			// 连接第一个网关
			client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", gateway.Host, gateway.Port), config)
			if err != nil {
				return nil, fmt.Errorf("连接第一个网关失败: %v", err)
			}
		} else {
			// 这里你应该确保 client 是继续在原连接上做的，而不是新建连接
			client, err = ConnectToNextJumpHost(client, gateway.Username, gateway.Host, gateway.Port, authMethod)
			if err != nil {
				return nil, fmt.Errorf("通过网关 %s 连接失败: %v", gateway.Host, err)
			}
		}
	}
	
	return client, nil
}

func Auth(gateway *GatewayConfig) (ssh.AuthMethod, error) {
	switch gateway.AuthType {
	case "passwd":
		return ssh.Password(gateway.Password), nil
	case "passphrase":
		privateKeyBytes := []byte(gateway.PrivateKey)
		passwordBytes := []byte(gateway.Password)
		signer, err := ssh.ParsePrivateKeyWithPassphrase(privateKeyBytes, passwordBytes)
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %v", err)
		}
		return ssh.PublicKeys(signer), nil
	case "publickey":
		privateKeyBytes := []byte(gateway.PrivateKey)
		signer, err := ssh.ParsePrivateKey(privateKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %v", err)
		}

		return ssh.PublicKeys(signer), nil
	}

	return nil, nil
}
