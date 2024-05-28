#!/bin/bash

# -*-  Coding: UTF-8  -*- #
# -*-  System: Linux  -*- #
# -*-  Usage:   *.*   -*- #

# See Bash Set-Options Reference Below

set -euo pipefail # (0)
set -o xtrace     # (6)

# --------------------------------------------------------------------------------
# Bash Set-Options Reference
#     - https://tldp.org/LDP/abs/html/options.html
# --------------------------------------------------------------------------------

# 0. An Opinionated, Well Agreed Upon Standard for Bash Script Execution
# 1. set -o verbose     ::: Print Shell Input upon Read
# 2. set -o allexport   ::: Export all Variable(s) + Function(s) to Environment
# 3. set -o errexit     ::: Exit Immediately upon Pipeline'd Failure
# 4. set -o monitor     ::: Output Process-Separated Command(s)
# 5. set -o privileged  ::: Ignore Externals - Ensures of Pristine Run Environment
# 6. set -o xtrace      ::: Print a Trace of Simple Commands
# 7. set -o braceexpand ::: Enable Brace Expansion
# 8. set -o no-exec     ::: Bash Syntax Debugging

# --> script is for demonstrating how to debug kyverno -- main() has no purpose other than keeping records of
# ... these useful command(s)

function main() {
    kubectl get --raw /api/v1/namespaces | jq

    # --> true || false
    kubectl get --raw /api/v1/namespaces | kyverno jp query "items[*].metadata.name | contains(@, 'flux-system')"

    # --> yes || no
    kubectl auth can-i create ExternalSecret --as system:serviceaccount:kyverno:kyverno-background-controller

    kubectl get clusterrole kyverno:background-controller -o yaml
}

main
