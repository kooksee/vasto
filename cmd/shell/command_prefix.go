package shell

import (
	"bytes"
	"github.com/chrislusf/vasto/cmd/client"
	"strconv"
)

func init() {
	commands = append(commands, &CommandPrefix{})
}

type CommandPrefix struct {
	client *client.VastoClient
}

func (c *CommandPrefix) Name() string {
	return "prefix"
}

func (c *CommandPrefix) Help() string {
	return "prefix [limit lastSeenKey], prefix should also be the partition key"
}

func (c *CommandPrefix) SetCilent(client *client.VastoClient) {
	c.client = client
}

func (c *CommandPrefix) Do(args []string, env map[string]string) (string, error) {
	options, err := parseEnv(env)
	if err != nil {
		return "", err
	}

	prefix := []byte(args[0])
	limit := uint32(100)
	var lastSeenKey []byte
	if len(args) >= 2 {
		t, err := strconv.ParseUint(args[1], 10, 32)
		if err != nil {
			return "", err
		}
		limit = uint32(t)
	}
	if len(args) >= 3 {
		lastSeenKey = []byte(args[2])
	}

	keyValues, err := c.client.GetByPrefix(nil, prefix, limit, lastSeenKey, options...)

	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	for _, keyValue := range keyValues {
		output.Write(keyValue.Key)
		output.WriteString(" : ")
		output.Write(keyValue.Value)
		output.WriteString("\n")
	}
	return output.String(), nil
}