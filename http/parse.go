package http

import (
	"net/http"
	"strings"

	cmds "github.com/jbenet/go-ipfs/commands"
)

// Parse parses the data in a http.Request and returns a command Request object
func Parse(r *http.Request, root *cmds.Command) (cmds.Request, error) {
	path := strings.Split(r.URL.Path, "/")[3:]
	stringArgs := make([]string, 0)

	cmd, err := root.Get(path[:len(path)-1])
	if err != nil {
		// 404 if there is no command at that path
		return nil, ErrNotFound

	} else if sub := cmd.Subcommand(path[len(path)-1]); sub == nil {
		if len(path) <= 1 {
			return nil, ErrNotFound
		}

		// if the last string in the path isn't a subcommand, use it as an argument
		// e.g. /objects/Qabc12345 (we are passing "Qabc12345" to the "objects" command)
		stringArgs = append(stringArgs, path[len(path)-1])
		path = path[:len(path)-1]

	} else {
		cmd = sub
	}

	opts, stringArgs2 := parseOptions(r)
	stringArgs = append(stringArgs, stringArgs2...)

	// Note that the argument handling here is dumb, it does not do any error-checking.
	// (Arguments are further processed when the request is passed to the command to run)
	args := make([]interface{}, len(cmd.Arguments))
	for i, arg := range cmd.Arguments {
		if arg.Type == cmds.ArgString {
			if len(stringArgs) > 0 {
				args[i] = stringArgs[0]
				stringArgs = stringArgs[1:]
			}

		} else {
			// TODO: create multipart streams for file args
			args[i] = r.Body
		}
	}

	return cmds.NewRequest(path, opts, args, cmd), nil
}

func parseOptions(r *http.Request) (map[string]interface{}, []string) {
	opts := make(map[string]interface{})
	var args []string

	query := r.URL.Query()
	for k, v := range query {
		if k == "arg" {
			args = v
		} else {
			opts[k] = v[0]
		}
	}

	// default to setting encoding to JSON
	_, short := opts[cmds.EncShort]
	_, long := opts[cmds.EncLong]
	if !short && !long {
		opts[cmds.EncShort] = cmds.JSON
	}

	return opts, args
}