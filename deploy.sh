#!/bin/sh

#       _ _   _           _           _            _
#  __ _(_) |_| |__  _   _| |__     __| | ___ _ __ | | ___  _   _
# / _` | | __| '_ \| | | | '_ \   / _` |/ _ \ '_ \| |/ _ \| | | |
#| (_| | | |_| | | | |_| | |_) | | (_| |  __/ |_) | | (_) | |_| |
# \__, |_|\__|_| |_|\__,_|_.__/   \__,_|\___| .__/|_|\___/ \__, |
# |___/                                     |_|            |___/

# =============================================================================
# deletes intermediate files
# =============================================================================
cleanup() {
  rm binadox-cloud-agent-*.zip 2> /dev/null
  rm bootstrap_agent 2> /dev/null
}

check_build() {
  exit_code=$1
  oname=$2

  if [ "${exit_code}" -ne "0" ]; then
    echo "Failed to build ${oname}"
    cleanup
    exit 1
  fi
}
# =============================================================================
# Build agent for specified platform
# build <tag> <os> <arch> <private key>
# =============================================================================
build() {
  tag=$1
  os=$2
  arch=$3
  priv=$4
  zip_oname="binadox-cloud-agent-${os}.zip"
  if [ "${os}" = "windows" ]; then
    exe_name="binadox-cloud-agent.exe"
  else
    exe_name="binadox-cloud-agent"
  fi
  linker_flags="-X main.sha1ver=${sha1ver} -X main.buildTime=${now} -X main.versionTag=${tag}"

  if [ ! -f "bootstrap_agent" ]; then
    echo "Bootstrap..."
    go build -ldflags "${linker_flags}" -o "bootstrap_agent"
    check_build "$?" bootstrap_agent
  fi

  echo "Building for ${os}/${arch}"
  now=$(date +'%Y-%m-%d_%T')
  sha1ver=$(git rev-parse HEAD)

  (export GOOS=${os}; export GOARCH=${arch}; go build -ldflags "${linker_flags}" .)
  check_build "$?" "${exe_name}"

  ./bootstrap_agent --zip --out "${zip_oname}" --in ${exe_name} --priv "${priv}"
  check_build "$?" "${zip_oname}"
  rm ${exe_name}
}

# ==================================================================================
# create release
#
# create_release <tag> <api key>
# ==================================================================================
generate_post_data()
{
  version=$1
  branch=$(git rev-parse --abbrev-ref HEAD)

  cat <<EOF
{
  "tag_name": "$version",
  "target_commitish": "$branch",
  "name": "release: $version",
  "body": "release: $version",
  "draft": false,
  "prerelease": false
}
EOF
}

create_release() {
  tag=$1
  token=$2
  echo "Building release ${tag}, token = ${token}"
  branch=$(git rev-parse --abbrev-ref HEAD)
  auth="Authorization: token ${token}"
  req="$(generate_post_data "${tag}")"

  resp=$(curl -s -X POST -H "${auth}" "https://api.github.com/repos/binadox-public/binadox-cloud-agent/releases" -d "${req}")
  release_id=$(echo "${resp}" | jq ".id")
  if [ "${release_id}" = "null" ] ; then
    echo "Failed to create release"
    echo "--------------------------------------------"
    echo "Request:"
    echo "${req}"
    echo "Server response:"
    echo "${resp}"
    echo "--------------------------------------------"
    cleanup
    exit 1
  fi

  echo "Release ID ${release_id}"
  for filename in binadox-cloud-agent-*.zip ; do
    [ -e "$filename" ] || break
    echo "uploading ${filename}"
    asset="https://uploads.github.com/repos/binadox-public/binadox-cloud-agent/releases/${release_id}/assets?name=${filename}"
    resp=$(curl --data-binary @"${filename}" -H "${auth}" -H "Content-Type: application/octet-stream" "${asset}")

    rm "${filename}"
    download_url=$(echo "${resp}" | jq ".browser_download_url")
    if [ -z "${download_url}" ]; then
      echo "Failed to upload ${filename}"
      echo "--------------------------------------------"
      echo "Server response:"
      echo "${resp}"
      echo "--------------------------------------------"
      cleanup
      exit 1
    else
      echo "Uploaded to ${download_url}"
    fi
  done
}
# ===========================================================================
# usage
# ===========================================================================
usage() {
  echo "deploy.sh tag=<versionTag> token=<apiKey> priv=<private key>"
}

# ===========================================================================
# main
# ===========================================================================
CONFIG=$@

for line in $CONFIG; do
  eval "$line"
done

if [ -z "${tag}" ]; then
    usage
    exit 1
fi

if [ -z "${token}" ]; then
    token=${GITHUB_API_TOKEN}
    if [ -z "${token}" ]; then
      usage
      exit 1
    fi
fi

if [ -z "${priv}" ]; then
  usage
  exit 1
fi

build "${tag}" "linux" "amd64" "${priv}"
build "${tag}" "windows" "amd64" "${priv}"
create_release "${tag}" "${token}"

cleanup

# https://github.com/binadox-public/binadox-cloud-agent/releases.atom
