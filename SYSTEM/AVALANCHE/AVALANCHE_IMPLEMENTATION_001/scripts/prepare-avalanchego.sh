#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 OUTPUT_SOURCE_DIRECTORY" >&2
  exit 2
fi

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
output_source="$1"
go_binary="${GO_BINARY:-go}"
overlay="${root_dir}/compat/AVALANCHEGO_SECURITY_OVERLAY_001.json"

for command_name in git jq sha256sum; do
  if ! command -v "${command_name}" >/dev/null 2>&1; then
    echo "required command is unavailable: ${command_name}" >&2
    exit 1
  fi
done
if ! go_binary="$(command -v "${go_binary}")"; then
  echo "Go compiler is unavailable" >&2
  exit 1
fi
if [[ "$("${go_binary}" env GOVERSION)" != "go1.25.13" ]]; then
  echo "AvalancheGo preparation requires exact Go 1.25.13" >&2
  exit 1
fi
go_binary_dir="$(dirname "${go_binary}")"
export PATH="${go_binary_dir}:${PATH}"
if [[ -e "${output_source}" ]]; then
  echo "refusing to overwrite AvalancheGo source path: ${output_source}" >&2
  exit 1
fi

repository="$(jq -r '.base.repository' "${overlay}")"
tag="$(jq -r '.base.tag' "${overlay}")"
expected_tag_object="$(jq -r '.base.tag_object' "${overlay}")"
expected_commit="$(jq -r '.base.peeled_commit' "${overlay}")"
expected_module_sum="$(jq -r '.base.go_module_sum' "${overlay}")"
expected_go_mod_sum="$(jq -r '.base.go_mod_sum' "${overlay}")"
main_module_path="github.com/ava-labs/avalanchego"

remote_refs="$(git ls-remote --tags "${repository}" \
  "refs/tags/${tag}" "refs/tags/${tag}^{}" \
  "refs/tags/graft/coreth/${tag}" "refs/tags/graft/coreth/${tag}^{}" \
  "refs/tags/graft/subnet-evm/${tag}" "refs/tags/graft/subnet-evm/${tag}^{}" \
  "refs/tags/graft/evm/${tag}" "refs/tags/graft/evm/${tag}^{}")"
observed_tag_object="$(awk -v ref="refs/tags/${tag}" '$2 == ref {print $1}' <<<"${remote_refs}")"
observed_commit="$(awk -v ref="refs/tags/${tag}^{}" '$2 == ref {print $1}' <<<"${remote_refs}")"
if [[ "${observed_tag_object}" != "${expected_tag_object}" || "${observed_commit}" != "${expected_commit}" ]]; then
  echo "AvalancheGo remote tag identity mismatch" >&2
  exit 1
fi

module_json="$(GOTOOLCHAIN=local GOWORK=off "${go_binary}" mod download -json "${main_module_path}@${tag}")"
module_dir="$(jq -r '.Dir' <<<"${module_json}")"
module_sum="$(jq -r '.Sum' <<<"${module_json}")"
go_mod_sum="$(jq -r '.GoModSum' <<<"${module_json}")"
origin_vcs="$(jq -r '.Origin.VCS' <<<"${module_json}")"
origin_url="$(jq -r '.Origin.URL' <<<"${module_json}")"
origin_hash="$(jq -r '.Origin.Hash' <<<"${module_json}")"
origin_ref="$(jq -r '.Origin.Ref' <<<"${module_json}")"
if [[ ! -d "${module_dir}" ]]; then
  echo "authenticated AvalancheGo module source is unavailable" >&2
  exit 1
fi
if [[ "${module_sum}" != "${expected_module_sum}" || "${go_mod_sum}" != "${expected_go_mod_sum}" ]]; then
  echo "AvalancheGo authenticated module checksum mismatch" >&2
  exit 1
fi
if [[ "${origin_vcs}" != "git" || "${origin_url}" != "${repository%.git}" || "${origin_hash}" != "${expected_commit}" || "${origin_ref}" != "refs/tags/${tag}" ]]; then
  echo "AvalancheGo authenticated module origin mismatch" >&2
  exit 1
fi

mkdir -p "${output_source}"
cp -R "${module_dir}/." "${output_source}/"
chmod -R u+w "${output_source}"

