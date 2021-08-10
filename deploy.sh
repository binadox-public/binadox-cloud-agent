#!/bin/sh

#  ______ _____ _______  ______ _     _ ______    _______  _____  _______ ______  _______  _____          _____  __   __
# |  ____   |      |    |  ____ |     | |_____]   |       |     | |  |  | |     \ |______ |_____] |      |     |   \_/
# |_____| __|__    |    |_____| |_____| |_____] . |_____  |_____| |  |  | |_____/ |______ |       |_____ |_____|    |

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
  fi

  echo "Building for ${os}/${arch}"
  now=$(date +'%Y-%m-%d_%T')
  sha1ver=$(git rev-parse HEAD)

  (export GOOS=${os}; export GOARCH=${arch}; go build -ldflags "${linker_flags}")
  ./bootstrap_agent --zip --out "${zip_oname}" --in ${exe_name} --priv "${priv}"
  rm ${exe_name}
}

# ==================================================================================
# create release
#
# create_release <tag> <api key>
# ==================================================================================
create_release() {
  tag=$1
  token=$2
  comment="release ${tag}"
  git tag -a "${tag}" -m "${comment}"
  git push origin "${tag}"
  auth="Authorization: token ${token}"
  req="{\"tag_name\":\"${tag}\", \"name\": \"${tag}\",\"draft\": false,\"prerelease\": false}"
  resp=$(curl -s -X POST -H "${auth}" "https://api.github.com/repos/binadox-public/binadox-cloud-agent/releases" -d "${req}")
  release_id=$(echo "${resp}" | jq ".id")
  echo "Release ID ${release_id}"
  for filename in $(ls -1 binadox-cloud-agent-*.zip) ; do
    [ -e "$filename" ] || break
    echo "uploading ${filename}"
    asset="https://uploads.github.com/repos/binadox-public/binadox-cloud-agent/releases/${release_id}/assets?name=${filename}"
    resp=$(curl --data-binary @"${filename}" -H "${auth}" -H "Content-Type: application/octet-stream" "${asset}")
    rm "${filename}"
    download_url=$(echo "${resp}" | jq ".browser_download_url")
    if [ -z "${download_url}" ]; then
      echo "Failed to upload ${filename}"
      echo "${resp}"
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
  usage
  exit 1
fi

if [ -z "${priv}" ]; then
  usage
  exit 1
fi

build "${tag}" "linux" "amd64" "${priv}"
build "${tag}" "windows" "amd64" "${priv}"
create_release "${tag}" "${token}"

if [ -f "bootstrap_agent" ]; then
    rm "bootstrap_agent"
fi

# https://github.com/binadox-public/binadox-cloud-agent/releases.atom
