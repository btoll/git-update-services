# git-update-services

This is to be used either as a [Git extension] or as a standalone tool.  It's purpose is to generate a `kustomization.yaml` file that lists all of the particular services or applications that reside in a [`GitOps` directory structure] that should be synced and deployed by a tool like [Flux] or [Argo CD].

The reason this exists is because [`kustomize`] doesn't currently have [support for globbing], which makes maintaining a `kustomization.yaml` that references other "kustomized" directories potentially painstaking and error prone.

This is especially helpful in a multi-cluster environment where all services for a particular environment stage like `production`, `beta`, `development`, etc. are located in a location with `kustomize` overlays for each service.

## Installation

Compile the binary for your system architecture:

```bash
$ go build
```

Then, move this anywhere in your `PATH`.

If you want to build and install this in a file in your `GOBIN`, simply:

```bash
$ go install
```

This is probably the easiest method, as that directory will already be in your `PATH`.

To use as an extension to Git, move the `git-kustomize` file anywhere in your path after you've compiled the binary.

## Examples

```bash
$ git kustomize -p devops --env beta
```

## License

[GPLv3](COPYING)

## Author

[Benjamin Toll](https://benjamintoll.com)

[Git extension]: https://benjamintoll.com/2019/07/05/on-extending-git/
[`GitOps` directory structure]: https://github.com/btoll/gitops
[Flux]: https://fluxcd.io/
[Argo CD]: https://argoproj.github.io/cd/
[`kustomize`]: https://kustomize.io/
[support for globbing]: https://github.com/kubernetes-sigs/kustomize/issues/3205

