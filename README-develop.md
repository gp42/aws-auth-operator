# README-release
This document defines the release process for the operator

## Development

1. Build manifests and operator: `make build manifests`
1. Build all images for olm
```
N=0 make all-olm
```

### Run
Use any of these options:

1. Locally (outside cluster): `make install run`
1. Deployment in the cluster: `make deploy`
1. OLM
You can run olm withou catalog with:
`operator-sdk run bundle docker.io/gp42/aws-auth-operator-bundle:v0.0.6-alpha.1`

Otherwise runn full cycle:
`kubectl apply -f deploy`

## Release a Candidate

Use a feature branch for candidates

1. Bump Operator version in VERSION file
2. Build
```
export DEFAULT_CHANNEL=candidate
export CHANNELS=$DEFAULT_CHANNEL
export N=0 # candidate number

# Build and push
make all-olm
```

## Release Stable

Use a feature branch

1. Bump Operator version in VERSION file
2. Build
```
export DEFAULT_CHANNEL=stable
export CHANNELS=$DEFAULT_CHANNEL

# Build and push
make all-olm
3. Add updated manifests to a feature branch and merge them to 'main'
4. Create a Release from autoprovisioned Release draft
```

## Test mkdocs
Serve mkdocs locally:
```bash
docker run --rm -it -p 8000:8000 -v ${PWD}:/docs squidfunk/mkdocs-material
```
