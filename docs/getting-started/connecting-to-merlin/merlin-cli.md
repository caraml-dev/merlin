# Merlin CLI

The Merlin CLI can be installed directly using pip:

```bash
pip install merlin-sdk
```

The CLI is a wrapper of [Merlin Python SDK](./python-sdk.md).

```bash
$ merlin
Usage: merlin [OPTIONS] COMMAND [ARGS]...

  A simple command line tool.

  The Merlin CLI assumes that you already have a serialized model.

  To see the options for each command: merlin COMMAND --help

Options:
  --help  Show this message and exit.

Commands:
  deploy    Deploy the model
  scaffold  Generate PyFunc project
  undeploy  Undeploy the model
```
