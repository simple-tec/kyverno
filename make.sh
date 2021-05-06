#!/bin/bash
make docker-build-initContainer
make docker-tag-repo-initContainer
make docker-push-initContainer
make docker-build-kyverno
make docker-tag-repo-kyverno
make docker-push-kyverno
