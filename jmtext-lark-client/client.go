package jmtextLarkClient

import (
	"errors"
	jmtextProto "gitlab.int.jumei.com/JMArch/go-rpc/jmtext-proto"
	larkProto "gitlab.int.jumei.com/JMArch/go-rpc/lark-proto"
	"net"
)

type Client struct {
	service   string
	class     string
	method    string
	user      string
	secretKey string
	network   string
	netaddr   string
}

type RespOption struct {
	ServerIP string `json:"server_ip"`
}

func NewClient() *Client {
	c := &Client{
		secretKey: "769af463a39f077a0340a189e9c1ec28",
		network:   "tcp4",
		netaddr:   "127.0.0.1:12311",
	}
	return c
}

func (c *Client) WithLarkNetwork(network string) *Client {
	c.network = network
	return c
}

func (c *Client) WithLarkAddr(addr string) *Client {
	c.netaddr = addr
	return c
}

func (c *Client) WithServiceName(service string) *Client {
	c.service = service
	return c
}

func (c *Client) WithUser(user string) *Client {
	c.user = user
	return c
}

func (c *Client) WithSecretKey(secretKey string) *Client {
	c.secretKey = secretKey
	return c
}

func (c *Client) WithClass(class string) *Client {
	c.class = class
	return c
}

func (c *Client) Valid() (*Client, error) {
	if c.network == "" {
		return nil, errors.New("must have lark network")
	}

	if c.netaddr == "" {
		return nil, errors.New("must have lark netaddr")
	}

	if c.service == "" {
		return nil, errors.New("must have srevice name")
	}

	if c.user == "" {
		return nil, errors.New("must have user")
	}

	if c.secretKey == "" {
		return nil, errors.New("must have secret key")
	}

	if c.class == "" {
		return nil, errors.New("must have class")
	}

	return c, nil
}

func (c *Client) MustValid() *Client {
	_, err := c.Valid()
	if err != nil {
		panic(err)
	}

	return c
}

func (c *Client) Call(method string, owlContext map[string]interface{}, params ...interface{}) ([]byte, *larkProto.RespOption, error) {
	conn, err := net.Dial(c.network, c.netaddr)
	if err != nil {
		return nil, nil, err
	}

	//TODO 尝试长连接
	defer conn.Close()

	//打包 jmtext 数据
	b, err := jmtextProto.MarshalReq(c.user, c.secretKey, c.class, method, params, owlContext)
	if err != nil {
		return nil, nil, err
	}

	//打包 lark 数据
	b, err = larkProto.PackBizReq(b, &larkProto.LarkReqOption{
		TargetService: c.service,
	})
	if err != nil {
		return nil, nil, err
	}

	//发送数据
	_, err = conn.Write(b)
	if err != nil {
		return nil, nil, err
	}

	//读取 lark 协议
	b, o, err := larkProto.UnPackBizResp(conn)
	if err != nil {
		return nil, nil, err
	}

	//解析 jmtext 协议
	b, err = jmtextProto.UnmarshalResp(b)
	if err != nil {
		return nil, nil, err
	}

	//检查 exception
	err = jmtextProto.CheckException(b)

	return b, o, err
}
