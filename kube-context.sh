#!/bin/bash

usage() {
  echo "Usage: $0 [options]"
  echo "Options:"
  echo "  -h, --help              Display this help message"
  echo "  -m, --merge <file>      Merge kubeconfig file"
  echo "  -a, --add               Add a new context (requires --token, --endpoint, --name)"
  echo "  -c, --clean             Clean up unusable contexts"
  echo "  -d, --delete            Delete context by name"
  echo "  -t, --token <value>     Authentication token"
  echo "  -e, --endpoint <value>  Cluster endpoint URL"
  echo "  -n, --name <value>      Context name"
  echo "Examples:"
  echo "  $0 --merge /path/to/kubeconfig.yaml"
  echo "  $0 --add --token <token> --endpoint <url> --name my-context"
  echo "  $0 --delete --name my-context"
  exit 0
}

merge-context() {
  local kubeconfig="$1"
  if [[ ! -f "$kubeconfig" ]]; then
    echo "Error: Kubeconfig file '$kubeconfig' not found"
    exit 2
  fi
  if command -v yq >/dev/null 2>&1; then
    echo >/dev/null
  else
    echo "command yq not found, need yq install"
    exit 1
  fi

  TOKEN=$(yq '.users[] | select(.name == "admin").user.token' "$kubeconfig")
  ENDPOINT=$(yq '.clusters[] | select(.name == "global").cluster.server' "$kubeconfig")
  NAME=$(basename "$kubeconfig" .yaml | sed 's/ //g')

  if [[ -z "$TOKEN" || -z "$ENDPOINT" || -z "$NAME" ]]; then
    echo "Error: Failed to extract token, endpoint, or name from '$kubeconfig'"
    exit 3
  fi

  create-context
}

create-context() {
  new-context
  if [ $? -eq 0 ]; then
    kubectl config use-context $NAME
    export count=$(kubectl api-resources | grep "platform.tkestack.io/v1" | wc -l)
    if [ $count -gt 0 ]; then
      export cls=$(kubectl get cls | grep -E -vi "name|global" | awk '{print $1}')
      export cls_count=$(echo $cls | wc -l)
      if [ $cls_count -gt 0 ]; then
        echo "add clusters in $NAME"
        for i in $cls; do
          export ONAME=$NAME
          export NAME=$NAME-"$i"
          export ENDPOINT="${ENDPOINT%/*}/$i"
          new-context
          export NAME=${ONAME}
        done
      fi
    fi
  fi
}

clean-context() {
  for ctx in $(kubectl config get-contexts -o name); do
    kubectl config use-context $ctx
    if ! timeout 5 kubectl cluster-info &>/dev/null; then
      echo "Context $ctx is invalid, deleting..."
      kube-context --delete --name $ctx
    fi
  done
}

new-context() {
  if [[ -z "$TOKEN" || -z "$ENDPOINT" || -z "$NAME" ]]; then
    echo "Error: Missing required parameters (--token, --endpoint, --name)"
    exit 4
  fi

  echo "Creating context:"
  echo "  Name: $NAME"
  echo "  Endpoint: $ENDPOINT"
  echo "  Token: $TOKEN"

  if kubectl config get-contexts -o name | grep -Fx "$NAME" >/dev/null; then
    echo "Error: Context '$NAME' already exists"
    exit 5
  fi

  kubectl --token "$TOKEN" --server "$ENDPOINT" --insecure-skip-tls-verify get ns &>/dev/null
  if [[ $? -ne 0 ]]; then
    echo "Error: Invalid token or endpoint"
    exit 6
  fi

  kubectl config set-cluster "$NAME" --server="$ENDPOINT" --insecure-skip-tls-verify
  kubectl config set-credentials "$NAME" --token="$TOKEN"
  kubectl config set-context "$NAME" --cluster="$NAME" --user="$NAME"
  echo "Context '$NAME' created successfully"
}

delete-context() {
  if [[ -z "$NAME" ]]; then
    echo "Error: Missing required parameter --name"
    exit 4
  fi

  echo "Deleting context: $NAME"
  kubectl config delete-cluster "$NAME"
  kubectl config delete-context "$NAME"
  kubectl config delete-user "$NAME"
  echo "Context '$NAME' deleted successfully"
}

ACTION=""
KUBE_CONFIG=""
TOKEN=""
ENDPOINT=""
NAME=""
DELETE="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
  -h | --help)
    usage
    ;;
  -m | --merge)
    if [[ -z "$2" ]]; then
      echo "Error: --merge requires a kubeconfig file path"
      exit 1
    fi
    KUBE_CONFIG="$2"
    ACTION="merge"
    shift 2
    ;;
  -a | --add)
    ACTION="add"
    shift
    ;;
  -c | --clean)
    ACTION="clean"
    shift
    ;;
  -d | --delete)
    DELETE="true"
    ACTION="delete"
    shift
    ;;
  -t | --token)
    if [[ -z "$2" ]]; then
      echo "Error: --token requires a value"
      exit 1
    fi
    TOKEN="$2"
    shift 2
    ;;
  -e | --endpoint)
    if [[ -z "$2" ]]; then
      echo "Error: --endpoint requires a value"
      exit 1
    fi
    ENDPOINT="$2"
    shift 2
    ;;
  -n | --name)
    if [[ -z "$2" ]]; then
      echo "Error: --name requires a value"
      exit 1
    fi
    NAME="$2"
    shift 2
    ;;
  *)
    echo "Error: Invalid option '$1'"
    echo "Use '$0 --help' for more information"
    exit 1
    ;;
  esac
done

# Execute the appropriate action
case "$ACTION" in
merge)
  merge-context "$KUBE_CONFIG"
  ;;
clean)
  clean-context
  ;;
add)
  create-context
  ;;
delete)
  delete-context
  ;;
"")
  echo "Error: No action specified. Use --merge, --add, or --delete"
  usage
  ;;
esac
