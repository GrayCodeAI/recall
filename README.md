# Recall

Sourced project memory for GrayCode agents. Recall is a small Go library and CLI. Every memory has a project scope and at least one source reference, so an agent can show where a decision came from.

## Try it

```sh
go build -o recall ./cmd/recall
./recall remember --scope my-project --kind decision --source checkpoint:abc123 --text 'Use device login for Reflect Cloud'
./recall search --scope my-project 'device login'
./recall list --scope my-project
```

Use `--store <directory>` on any command to choose the data directory. The default is the user data directory under `recall`. Records are individual JSON files written with mode `0600`; the directory uses mode `0700`. `forget <id>` deletes one record. No model calls or network sync occur.

## Boundary

Reflect can turn verified decisions into memories and cite their source IDs. Record can supply checkpoint IDs as references through Reflect. Relay supplies models when Reflect chooses to summarize; Ridge can read Recall's JSON output. Recall has no compile-time dependency on those products.

The first release uses local keyword search. Embeddings, automatic extraction, shared sync, and access control are future work.

## License

MIT — see [LICENSE](LICENSE) for details. Contributions are welcome under the terms in [CONTRIBUTING.md](CONTRIBUTING.md); report security issues privately as described in [SECURITY.md](SECURITY.md).

© 2026 GrayCode AI
