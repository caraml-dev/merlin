# Docs

To learn about the basic concepts behind Merlin and how to use it, refer to the [User Docs](./user/generated).

To configure / deploy Merlin into a production cluster or troubleshoot an existing deployment, refer to the [Maintainer Docs](./maintainer).

To understand the development process and the architecture, refer to the [Developer Docs](./developer).

## Contributing to the Docs

All docs are created for Gitbook.

Currently, the user docs are templated using Jinja2, under [user/templates](./user/templates) and the values for the templates reside in [values.json](./user/values.json). To generate the final docs into [user/generated](./user/generated), run:

```sh
make docs
```