for graft_name in coreth subnet-evm evm; do
  graft_module_path="${main_module_path}/graft/${graft_name}"
  graft_json="$(GOTOOLCHAIN=local GOWORK=off "${go_binary}" mod download -json "${graft_module_path}@${tag}")"
  graft_dir="$(jq -r '.Dir' <<<"${graft_json}")"
  graft_sum="$(jq -r '.Sum' <<<"${graft_json}")"
  graft_go_mod_sum="$(jq -r '.GoModSum' <<<"${graft_json}")"
  graft_origin_url="$(jq -r '.Origin.URL' <<<"${graft_json}")"
  graft_origin_subdir="$(jq -r '.Origin.Subdir' <<<"${graft_json}")"
  graft_origin_hash="$(jq -r '.Origin.Hash' <<<"${graft_json}")"
  graft_origin_ref="$(jq -r '.Origin.Ref' <<<"${graft_json}")"
  expected_graft_sum="$(jq -r ".base.graft_modules[\"${graft_name}\"].module_sum" "${overlay}")"
  expected_graft_go_mod_sum="$(jq -r ".base.graft_modules[\"${graft_name}\"].go_mod_sum" "${overlay}")"
  expected_graft_tag_object="$(jq -r ".base.graft_modules[\"${graft_name}\"].tag_object" "${overlay}")"
  graft_tag_ref="refs/tags/graft/${graft_name}/${tag}"
  observed_graft_tag_object="$(awk -v ref="${graft_tag_ref}" '$2 == ref {print $1}' <<<"${remote_refs}")"
  observed_graft_commit="$(awk -v ref="${graft_tag_ref}^{}" '$2 == ref {print $1}' <<<"${remote_refs}")"
  if [[ ! -d "${graft_dir}" || "${graft_sum}" != "${expected_graft_sum}" || "${graft_go_mod_sum}" != "${expected_graft_go_mod_sum}" ]]; then
    echo "AvalancheGo graft module checksum mismatch: ${graft_name}" >&2
    exit 1
  fi
  if [[ "${graft_origin_url}" != "${repository%.git}" || "${graft_origin_subdir}" != "graft/${graft_name}" || "${graft_origin_hash}" != "${expected_commit}" || "${graft_origin_ref}" != "${graft_tag_ref}" ]]; then
    echo "AvalancheGo graft module origin mismatch: ${graft_name}" >&2
    exit 1
  fi
  if [[ "${observed_graft_tag_object}" != "${expected_graft_tag_object}" || "${observed_graft_commit}" != "${expected_commit}" ]]; then
    echo "AvalancheGo graft remote tag identity mismatch: ${graft_name}" >&2
    exit 1
  fi
  mkdir -p "${output_source}/graft/${graft_name}"
  cp -R "${graft_dir}/." "${output_source}/graft/${graft_name}/"
  chmod -R u+w "${output_source}/graft/${graft_name}"
done

for filename in go.mod go.sum; do
  expected="$(jq -r ".base.${filename//./_}_sha256" "${overlay}")"
  observed="$(sha256sum "${output_source}/${filename}" | awk '{print $1}')"
  if [[ "${observed}" != "${expected}" ]]; then
    echo "AvalancheGo base ${filename} digest mismatch" >&2
    exit 1
  fi
done

source_tree_before="$(
  cd "${output_source}"
  find . -type f ! -path './go.mod' ! -path './go.sum' -print0 \
    | sort -z \
    | xargs -0 sha256sum \
    | sha256sum \
    | awk '{print $1}'
)"
(
  cd "${output_source}"
  GOTOOLCHAIN=local GOWORK=off "${go_binary}" get google.golang.org/grpc@v1.83.2
  GOTOOLCHAIN=local GOWORK=off "${go_binary}" mod tidy
)
source_tree_after="$(
  cd "${output_source}"
  find . -type f ! -path './go.mod' ! -path './go.sum' -print0 \
    | sort -z \
    | xargs -0 sha256sum \
    | sha256sum \
    | awk '{print $1}'
)"
if [[ "${source_tree_before}" != "${source_tree_after}" ]]; then
  echo "security overlay changed AvalancheGo files other than go.mod and go.sum" >&2
  exit 1
fi

for filename in go.mod go.sum; do
  expected="$(jq -r ".result.${filename//./_}_sha256" "${overlay}")"
  observed="$(sha256sum "${output_source}/${filename}" | awk '{print $1}')"
  if [[ "${observed}" != "${expected}" ]]; then
    echo "AvalancheGo overlaid ${filename} digest mismatch" >&2
    exit 1
  fi
done

while IFS=$'\t' read -r module expected_version; do
  observed_version="$(cd "${output_source}" && GOTOOLCHAIN=local GOWORK=off "${go_binary}" list -m -f '{{.Version}}' "${module}")"
  if [[ "${observed_version}" != "${expected_version}" ]]; then
    echo "AvalancheGo overlay module mismatch for ${module}: ${observed_version}" >&2
    exit 1
  fi
done < <(jq -r '.resolved_modules | to_entries[] | [.key, .value] | @tsv' "${overlay}")

jq -n \
  --arg repository "${repository}" \
  --arg tag "${tag}" \
  --arg tag_object "${observed_tag_object}" \
  --arg commit "${observed_commit}" \
  --arg module_sum "${module_sum}" \
  --arg go_mod_sum "${go_mod_sum}" \
  --argjson graft_modules "$(jq -c '.base.graft_modules' "${overlay}")" \
  '{
    schema: "AVALANCHEGO_SOURCE_IDENTITY_001",
    acquisition: "GO_AUTHENTICATED_MODULE_WITH_REMOTE_TAG_CONFIRMATION",
    repository: $repository,
    tag: $tag,
    tag_object: $tag_object,
    peeled_commit: $commit,
    module_sum: $module_sum,
    go_mod_sum: $go_mod_sum,
    graft_modules: $graft_modules
  }' >"${output_source}/AVALANCHEGO_SOURCE_IDENTITY.json"

echo "prepared AvalancheGo ${tag} (${observed_commit}) + PRESENCE_AVALANCHE_SECURITY_OVERLAY_001"
echo "source: ${output_source}"
