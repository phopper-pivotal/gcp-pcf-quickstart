#!/usr/bin/env bash

set -ex

my_dir="$( cd $(dirname $0) && pwd )"
pushd ${my_dir} > /dev/null
	source utils.sh
	set_resource_dirs
    check_param 'google_project'
    check_param 'google_json_key_data'
    check_param 'env_config'
    check_param 'PIVNET_API_TOKEN'
    check_param 'PIVNET_ACCEPT_EULA'
    export
    set_gcloud_config
    generate_env_config
popd > /dev/null

go install omg-cli
set -o allexport
eval $(omg-cli source-config --env-dir="${env_dir}")
set +o allexport

trap save_terraform_state EXIT
pushd "${release_dir}/src/omg-tf"
	ENV_DIR=${env_dir} ./init.sh
popd