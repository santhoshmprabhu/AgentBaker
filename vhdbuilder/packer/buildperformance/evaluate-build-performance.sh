#!/bin/bash

log_and_exit () {
  local FILE=${1}
  local ERR=${2}
  local SHOW_FILE=${3:-false}
  echo "##vso[task.logissue type=warning;sourcepath=$(basename $0);]${FILE} ${ERR}. Skipping build performance evaluation."
  echo "##vso[task.complete result=SucceededWithIssues;]"
  if [[ ${SHOW_FILE} == true ]]; then
    cat ${FILE}
  fi
  exit 0
}

if [[ ! -f ${PERFORMANCE_DATA_FILE} ]]; then
  log_and_exit ${PERFORMANCE_DATA_FILE} "not found"
fi

SCRIPT_COUNT=$(jq -e 'keys | length' ${PERFORMANCE_DATA_FILE})
if [[ $? -ne 0 ]]; then
  log_and_exit ${PERFORMANCE_DATA_FILE} "contains invalid json" true
fi

if [[ ${SCRIPT_COUNT} -eq 0 ]]; then
  log_and_exit ${PERFORMANCE_DATA_FILE} "contains no data"
fi

echo -e "\nGenerating build performance data for ${SIG_IMAGE_NAME}...\n"

jq --arg sig_name "${SIG_IMAGE_NAME}" \
  --arg arch "${ARCHITECTURE}" \
  --arg captured_sig_version "${CAPTURED_SIG_VERSION}" \
  --arg build_id "${BUILD_ID}" \
  --arg date "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
  --arg status "${JOB_STATUS}" \
  --arg branch "${GIT_BRANCH}" \
  --arg commit "${GIT_VERSION}" \
  'to_entries | ([
  {key: "sig_image_name", value: $sig_name},
  {key: "architecture", value: $arch},
  {key: "captured_sig_version", value: $captured_sig_version},
  {key: "build_id", value: $build_id},
  {key: "build_datetime", value: $date},
  {key: "outcome", value: $status},
  {key: "branch", value: $branch},
  {key: "commit", value: $commit}
] + .) | from_entries' ${PERFORMANCE_DATA_FILE} > ${SIG_IMAGE_NAME}-build-performance.json

rm ${PERFORMANCE_DATA_FILE}

echo "##[group]Build Performance"
jq . -C ${SIG_IMAGE_NAME}-build-performance.json
echo "##[endgroup]"

BLOB_URL_REGEX="^https:\/\/.+\.blob\.core\.windows\.net\/vhd(s)?$"
if [[ $CLASSIC_BLOB =~ $BLOB_URL_REGEX ]]; then
    STORAGE_ACCOUNT_NAME=$(echo $CLASSIC_BLOB | sed -E 's|https://(.*)\.blob\.core\.windows\.net(:443)?/(.*)?|\1|')
fi

echo -e "\nENVIRONMENT is: ${ENVIRONMENT}"
if [ "${ENVIRONMENT,,}" != "prod" ]; then
  mv ${SIG_IMAGE_NAME}-build-performance.json vhdbuilder/packer/buildperformance
  pushd vhdbuilder/packer/buildperformance || exit 0
    echo -e "\nRunning build performance evaluation program...\n"
    az storage blob download --account-name ${STORAGE_ACCOUNT_NAME} \
      --container-name ${BUILD_PERFORMANCE_CONTAINER_NAME} \
      --file ${BUILD_PERFORMANCE_BINARY} \
      --name ${BUILD_PERFORMANCE_BINARY} \
      --auth-mode login

    chmod +x ${BUILD_PERFORMANCE_BINARY}
    ./${BUILD_PERFORMANCE_BINARY}
    rm ${SIG_IMAGE_NAME}-build-performance.json
  popd || exit 0
else
  rm ${SIG_IMAGE_NAME}-build-performance.json
  echo -e "Skipping build performance evaluation for prod"
  exit 0
fi

echo -e "\nBuild performance evaluation script completed."
